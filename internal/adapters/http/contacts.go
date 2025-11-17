package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/app/contact"
)

func RegisterContactRoutes(r chi.Router, svc *contact.Service) {
	r.Post("/contact/add", addContactHandler(svc))
	r.Post("/contact/sync", syncContactsHandler(svc))
	r.Delete("/contacts/{id}", removeContactHandler(svc))
	r.Get("/contacts", contactsListHandler(svc))
}

func contactsListHandler(svc *contact.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerStr := UserIDFromContext(r.Context())
		if ownerStr == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		owner, _ := uuid.Parse(ownerStr)

		profiles, err := svc.List(r.Context(), owner)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, profiles)
	}
}

func removeContactHandler(svc *contact.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerStr := UserIDFromContext(r.Context())
		if ownerStr == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		owner, _ := uuid.Parse(ownerStr)
		targetStr := chi.URLParam(r, "contact_user_id")
		target, err := uuid.Parse(targetStr)
		if err != nil {
			asError(w, http.StatusBadRequest, "invalid contact user id")
			return
		}

		if err := svc.Remove(r.Context(), owner, target); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func syncContactsHandler(svc *contact.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerStr := UserIDFromContext(r.Context())
		if ownerStr == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		owner, _ := uuid.Parse(ownerStr)
		var req contact.SyncContactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid input request")
			return
		}

		if err := svc.Sync(r.Context(), owner, req.PhoneHashes); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func addContactHandler(svc *contact.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ownerStr := UserIDFromContext(r.Context())
		if ownerStr == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		owner, _ := uuid.Parse(ownerStr)
		var req contact.AddContactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			asError(w, http.StatusBadRequest, "invalid input request")
			return
		}

		if err := svc.AddContact(r.Context(), owner, req.PhoneHash); err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
