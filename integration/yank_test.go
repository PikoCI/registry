//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYank(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("oldpkg", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "oldpkg", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Yank
	resp = doAuthPost(t, env.server.URL, "/api/plugins/alice/oldpkg/1.0.0/yank", token, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify yanked
	resp = doGet(t, env.server.URL, "/api/plugins/alice/oldpkg/1.0.0")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, true, result["Yanked"])
}

func TestYank_StillServed(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("yanked", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "yanked", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Yank
	resp = doAuthPost(t, env.server.URL, "/api/plugins/alice/yanked/1.0.0/yank", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Still fetchable
	resp = doGet(t, env.server.URL, "/api/plugins/alice/yanked/1.0.0")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.NotEmpty(t, result["Content"])
	assert.Equal(t, true, result["Yanked"])
}

func TestYank_NotOwner(t *testing.T) {
	env := setupTestEnv(t)
	_, aliceToken := seedUser(t, env, "alice")
	_, bobToken := seedUser(t, env, "bob")

	manifest := sampleManifest("mypkg", "1.0.0", "resource_type")
	resp := publishManifest(t, env, aliceToken, "alice", "mypkg", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Bob tries to yank alice's version
	resp = doAuthPost(t, env.server.URL, "/api/plugins/alice/mypkg/1.0.0/yank", bobToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestYank_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	resp := doAuthPost(t, env.server.URL, "/api/plugins/alice/nosuch/1.0.0/yank", token, nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}
