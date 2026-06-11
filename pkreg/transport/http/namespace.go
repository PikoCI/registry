package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pikoci/registry/pkreg"
)

func listNamespaceTypes(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]

		types, err := s.ListNamespaceTypes(r.Context(), ns)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, types)
	})
}
