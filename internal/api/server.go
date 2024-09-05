package api

import (
	"fmt"
	"net/http"
	"raspberry_sensors/internal/config"
	"raspberry_sensors/internal/logger"
	"sync"
	"time"

	"github.com/rs/zerolog/hlog"
)

type Server struct {
	controlChannels [][2]chan bool
	QuitChan        chan struct{}
	mu              sync.Mutex // To handle concurrent requests safely
}

func NewServer(controlChannels [][2]chan bool) *Server {
	return &Server{
		controlChannels: controlChannels,
		QuitChan:        make(chan struct{}),
	}
}

func (s *Server) stopMonitoring(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, controlChannels := range s.controlChannels {
		controlChannels[0] <- false
		<-controlChannels[1] // wait for sensor confirmation
	}
}

func (s *Server) kill(w http.ResponseWriter, r *http.Request) {
	l := logger.Get()
	l.Info().Msg("Program killed. Shutting down...")
	s.stopMonitoring(w, r)
	l.Info().Msg("Bye!")
	s.QuitChan <- struct{}{}
}

func (s *Server) startMonitoring(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, controlChannels := range s.controlChannels {
		controlChannels[0] <- true
		<-controlChannels[1] // wait for sensor confirmation
	}
}

func (s *Server) Start() {
	l := logger.Get()
	c := config.Get()
	mux := http.NewServeMux()
	mux.HandleFunc("/sensors/stop", s.stopMonitoring)
	mux.HandleFunc("/sensors/start", s.startMonitoring)
	mux.HandleFunc("/sensors/kill", s.kill)
	addr := fmt.Sprintf("%s:%s", c.Api.Host, c.Api.Port)
	l.Info().Msgf("Listening/Serving on port %s:%s...", c.Api.Host, c.Api.Port)
	if err := http.ListenAndServe(addr, requestLogger(mux)); err != nil {
		l.Fatal().Err(err).Msg("Failed to start server")
	}
}

func requestLogger(next http.Handler) http.Handler {
	l := logger.Get()

	h := hlog.NewHandler(l)

	accessHandler := hlog.AccessHandler(
			func(r *http.Request, status, size int, duration time.Duration) {
					hlog.FromRequest(r).Info().
							Str("method", r.Method).
							Stringer("url", r.URL).
							Int("status_code", status).
							Int("response_size_bytes", size).
							Dur("elapsed_ms", duration).
							Msg("incoming request")
			},
	)

	userAgentHandler := hlog.UserAgentHandler("http_user_agent")

	return h(accessHandler(userAgentHandler(next)))
}