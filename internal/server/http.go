package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// HTTPServer represents the HTTP server.
type HTTPServer struct {
	router *chi.Mux
	logger *zap.Logger
	port   string
}

func NewHTTPServer(logger *zap.Logger, port string) *HTTPServer {
	r := chi.NewRouter()

	srv := &HTTPServer{
		router: r,
		logger: logger,
		port:   port,
	}

	// Register routers
	r.Get("/healthz", srv.healthHandler)

	return srv
}

func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("port", s.port))

	return http.ListenAndServe(":"+s.port, s.router)
}

func (s *HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
