package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/keys"
)

func RegisterKeyRoutes(r chi.Router, svc *keys.Service) {
	r.Route("/keys", func(r chi.Router) {
		r.Post("/upload", uploadHandler(svc))
		r.Post("/fetch", fetchHandler(svc))
	})
}

func fetchHandler(svc *keys.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req keys.FetchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		resp, err := svc.Fetch(r.Context(), req)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}

func uploadHandler(svc *keys.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req keys.UploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		resp, err := svc.Upload(r.Context(), req)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}
