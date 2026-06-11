package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pikoci/registry/pkreg"
)

func getType(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]
		name := vars["name"]

		rt, versions, tags, err := s.GetType(r.Context(), ns, name)
		if err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, map[string]interface{}{
			"type":     rt,
			"versions": versions,
			"tags":     tags,
		})
	})
}

func listMyTypes(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDContextKey).(string)

		types, err := s.ListMyTypes(r.Context(), userID)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, types)
	})
}

func updateTypeTags(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ns := vars["namespace"]
		name := vars["name"]
		userID := r.Context().Value(UserIDContextKey).(string)

		var body struct {
			Tags []string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			encodeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := s.UpdateTypeTags(r.Context(), userID, ns, name, body.Tags); err != nil {
			status := http.StatusBadRequest
			if strings.Contains(err.Error(), "access denied") {
				status = http.StatusForbidden
			}
			encodeError(w, status, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	})
}
