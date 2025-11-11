package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/messages"
)

func RegisterMessagesRoutes(r chi.Router, svc *messages.Service) {
	r.Route("/messages", func(r chi.Router) {
		r.Post("/send", sendHandler(svc))
		r.Get("/pending", pendingHandler(svc))
		r.Post("/ack", ackHandler(svc))
	})
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
