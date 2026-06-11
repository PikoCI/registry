package manifest_test

import (
	"testing"

	"github.com/pikoci/registry/pkreg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_ResourceType(t *testing.T) {
	data := []byte(`name = "postgres"
version = "1.0.0"
description = "PostgreSQL resource type"
repository = "https://github.com/example/postgres"
license = "Apache-2.0"
tags = ["database", "postgresql"]

resource_type "postgres" {
  check {}
  pull {}
}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "postgres", m.Name)
	assert.Equal(t, "1.0.0", m.Version)
	assert.Equal(t, "resource_type", m.Kind)
	assert.Equal(t, "PostgreSQL resource type", m.Description)
	assert.Equal(t, "Apache-2.0", m.License)
	assert.Equal(t, []string{"database", "postgresql"}, m.Tags)
}

func TestParse_RunnerType(t *testing.T) {
	data := []byte(`name = "docker"
version = "2.1.0"
description = "Docker runner"

runner_type "docker" {
  run {}
}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "runner_type", m.Kind)
}

func TestParse_ServiceType(t *testing.T) {
	data := []byte(`name = "redis"
version = "1.0.0"

service_type "redis" {
  start {}
}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "service_type", m.Kind)
}

func TestParse_SecretType(t *testing.T) {
	data := []byte(`name = "vault"
version = "1.0.0"

secret_type "vault" {
  get {}
}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "secret_type", m.Kind)
}

func TestParse_NotificationType(t *testing.T) {
	data := []byte(`name = "slack"
version = "1.0.0"

notification_type "slack" {
  notify {}
}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "notification_type", m.Kind)
}

func TestParse_NameFromBlockLabel(t *testing.T) {
	data := []byte(`version = "1.0.0"

resource_type "myname" {}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "myname", m.Name)
}

func TestParse_TopLevelNameOverridesBlockLabel(t *testing.T) {
	data := []byte(`name = "override"
version = "1.0.0"

resource_type "blocklabel" {}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "override", m.Name)
}

func TestParse_NoVersion(t *testing.T) {
	data := []byte(`resource_type "test" {}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "test", m.Name)
	assert.Equal(t, "", m.Version)
}

func TestParse_InvalidSemver(t *testing.T) {
	data := []byte(`name = "test"
version = "not-a-version"

resource_type "test" {}
`)
	_, err := manifest.Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid semver")
}

func TestParse_NoImplementationBlock(t *testing.T) {
	data := []byte(`name = "test"
version = "1.0.0"
`)
	_, err := manifest.Parse(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no implementation block found")
}

func TestParse_SemverWithPrerelease(t *testing.T) {
	data := []byte(`name = "test"
version = "1.0.0-rc.1"

resource_type "test" {}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0-rc.1", m.Version)
}

func TestParse_WithParams(t *testing.T) {
	data := []byte(`resource_type "git" {
  params = ["url", "branch", "token"]
  check {}
}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "git", m.Name)
	assert.Len(t, m.Params, 3)
	assert.Equal(t, "url", m.Params[0].Name)
	assert.Equal(t, "branch", m.Params[1].Name)
	assert.Equal(t, "token", m.Params[2].Name)
}

func TestParse_WithReadmePath(t *testing.T) {
	data := []byte(`name = "test"
version = "1.0.0"
readme = "docs/README.md"

resource_type "test" {}
`)
	m, err := manifest.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "docs/README.md", m.ReadmePath)
}
