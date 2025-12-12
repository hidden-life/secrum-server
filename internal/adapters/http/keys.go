package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/app/keys"
)

func RegisterKeyRoutes(r chi.Router, svc *keys.Service) {
	r.Route("/keys", func(r chi.Router) {
		r.Post("/device", uploadDeviceKeys(svc))
		r.Get("/bundle/{user_id}", getBundle(svc))
	})
}

func getBundle(svc *keys.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		if userID == "" {
			asError(w, http.StatusBadRequest, "missing user_id in request")
			return
		}

		uid, err := uuid.Parse(userID)
		if err != nil {
			asError(w, http.StatusBadRequest, "invalid user_id")
			return
		}

		bundle, err := svc.GetUserBundle(r.Context(), uid)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, bundle)
	}
}

func uploadDeviceKeys(svc *keys.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devID := DeviceIDFromContext(r.Context())
		if devID == "" {
			asError(w, http.StatusUnauthorized, "missing device_id")
			return
		}

		deviceID, err := uuid.Parse(devID)
		if err != nil {
			asError(w, http.StatusBadRequest, "invalid device_id")
			return
		}

		var req keys.UploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if err := svc.UploadDeviceKeys(r.Context(), deviceID, req); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusNoContent, nil)
	}
}
