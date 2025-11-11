package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/auth"
)

func RegisterAuthRoutes(r chi.Router, svc *auth.Service) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/begin", beginHandler(svc))
		r.Post("/verify", verifyHandler(svc))
	})
}

func beginHandler(svc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req auth.BeginRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid request payload")
			return
		}

		resp, err := svc.BeginRegistration(r.Context(), req)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}

func verifyHandler(svc *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req auth.VerifyRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid request payload")
			return
		}

		resp, err := svc.VerifyRegistration(r.Context(), req)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}

func asJson(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func asError(w http.ResponseWriter, status int, msg string) {
	asJson(w, status, map[string]string{"error": msg})
}
