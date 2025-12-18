package http

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/attachments"
	"github.com/hidden-life/secrum-server/internal/app/context"
)

func RegisterAttachmentsRoutes(r chi.Router, svc *attachments.Service) {
	r.Route("/attachments", func(r chi.Router) {
		r.Post("/", uploadAttachmentHandler(svc))
		r.Get("/{id}", downloadAttachmentHandler(svc))
	})
}

func uploadAttachmentHandler(svc *attachments.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := context.UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var (
			fileSize *int64
			mimeType *string
		)

		if ct := r.Header.Get("Content-Type"); ct != "" {
			mimeType = &ct
		}

		if cl := r.Header.Get("Content-Length"); cl != "" {
			if v, err := strconv.ParseInt(cl, 10, 64); err == nil {
				fileSize = &v
			}
		}

		res, err := svc.Upload(r.Context(), userID, r.Body, fileSize, mimeType)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusCreated, res)
	}
}

func downloadAttachmentHandler(svc *attachments.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := context.UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		id := chi.URLParam(r, "id")
		if id == "" {
			asError(w, http.StatusBadRequest, "missing attachment id")
			return
		}

		info, err := svc.Download(r.Context(), userID, id)
		if err != nil {
			asError(w, http.StatusNotFound, err.Error())
			return
		}
		defer info.Reader.Close()

		if info.MimeType != nil && *info.MimeType != "" {
			w.Header().Set("Content-Type", *info.MimeType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}

		if info.Size != nil && *info.Size > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(*info.Size, 10))
		}

		if _, err := io.Copy(w, info.Reader); err != nil {
			return
		}
	}
}
