package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/chats"
)

func RegisterChatRoutes(r chi.Router, svc *chats.Service) {
	r.Route("/chats", func(r chi.Router) {
		r.Get("/", chatsListHandler(svc))
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
