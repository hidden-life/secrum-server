package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/app/users"
)

func RegisterUserSearchRoutes(r chi.Router, svc *users.SearchService) {
	r.Get("/users/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		res, err := svc.Search(r.Context(), q)
		if err != nil {
			asError(w, http.StatusBadRequest, err.Error())
			return
		}

		asJson(w, http.StatusOK, res) // nil - ok
	})
}
