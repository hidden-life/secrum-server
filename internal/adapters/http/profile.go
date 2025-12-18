package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/context"
	"github.com/hidden-life/secrum-server/internal/app/profile"
)

func RegisterProfileRoutes(r chi.Router, svc *profile.Service) {
	r.Get("/me", getMeHandler(svc))
	r.Put("/me/profile", updateProfileHandler(svc))
	r.Get("/me/{id}/safety", getUserSafetyHandler(svc))
}

func getUserSafetyHandler(svc *profile.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			asError(w, http.StatusBadRequest, "missing user id")
			return
		}

		resp, err := svc.GetMe(r.Context(), id)
		if err != nil {
			asError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp.PhoneHash = ""

		asJson(w, http.StatusOK, resp)
	}
}

func updateProfileHandler(svc *profile.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := context.UserIDFromContext(r.Context())
		if uid == "" {
			asError(w, http.StatusBadRequest, "unauthorized")
			return
		}

		var req profile.UpdateProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		resp, err := svc.UpdateProfile(r.Context(), uid, &req)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}

func getMeHandler(svc *profile.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid := context.UserIDFromContext(r.Context())
		if uid == "" {
			asError(w, http.StatusBadRequest, "unauthorized")
			return
		}

		resp, err := svc.GetMe(r.Context(), uid)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}
