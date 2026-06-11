package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pikoci/registry/pkreg"
	"github.com/pikoci/registry/pkreg/transport/http/assets"
	"github.com/pikoci/registry/pkreg/transport/http/templates"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

func Handler(s pkreg.Service, jwtSecret []byte, l *slog.Logger) http.Handler {
	r := mux.NewRouter()

	// Rate limiter: 60 req/min per IP for public API
	rl := NewRateLimiter(1, 60, 5*time.Minute)

	// Auth middleware
	auth := authMiddleware(jwtSecret, l)

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(rl.Middleware)

	// Public routes
	api.Methods(http.MethodGet).Path("/plugins").Handler(searchTypes(s))
	api.Methods(http.MethodGet).Path("/plugins/{namespace}").Handler(listNamespaceTypes(s))
	api.Methods(http.MethodGet).Path("/plugins/{namespace}/{name}").Handler(getType(s))
	api.Methods(http.MethodGet).Path("/plugins/{namespace}/{name}/{version}").Handler(fetchVersion(s))
	api.Methods(http.MethodGet).Path("/tags").Handler(listTags(s))
	api.Methods(http.MethodGet).Path("/tags/{tag}").Handler(listTypesByTag(s))
	api.Methods(http.MethodGet).Path("/auth/github/client-id").Handler(githubClientID(s))
	api.Methods(http.MethodGet).Path("/auth/github").Handler(githubRedirect(s))
	api.Methods(http.MethodGet).Path("/auth/github/callback").Handler(githubCallback(s))
	api.Methods(http.MethodPost).Path("/auth/github").Handler(githubAuth(s))

	// Authenticated routes
	authed := api.PathPrefix("/").Subrouter()
	authed.Use(auth)

	authed.Methods(http.MethodPost).Path("/plugins/{namespace}/{name}/{version}").Handler(publishVersion(s))
	authed.Methods(http.MethodPost).Path("/plugins/{namespace}/{name}/{version}/yank").Handler(yankVersion(s))
	authed.Methods(http.MethodPost).Path("/plugins/{namespace}/{name}/{version}/deprecate").Handler(deprecateVersion(s))
	authed.Methods(http.MethodDelete).Path("/plugins/{namespace}/{name}/{version}").Handler(deleteVersion(s))
	authed.Methods(http.MethodDelete).Path("/plugins/{namespace}/{name}").Handler(deleteType(s))
	authed.Methods(http.MethodPatch).Path("/plugins/{namespace}/{name}/tags").Handler(updateTypeTags(s))
	authed.Methods(http.MethodGet).Path("/me/types").Handler(listMyTypes(s))
	authed.Methods(http.MethodPost).Path("/me/tokens").Handler(createToken(s))
	authed.Methods(http.MethodGet).Path("/me/tokens").Handler(listTokens(s))
	authed.Methods(http.MethodDelete).Path("/me/tokens/{token_id}").Handler(revokeToken(s))
	authed.Methods(http.MethodPost).Path("/orgs").Handler(claimOrgNamespace(s))
	authed.Methods(http.MethodPost).Path("/orgs/{org}/members").Handler(inviteOrgMember(s))

	// Static assets
	r.PathPrefix("/css/").Handler(http.FileServer(http.FS(assets.Assets)))
	r.PathPrefix("/js/").Handler(http.FileServer(http.FS(assets.Assets)))
	r.PathPrefix("/images/").Handler(http.FileServer(http.FS(assets.Assets)))
	r.PathPrefix("/fonts/").Handler(http.FileServer(http.FS(assets.Assets)))

	// SPA fallback
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, ok := templates.Templates["views/layouts/index.tmpl"]
		if !ok {
			http.Error(w, "template not found", http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	})

	return r
}

// encodeJSON writes a JSON response with the given status code.
func encodeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// encodeError writes a JSON error response.
func encodeError(w http.ResponseWriter, status int, msg string) {
	encodeJSON(w, status, map[string]string{"error": msg})
}
