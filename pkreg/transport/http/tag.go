package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pikoci/registry/pkreg"
)

func listTags(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tags, err := s.ListTags(r.Context())
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, tags)
	})
}

func listTypesByTag(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tagName := vars["tag"]

		types, err := s.ListTypesByTag(r.Context(), tagName)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, types)
	})
}
