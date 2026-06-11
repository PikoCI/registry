//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish_NewType(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("postgres", "1.0.0", "resource_type")
	readme := []byte("# PostgreSQL Resource Type\nManages PostgreSQL databases.")

	resp := publishManifest(t, env, token, "alice", "postgres", "1.0.0", manifest, readme)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, "1.0.0", result["Version"])
}

func TestPublish_NewVersion(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	// Publish v1
	manifest1 := sampleManifest("postgres", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "postgres", "1.0.0", manifest1, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Publish v2
	manifest2 := sampleManifest("postgres", "2.0.0", "resource_type")
	resp = publishManifest(t, env, token, "alice", "postgres", "2.0.0", manifest2, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()
}

func TestPublish_WithReadme(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("mytype", "1.0.0", "runner_type")
	readme := []byte("# My Runner\n\nThis is a test runner type.")

	resp := publishManifest(t, env, token, "alice", "mytype", "1.0.0", manifest, readme)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Fetch and verify readme is included
	resp = doGet(t, env.server.URL, "/api/plugins/alice/mytype/1.0.0")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, "# My Runner\n\nThis is a test runner type.", result["Readme"])
}

func TestPublish_DuplicateVersion(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("postgres", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "postgres", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Attempt duplicate
	resp = publishManifest(t, env, token, "alice", "postgres", "1.0.0", manifest, nil)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()
}

func TestPublish_WrongNamespace(t *testing.T) {
	env := setupTestEnv(t)
	seedUser(t, env, "alice")
	_, bobToken := seedUser(t, env, "bob")

	manifest := sampleManifest("postgres", "1.0.0", "resource_type")
	resp := publishManifest(t, env, bobToken, "alice", "postgres", "1.0.0", manifest, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestPublish_InvalidManifest(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	badManifest := []byte(`this is not valid HCL or a manifest`)
	resp := publishManifest(t, env, token, "alice", "broken", "1.0.0", badManifest, nil)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestPublish_Unauthenticated(t *testing.T) {
	env := setupTestEnv(t)
	seedUser(t, env, "alice")

	manifest := sampleManifest("postgres", "1.0.0", "resource_type")
	resp := publishManifest(t, env, "bad-token", "alice", "postgres", "1.0.0", manifest, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestPublish_AllFiveKinds(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	kinds := []string{"resource_type", "runner_type", "service_type", "secret_type", "notification_type"}
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			name := "test-" + kind
			manifest := sampleManifest(name, "1.0.0", kind)
			resp := publishManifest(t, env, token, "alice", name, "1.0.0", manifest, nil)
			assert.Equal(t, http.StatusCreated, resp.StatusCode, "kind: %s", kind)
			resp.Body.Close()
		})
	}
}
