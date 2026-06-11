//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_MissingHeader(t *testing.T) {
	env := setupTestEnv(t)

	// Authenticated endpoint without token
	resp := doGet(t, env.server.URL, "/api/me/types")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestAuth_InvalidToken(t *testing.T) {
	env := setupTestEnv(t)

	resp := doAuthGet(t, env.server.URL, "/api/me/types", "not-a-jwt")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestAuth_ValidToken(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	resp := doAuthGet(t, env.server.URL, "/api/me/types", token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestAuth_ExpiredToken(t *testing.T) {
	env := setupTestEnv(t)
	seedUser(t, env, "alice")

	// Create a token with past expiry
	expiredToken := signExpiredJWT("user-1", "alice")
	resp := doAuthGet(t, env.server.URL, "/api/me/types", expiredToken)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestListMyTypes(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	// Publish a type
	manifest := sampleManifest("pg", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// List my types
	resp = doAuthGet(t, env.server.URL, "/api/me/types", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var types []map[string]interface{}
	readJSON(t, resp, &types)
	assert.Len(t, types, 1)
	assert.Equal(t, "pg", types[0]["Name"])
}
