//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListNamespaceTypes(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	// Publish two types in alice's namespace
	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0",
		sampleManifest("pg", "1.0.0", "resource_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	resp = publishManifest(t, env, token, "alice", "docker", "1.0.0",
		sampleManifest("docker", "1.0.0", "runner_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// List
	resp = doGet(t, env.server.URL, "/api/plugins/alice")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var types []map[string]interface{}
	readJSON(t, resp, &types)
	assert.Len(t, types, 2)
}

func TestListNamespaceTypes_Empty(t *testing.T) {
	env := setupTestEnv(t)
	seedUser(t, env, "alice")

	resp := doGet(t, env.server.URL, "/api/plugins/alice")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestListNamespaceTypes_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	resp := doGet(t, env.server.URL, "/api/plugins/nosuchnamespace")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestGetType(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("pg", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Publish a second version
	manifest2 := sampleManifest("pg", "2.0.0", "resource_type")
	resp = publishManifest(t, env, token, "alice", "pg", "2.0.0", manifest2, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Get type details
	resp = doGet(t, env.server.URL, "/api/plugins/alice/pg")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, "pg", result["type"].(map[string]interface{})["Name"])

	versions, ok := result["versions"].([]interface{})
	if ok {
		assert.Len(t, versions, 2)
	}
}

func TestGetType_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	seedUser(t, env, "alice")

	resp := doGet(t, env.server.URL, "/api/plugins/alice/nosuchtype")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}
