//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToken(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	body := map[string]string{
		"namespace": "alice",
		"name":      "ci-token",
	}
	resp := doAuthPost(t, env.server.URL, "/api/me/tokens", token, body)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.NotEmpty(t, result["raw_token"])
	assert.Equal(t, "ci-token", result["token"].(map[string]interface{})["Name"])
}

func TestListTokens(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	// Create two tokens
	for _, name := range []string{"token-a", "token-b"} {
		body := map[string]string{"namespace": "alice", "name": name}
		resp := doAuthPost(t, env.server.URL, "/api/me/tokens", token, body)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}

	// List
	resp := doAuthGet(t, env.server.URL, "/api/me/tokens?namespace=alice", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokens []map[string]interface{}
	readJSON(t, resp, &tokens)
	assert.Len(t, tokens, 2)
}

func TestRevokeToken(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	body := map[string]string{"namespace": "alice", "name": "to-revoke"}
	resp := doAuthPost(t, env.server.URL, "/api/me/tokens", token, body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	tokenObj := result["token"].(map[string]interface{})
	tokenID := tokenObj["ID"].(string)

	// Revoke
	resp = doAuthDelete(t, env.server.URL, "/api/me/tokens/"+tokenID, token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// List should be empty
	resp = doAuthGet(t, env.server.URL, "/api/me/tokens?namespace=alice", token)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
