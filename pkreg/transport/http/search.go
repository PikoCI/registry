package http

import (
	"net/http"
	"strconv"

	"github.com/pikoci/registry/pkreg"
)

func searchTypes(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		query := q.Get("q")
		kind := q.Get("kind")
		tagFilter := q.Get("tag")
		ns := q.Get("namespace")

		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}
		perPage, _ := strconv.Atoi(q.Get("per_page"))
		if perPage < 1 || perPage > 100 {
			perPage = 20
		}

		result, err := s.SearchTypes(r.Context(), query, kind, tagFilter, ns, page, perPage)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, result)
	})
}
