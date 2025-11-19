package http

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/attachments"
)

func RegisterAttachmentsRoutes(r chi.Router, svc *attachments.Service) {
	r.Route("/attachments", func(r chi.Router) {
		r.Post("/", uploadAttachmentHandler(svc))
		r.Get("/{id}", downloadAttachmentHandler(svc))
	})
}

func uploadAttachmentHandler(svc *attachments.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID == "" {
			asError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var (
			fileSize *int64
			mimeType *string
		)

		if sizeStr := r.Header.Get("X-File-Size"); sizeStr != "" {
			if v, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
				fileSize = &v
			}
		}

		if mt := r.Header.Get("X-Mime-Type"); mt != "" {
			mimeType = &mt
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
		userID := UserIDFromContext(r.Context())
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

		if info.MimeType != nil {
			w.Header().Set("Content-Type", *info.MimeType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		if info.Size != nil {
			w.Header().Set("Content-Length", strconv.FormatInt(*info.Size, 10))
		}

		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, info.Reader)
	}
}
