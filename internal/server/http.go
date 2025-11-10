package server

import (
	"context"
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
	srv    *http.Server
}

func NewHTTPServer(logger *zap.Logger, port string) *HTTPServer {
	r := chi.NewRouter()

	s := &HTTPServer{
		router: r,
		logger: logger,
		port:   port,
	}

	// Register routes
	r.Get("/healthz", s.healthHandler)

	s.srv = &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	return s
}

func (s *HTTPServer) Router() *chi.Mux {
	return s.router
}

func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", zap.String("port", s.port))

	return http.ListenAndServe(":"+s.port, s.router)
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	return s.srv.Shutdown(ctx)
}

func (s *HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}
