package api

import (
	"fmt"
	"raspberry_sensors/internal/logger"
	"net/http"
	"sync"
)

type Server struct {
	controlChannels [][2]chan bool
	QuitChan chan struct{}
	mu   sync.Mutex // To handle concurrent requests safely
}

func NewServer(controlChannels [][2]chan bool) *Server {
	return &Server{
		controlChannels: controlChannels,
		QuitChan: make(chan struct{}),
	}
}

func (s *Server) stopMonitoring(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, controlChannels := range s.controlChannels {
		controlChannels[0] <- false
		<- controlChannels[1]  // wait for sensor confirmation
	}
}

func (s *Server) kill(w http.ResponseWriter, r *http.Request) {
	logger.GlobalLogger.Info("Program killed. Shutting down...")
	s.stopMonitoring(w, r)
	logger.GlobalLogger.Info("Bye!")
	s.QuitChan <- struct{}{}
}

func (s *Server) startMonitoring(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, controlChannels := range s.controlChannels {
		controlChannels[0] <- true
		<- controlChannels[1]  // wait for sensor confirmation
	}
}

func (s *Server) Start(port int) {
	http.HandleFunc("/sensors/stop", s.stopMonitoring)
	http.HandleFunc("/sensors/start", s.startMonitoring)
	http.HandleFunc("/sensors/kill", s.kill)
	addr := fmt.Sprintf(":%d", port)
	logger.GlobalLogger.Infof("Listening on port %d...", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.GlobalLogger.Fatalf("Failed to start server: %v", err)
	}
}