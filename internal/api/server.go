package api

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	controlChannels []chan bool
	QuitChan chan struct{}
	mu   sync.Mutex // To handle concurrent requests safely
}

func NewServer(controlChannels []chan bool) *Server {
	return &Server{
		controlChannels: controlChannels,
		QuitChan: make(chan struct{}),
	}
}

func (s *Server) stopMonitoring(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, controlChannel := range s.controlChannels {
		controlChannel <- false
		<- controlChannel  // wait for sensor confirmation
	}
}

func (s *Server) kill(w http.ResponseWriter, r *http.Request) {
	log.Println("Shutting down...")
	s.stopMonitoring(w, r)
	log.Println("Bye!")
	s.QuitChan <- struct{}{}
}

func (s *Server) startMonitoring(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, controlChannel := range s.controlChannels {
		controlChannel <- true
		<- controlChannel  // wait for sensor confirmation
	}
}

func (s *Server) Start(port int) {
	http.HandleFunc("/sensors/stop", s.stopMonitoring)
	http.HandleFunc("/sensors/start", s.startMonitoring)
	http.HandleFunc("/sensors/kill", s.kill)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Listening on port %d...", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}