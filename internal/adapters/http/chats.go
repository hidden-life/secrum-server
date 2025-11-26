package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/chats"
)

func RegisterChatRoutes(r chi.Router, svc *chats.Service) {
	r.Route("/chats", func(r chi.Router) {
		r.Get("/", chatsListHandler(svc))

		r.Post("/{peer_id}/pin", pinChat(svc))
		r.Delete("/{peer_id}/pin", unPinChat(svc))

		r.Post("/{peer_id}/archive", archiveChat(svc))
		r.Delete("/{peer_id}/archive", unArchiveChat(svc))

		r.Post("/{peer_id}/mute", muteChat(svc))
		r.Delete("/{peer_id}/mute", unMuteChat(svc))
	})
}

func chatsListHandler(svc *chats.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		items, err := svc.List(r.Context(), userID)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, items)
	}
}

func pinChat(svc *chats.Service) http.HandlerFunc {
	return chatStateWrapper(svc.SetPinned, true)
}

func unPinChat(svc *chats.Service) http.HandlerFunc {
	return chatStateWrapper(svc.SetPinned, false)
}

func archiveChat(svc *chats.Service) http.HandlerFunc {
	return chatStateWrapper(svc.SetArchived, true)
}

func unArchiveChat(svc *chats.Service) http.HandlerFunc {
	return chatStateWrapper(svc.SetArchived, false)
}

func muteChat(svc *chats.Service) http.HandlerFunc {
	return chatStateWrapper(svc.SetMuted, true)
}

func unMuteChat(svc *chats.Service) http.HandlerFunc {
	return chatStateWrapper(svc.SetMuted, false)
}

func chatStateWrapper(fn func(context.Context, string, string, bool) error, val bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		peerID := chi.URLParam(r, "peer_id")
		if peerID == "" || peerID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if err := fn(r.Context(), peerID, userID, val); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, nil)
	}
}
