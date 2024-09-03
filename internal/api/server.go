package api

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	ctrlChan chan bool
	mu   sync.Mutex // To handle concurrent requests safely
}

func NewServer(ctrlChan chan bool) *Server {
	return &Server{
		ctrlChan: ctrlChan,
	}
}

func (s *Server) stopMonitoring(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctrlChan <- false
}

func (s *Server) startMonitoring(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctrlChan <- true
}

func (s *Server) Start(port int) {
	http.HandleFunc("/sensors/stop", s.stopMonitoring)
	http.HandleFunc("/sensors/start", s.startMonitoring)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Listening on port %d...", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}