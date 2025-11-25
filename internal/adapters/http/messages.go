package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/messages"
)

func RegisterMessagesRoutes(r chi.Router, svc *messages.Service) {
	r.Route("/messages", func(r chi.Router) {
		r.Post("/send", sendHandler(svc))
		r.Get("/pending", pendingHandler(svc))
		r.Post("/ack", ackHandler(svc))
		r.Get("/history/{peer_id}", getChatHistoryHandler(svc))
	})
}

func getChatHistoryHandler(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		peerID := chi.URLParam(r, "peer_id")
		limitStr := r.URL.Query().Get("limit")
		beforeStr := r.URL.Query().Get("before")

		limit := 50
		if limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
				limit = v
			}
		}

		var before *time.Time
		if beforeStr != "" {
			if v, err := time.Parse(time.RFC3339Nano, beforeStr); err == nil {
				before = &v
			}
		}

		msgs, err := svc.GetChatHistory(r.Context(), userID, peerID, limit, before)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, msgs)
	}
}

func ackHandler(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devID := DeviceIDFromContext(r.Context())
		if devID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req messages.AckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := svc.AckDelivered(r.Context(), devID, req); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func pendingHandler(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		devID := DeviceIDFromContext(r.Context())
		if devID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 0
		if limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil {
				limit = v
			}
		}

		msgs, err := svc.FetchPending(r.Context(), devID, limit)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, msgs)
	}
}

func sendHandler(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		senderUID := UserIDFromContext(r.Context())
		senderDevID := DeviceIDFromContext(r.Context())
		if senderDevID == "" || senderUID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req messages.SendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "failed to decode request body")
			return
		}

		resp, err := svc.Send(r.Context(), senderUID, senderDevID, &req)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}
