//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchVersion(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("postgres", "1.0.0", "resource_type")
	readme := []byte("# PostgreSQL")
	resp := publishManifest(t, env, token, "alice", "postgres", "1.0.0", manifest, readme)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Fetch
	resp = doGet(t, env.server.URL, "/api/plugins/alice/postgres/1.0.0")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, "1.0.0", result["Version"])
	assert.Equal(t, "# PostgreSQL", result["Readme"])
	assert.NotEmpty(t, result["Content"])
}

func TestFetchVersion_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	resp := doGet(t, env.server.URL, "/api/plugins/nosuchns/nosuchtype/1.0.0")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestFetchVersion_DownloadCounting(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("counter", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "counter", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// First fetch should count as a download
	resp = doGet(t, env.server.URL, "/api/plugins/alice/counter/1.0.0")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	downloads, _ := result["Downloads"].(float64)
	assert.Equal(t, float64(1), downloads)

	// Second fetch from same IP on same day should be deduped
	resp = doGet(t, env.server.URL, "/api/plugins/alice/counter/1.0.0")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	readJSON(t, resp, &result)
	downloads2, _ := result["Downloads"].(float64)
	assert.Equal(t, float64(1), downloads2, "download should be deduped")
}

func TestFetchVersion_Yanked(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("oldtype", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "oldtype", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Yank it
	resp = doAuthPost(t, env.server.URL, "/api/plugins/alice/oldtype/1.0.0/yank", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Should still be fetchable
	resp = doGet(t, env.server.URL, "/api/plugins/alice/oldtype/1.0.0")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, true, result["Yanked"])
}
