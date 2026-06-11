package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pikoci/registry/pkreg"
)

func createToken(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDContextKey).(string)

		var body struct {
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			encodeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		tok, rawToken, err := s.CreateToken(r.Context(), userID, body.Namespace, body.Name)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		encodeJSON(w, http.StatusCreated, map[string]interface{}{
			"token":     tok,
			"raw_token": rawToken,
		})
	})
}

func listTokens(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nsName := r.URL.Query().Get("namespace")

		tokens, err := s.ListTokens(r.Context(), nsName)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, tokens)
	})
}

func revokeToken(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenID := vars["token_id"]
		userID := r.Context().Value(UserIDContextKey).(string)

		if err := s.RevokeToken(r.Context(), userID, tokenID); err != nil {
			encodeError(w, http.StatusBadRequest, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
	})
}
