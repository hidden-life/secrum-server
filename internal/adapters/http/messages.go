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
		r.Delete("/{id}/delete/me", deleteForMe(svc))
		r.Delete("/{id}/delete/all", deleteForAll(svc))
		r.Patch("/{id}", edit(svc))
		r.Post("/{id}/reaction/add", addReaction(svc))
		r.Delete("/{id}/reaction/remove", removeReaction(svc))
		// search
		r.Get("/search", searchMessages(svc))
		// pin/unpin
		r.Post("/{id}/pin", pinMessage(svc))
		r.Post("/{id}/unpin", unpinMessage(svc))
		// forwarding
		r.Post("/{id}/forward/user/{user_id}", forwardToUser(svc))
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

func deleteForMe(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		messageID := chi.URLParam(r, "id")
		if err := svc.DeleteForMe(r.Context(), userID, messageID); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusNoContent, nil)
	}
}

func deleteForAll(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		messageID := chi.URLParam(r, "id")
		if err := svc.DeleteForAll(r.Context(), userID, messageID); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusNoContent, nil)
	}
}

func edit(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var req messages.EditMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "failed to decode request body")
			return
		}

		messageID := chi.URLParam(r, "id")
		if err := svc.EditMessage(r.Context(), userID, messageID, req); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func addReaction(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		messageID := chi.URLParam(r, "id")
		var body struct {
			Emoji string `json:"emoji"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			asError(w, http.StatusBadRequest, "failed to decode request body")
			return
		}

		if err := svc.AddReaction(r.Context(), userID, messageID, body.Emoji); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func removeReaction(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		messageID := chi.URLParam(r, "id")
		if err := svc.RemoveReaction(r.Context(), userID, messageID); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func searchMessages(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		q := r.URL.Query().Get("q")
		limitStr := r.URL.Query().Get("limit")
		beforeStr := r.URL.Query().Get("before")

		limit := 50
		if limitStr != "" {
			if v, err := strconv.Atoi(limitStr); err == nil {
				limit = v
			}
		}

		var before *time.Time
		if beforeStr != "" {
			if v, err := time.Parse(time.RFC3339, beforeStr); err == nil {
				before = &v
			}
		}

		msgs, err := svc.SearchMessages(r.Context(), userID, q, limit, before)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, msgs)
	}
}

func pinMessage(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		messageID := chi.URLParam(r, "id")
		if err := svc.PinMessage(r.Context(), userID, messageID); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, nil)
	}
}

func unpinMessage(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		messageID := chi.URLParam(r, "id")
		if err := svc.UnpinMessage(r.Context(), userID, messageID); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, nil)
	}
}

func forwardToUser(svc *messages.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actorID := UserIDFromContext(r.Context())
		if actorID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		deviceID := DeviceIDFromContext(r.Context())
		if deviceID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		srcMessageID := chi.URLParam(r, "id")
		targetUserID := chi.URLParam(r, "user_id")

		resp, err := svc.ForwardToUser(r.Context(), actorID, deviceID, srcMessageID, targetUserID)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, resp)
	}
}
