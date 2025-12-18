package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/context"
	"github.com/hidden-life/secrum-server/internal/app/sync"
)

func RegisterSyncEventRoutes(r chi.Router, svc *sync.Service) {
	r.Route("/sync", func(r chi.Router) {
		r.Get("/full", fullSync(svc))
		r.Get("/delta", deltaSync(svc))
	})
}

func fullSync(svc *sync.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := context.UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		resp, err := svc.Full(r.Context(), userID)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}

func deltaSync(svc *sync.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := context.UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		cursorStr := r.URL.Query().Get("cursor")
		limitStr := r.URL.Query().Get("limit")
		var cursor int64 = 0
		if cursorStr != "" {
			if v, err := strconv.ParseInt(cursorStr, 10, 64); err == nil && v >= 0 {
				cursor = v
			}
		}

		limit := 0
		if limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil {
				limit = v
			}
		}

		resp, err := svc.Delta(r.Context(), userID, cursor, limit)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}
