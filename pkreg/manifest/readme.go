package manifest

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveReadme resolves the README content for a manifest.
// 1. If manifest has explicit readme path, read that file
// 2. If README.md exists in the same directory, use it
// 3. Otherwise, return empty string (no error)
func ResolveReadme(manifestDir string, m *Manifest) ([]byte, error) {
	if m.ReadmePath != "" {
		path := filepath.Join(manifestDir, m.ReadmePath)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("readme file not found at %s: %w", m.ReadmePath, err)
		}
		return data, nil
	}

	readmePath := filepath.Join(manifestDir, "README.md")
	data, err := os.ReadFile(readmePath)
	if err == nil {
		return data, nil
	}

	return nil, nil
}
