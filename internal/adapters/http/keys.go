package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/keys"
	"github.com/hidden-life/secrum-server/internal/ports"
)

func RegisterKeyRoutes(
	r chi.Router,
	svc *keys.Service,
	manager ports.TokenManager,
	store ports.SessionStore,
	deviceRepo ports.DeviceRepository,
) {
	r.Route("/keys", func(r chi.Router) {
		r.Use(AuthMiddleware(manager, store, deviceRepo))
		r.Post("/upload", uploadHandler(svc))
		r.Post("/fetch", fetchHandler(svc))

		r.Get("/pre-key-bundle", preKeyBundleHandler(svc))
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

func preKeyBundleHandler(svc *keys.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		deviceID := r.URL.Query().Get("device_id")
		if userID == "" || deviceID == "" {
			asError(w, http.StatusBadRequest, "invalid request body: user_id or device_id are required")
			return
		}

		resp, err := svc.PreKeyBundle(r.Context(), &keys.PreKeyBundleRequest{
			UserID:   userID,
			DeviceID: deviceID,
		})

		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}
