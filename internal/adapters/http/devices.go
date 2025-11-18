package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/devices"
)

func RegisterDevicesRoutes(r chi.Router, svc *devices.Service) {
	r.Route("/devices", func(r chi.Router) {
		r.Get("/", listDevicesHandler(svc))
		r.Post("/{device_id}/deactivate", deactivateDeviceHandler(svc))
		r.Delete("/{device_id}", deleteDeviceHandler(svc))
	})
}

func deleteDeviceHandler(svc *devices.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "missing user id in context")
			return
		}
		currentDeviceID := DeviceIDFromContext(r.Context())
		deviceID := chi.URLParam(r, "device_id")
		if deviceID == "" {
			asError(w, http.StatusBadRequest, "missing device id")
			return
		}

		if err := svc.DeleteDevice(r.Context(), userID, deviceID, currentDeviceID); err != nil {
			asError(w, http.StatusBadRequest, "failed to delete device: "+err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func deactivateDeviceHandler(svc *devices.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "missing user id in context")
			return
		}
		currentDeviceID := DeviceIDFromContext(r.Context())
		deviceID := chi.URLParam(r, "device_id")
		if deviceID == "" {
			asError(w, http.StatusBadRequest, "missing deviceID")
			return
		}

		if err := svc.DeactivateDevice(r.Context(), userID, deviceID, currentDeviceID); err != nil {
			asError(w, http.StatusBadRequest, "failed to deactivate device: "+err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func listDevicesHandler(svc *devices.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "missing user id in context")
			return
		}

		currentDeviceID, _ := r.Context().Value("device_id").(string)
		list, err := svc.ListUserDevices(r.Context(), userID, currentDeviceID)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, list)
	}
}
