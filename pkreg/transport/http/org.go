package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/pikoci/registry/pkreg"
)

func githubClientID(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodeJSON(w, http.StatusOK, map[string]string{"client_id": s.GetGitHubClientID()})
	})
}

func githubRedirect(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := s.GetGitHubClientID()
		if clientID == "" {
			encodeError(w, http.StatusInternalServerError, "GitHub OAuth not configured")
			return
		}

		// Build the redirect URI — either for CLI or web
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		redirectURI := fmt.Sprintf("%s://%s/api/auth/github/callback", scheme, r.Host)

		// Pass CLI params through via state
		state := ""
		if r.URL.Query().Get("cli") == "true" {
			port := r.URL.Query().Get("port")
			state = "cli:" + port
		}

		ghURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user&state=%s",
			url.QueryEscape(clientID),
			url.QueryEscape(redirectURI),
			url.QueryEscape(state),
		)
		http.Redirect(w, r, ghURL, http.StatusFound)
	})
}

func githubCallback(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			encodeError(w, http.StatusBadRequest, "code is required")
			return
		}

		u, token, err := s.GitHubLogin(r.Context(), code)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		state := r.URL.Query().Get("state")

		// CLI flow: redirect token to the local CLI server
		if len(state) > 4 && state[:4] == "cli:" {
			port := state[4:]
			callbackURL := fmt.Sprintf("http://127.0.0.1:%s/callback?token=%s", port, url.QueryEscape(token))
			http.Redirect(w, r, callbackURL, http.StatusFound)
			return
		}

		// Web flow: return JSON with user and token for the SPA to consume
		// Redirect to the login/callback page with the token in a fragment
		userJSON, _ := json.Marshal(u)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html><html><body><script>
			localStorage.setItem('pkreg-token', %q);
			localStorage.setItem('pkreg-user', %q);
			window.location.href = '/me/types';
		</script></body></html>`, token, string(userJSON))
	})
}

func githubAuth(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			encodeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if body.Code == "" {
			encodeError(w, http.StatusBadRequest, "code is required")
			return
		}

		u, token, err := s.GitHubLogin(r.Context(), body.Code)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		encodeJSON(w, http.StatusOK, map[string]interface{}{
			"user":  u,
			"token": token,
		})
	})
}

func claimOrgNamespace(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDContextKey).(string)

		var body struct {
			GithubOrgLogin string `json:"github_org_login"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			encodeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if body.GithubOrgLogin == "" {
			encodeError(w, http.StatusBadRequest, "github_org_login is required")
			return
		}

		ns, err := s.ClaimOrgNamespace(r.Context(), userID, body.GithubOrgLogin)
		if err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		encodeJSON(w, http.StatusCreated, ns)
	})
}

func inviteOrgMember(s pkreg.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		org := vars["org"]
		userID := r.Context().Value(UserIDContextKey).(string)

		var body struct {
			Invitee string `json:"invitee"`
			Role    string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			encodeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if body.Invitee == "" {
			encodeError(w, http.StatusBadRequest, "invitee is required")
			return
		}
		if body.Role == "" {
			body.Role = "member"
		}

		if err := s.InviteOrgMember(r.Context(), userID, org, body.Invitee, body.Role); err != nil {
			encodeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}
