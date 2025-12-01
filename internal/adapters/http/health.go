package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterHealthRoutes(r chi.Router) {
	r.Route("/health", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			asJson(w, http.StatusOK, map[string]string{
				"status": "ok",
			})
		})
	})
}
