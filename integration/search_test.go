//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_Empty(t *testing.T) {
	env := setupTestEnv(t)

	resp := doGet(t, env.server.URL, "/api/plugins?q=nonexistent")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, float64(0), result["TotalCount"])
}

func TestSearch_ByQuery(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	// Publish a type
	manifest := sampleManifest("postgres", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "postgres", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Search for it
	resp = doGet(t, env.server.URL, "/api/plugins?q=postgres")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, float64(1), result["TotalCount"])
	types := result["Types"].([]interface{})
	assert.Len(t, types, 1)
}

func TestSearch_ByKind(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	// Publish resource_type and runner_type
	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0",
		sampleManifest("pg", "1.0.0", "resource_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	resp = publishManifest(t, env, token, "alice", "docker", "1.0.0",
		sampleManifest("docker", "1.0.0", "runner_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Search by kind
	resp = doGet(t, env.server.URL, "/api/plugins?kind=runner_type")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, float64(1), result["TotalCount"])
}

func TestSearch_ByTag(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	manifest := sampleManifest("pg", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Search by tag (sampleManifest adds tags = ["test", "<kind>"])
	resp = doGet(t, env.server.URL, "/api/plugins?tag=test")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, float64(1), result["TotalCount"])
}

func TestSearch_ByNamespace(t *testing.T) {
	env := setupTestEnv(t)
	_, aliceToken := seedUser(t, env, "alice")
	_, bobToken := seedUser(t, env, "bob")

	resp := publishManifest(t, env, aliceToken, "alice", "pg", "1.0.0",
		sampleManifest("pg", "1.0.0", "resource_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	resp = publishManifest(t, env, bobToken, "bob", "pg", "1.0.0",
		sampleManifest("pg", "1.0.0", "resource_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Search scoped to alice
	resp = doGet(t, env.server.URL, "/api/plugins?namespace=alice")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, float64(1), result["TotalCount"])
}

func TestSearch_Pagination(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	// Publish 3 types
	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("type%d", i)
		resp := publishManifest(t, env, token, "alice", name, "1.0.0",
			sampleManifest(name, "1.0.0", "resource_type"), nil)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}

	// Page 1, size 2
	resp := doGet(t, env.server.URL, "/api/plugins?per_page=2&page=1")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, float64(3), result["TotalCount"])
	types := result["Types"].([]interface{})
	assert.Len(t, types, 2)

	// Page 2, size 2
	resp = doGet(t, env.server.URL, "/api/plugins?per_page=2&page=2")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	readJSON(t, resp, &result)
	types = result["Types"].([]interface{})
	assert.Len(t, types, 1)
}

func TestSearch_Combined(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0",
		sampleManifest("pg", "1.0.0", "resource_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	resp = publishManifest(t, env, token, "alice", "docker", "1.0.0",
		sampleManifest("docker", "1.0.0", "runner_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Combined: query + kind
	resp = doGet(t, env.server.URL, "/api/plugins?q=pg&kind=resource_type")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	assert.Equal(t, float64(1), result["TotalCount"])
}
