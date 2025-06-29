package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thomasmarlow/the-trainman/internal/config"
)

type Server struct {
	router        chi.Router
	configManager *config.Manager
}

type PingResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewServer(configManager *config.Manager) *Server {
	s := &Server{
		router:        chi.NewRouter(),
		configManager: configManager,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))
}

func (s *Server) setupRoutes() {
	s.router.Get("/ping", s.handlePing)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	message := "pong"
	if s.configManager != nil {
		message = s.configManager.GetMessage()
	}

	response := PingResponse{
		Status:  "ok",
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
