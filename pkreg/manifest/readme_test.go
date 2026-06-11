package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pikoci/registry/pkreg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveReadme_ExplicitPath(t *testing.T) {
	dir := t.TempDir()
	readmeContent := []byte("# My Type\nSome docs")
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)
	os.WriteFile(filepath.Join(dir, "docs", "README.md"), readmeContent, 0644)

	m := &manifest.Manifest{ReadmePath: "docs/README.md"}
	data, err := manifest.ResolveReadme(dir, m)
	require.NoError(t, err)
	assert.Equal(t, readmeContent, data)
}

func TestResolveReadme_ExplicitPathNotFound(t *testing.T) {
	dir := t.TempDir()
	m := &manifest.Manifest{ReadmePath: "docs/MISSING.md"}
	_, err := manifest.ResolveReadme(dir, m)
	assert.Error(t, err)
}

func TestResolveReadme_ConventionFound(t *testing.T) {
	dir := t.TempDir()
	readmeContent := []byte("# Convention README")
	os.WriteFile(filepath.Join(dir, "README.md"), readmeContent, 0644)

	m := &manifest.Manifest{}
	data, err := manifest.ResolveReadme(dir, m)
	require.NoError(t, err)
	assert.Equal(t, readmeContent, data)
}

func TestResolveReadme_NoReadme(t *testing.T) {
	dir := t.TempDir()
	m := &manifest.Manifest{}
	data, err := manifest.ResolveReadme(dir, m)
	require.NoError(t, err)
	assert.Nil(t, data)
}
