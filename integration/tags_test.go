//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTags_Empty(t *testing.T) {
	env := setupTestEnv(t)

	resp := doGet(t, env.server.URL, "/api/tags")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestListTags_WithData(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	// Publish a type (sampleManifest adds tags = ["test", "<kind>"])
	manifest := sampleManifest("pg", "1.0.0", "resource_type")
	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0", manifest, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// List tags
	resp = doGet(t, env.server.URL, "/api/tags")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tags []map[string]interface{}
	readJSON(t, resp, &tags)
	assert.GreaterOrEqual(t, len(tags), 1)

	// At least "test" tag should exist
	found := false
	for _, tag := range tags {
		if tag["Tag"] == "test" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected 'test' tag")
}

func TestListTypesByTag(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0",
		sampleManifest("pg", "1.0.0", "resource_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// List types by tag
	resp = doGet(t, env.server.URL, "/api/tags/test")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var types []map[string]interface{}
	readJSON(t, resp, &types)
	assert.Len(t, types, 1)
	assert.Equal(t, "pg", types[0]["Name"])
}

func TestUpdateTags(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "alice")

	resp := publishManifest(t, env, token, "alice", "pg", "1.0.0",
		sampleManifest("pg", "1.0.0", "resource_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Update tags
	body := map[string]interface{}{
		"tags": []string{"database", "sql", "postgresql"},
	}
	resp = doAuthPatch(t, env.server.URL, "/api/plugins/alice/pg/tags", token, body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify tags changed
	resp = doGet(t, env.server.URL, "/api/plugins/alice/pg")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	readJSON(t, resp, &result)
	tags, ok := result["tags"].([]interface{})
	if ok {
		assert.Len(t, tags, 3)
	}
}

func TestUpdateTags_NotOwner(t *testing.T) {
	env := setupTestEnv(t)
	_, aliceToken := seedUser(t, env, "alice")
	_, bobToken := seedUser(t, env, "bob")

	resp := publishManifest(t, env, aliceToken, "alice", "pg", "1.0.0",
		sampleManifest("pg", "1.0.0", "resource_type"), nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Bob tries to update alice's tags
	body := map[string]interface{}{
		"tags": []string{"hacked"},
	}
	resp = doAuthPatch(t, env.server.URL, "/api/plugins/alice/pg/tags", bobToken, body)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}
