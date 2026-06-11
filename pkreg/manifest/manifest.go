package manifest

import (
	"fmt"
	"strings"
)

type Manifest struct {
	Name        string
	Version     string
	Kind        string
	Description string
	Repository  string
	Homepage    string
	License     string
	Tags        []string
	ReadmePath  string
	Params      []Param
	Examples    []Example
	RawContent  []byte
}

type Param struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
	Secret      bool   `json:"secret"`
}

type Example struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
}

var validKinds = map[string]bool{
	"resource_type":     true,
	"runner_type":       true,
	"service_type":      true,
	"secret_type":       true,
	"notification_type": true,
}

// Parse parses a pikoci.hcl manifest from raw bytes.
// This uses a simple line-based parser since the registry only needs metadata,
// not full HCL evaluation.
func Parse(data []byte) (*Manifest, error) {
	content := string(data)
	m := &Manifest{RawContent: data}

	// Detect kind and block label from block type
	for kind := range validKinds {
		label := extractBlockLabel(content, kind)
		if label != "" {
			m.Kind = kind
			// Use the block label as the name if no top-level name attribute
			if m.Name == "" {
				m.Name = label
			}
			break
		}
	}
	if m.Kind == "" {
		return nil, fmt.Errorf("no implementation block found (expected one of: resource_type, runner_type, service_type, secret_type, notification_type)")
	}

	// Top-level attributes override block label
	if name := extractAttr(content, "name"); name != "" {
		m.Name = name
	}
	m.Version = extractAttr(content, "version")
	m.Description = extractAttr(content, "description")
	m.Repository = extractAttr(content, "repository")
	m.Homepage = extractAttr(content, "homepage")
	m.License = extractAttr(content, "license")
	m.ReadmePath = extractAttr(content, "readme")

	if m.Name == "" {
		return nil, fmt.Errorf("missing required field: name (set it as a top-level attribute or as the block label)")
	}

	if m.Version != "" && !isValidSemver(m.Version) {
		return nil, fmt.Errorf("invalid semver version: %s", m.Version)
	}

	// Extract params from inside the block (e.g. params = ["url", "branch"])
	paramsStr := extractAttr(content, "params")
	if paramsStr != "" {
		paramsStr = strings.Trim(paramsStr, "[]")
		for _, p := range strings.Split(paramsStr, ",") {
			p = strings.TrimSpace(p)
			p = strings.Trim(p, `"`)
			if p != "" {
				m.Params = append(m.Params, Param{Name: p})
			}
		}
	}

	// Extract tags
	tagsStr := extractAttr(content, "tags")
	if tagsStr != "" {
		tagsStr = strings.Trim(tagsStr, "[]")
		for _, t := range strings.Split(tagsStr, ",") {
			t = strings.TrimSpace(t)
			t = strings.Trim(t, `"`)
			if t != "" {
				m.Tags = append(m.Tags, t)
			}
		}
	}

	return m, nil
}

func containsBlock(content, blockType string) bool {
	return extractBlockLabel(content, blockType) != ""
}

// extractBlockLabel returns the label from a block declaration like `resource_type "git" {`.
// Returns the block type name itself if no label (e.g. `resource_type {`).
func extractBlockLabel(content, blockType string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, blockType+" ") || strings.HasPrefix(trimmed, blockType+"{") {
			// Extract quoted label: resource_type "git" {
			rest := strings.TrimPrefix(trimmed, blockType)
			rest = strings.TrimSpace(rest)
			if rest == "{" || rest == "" {
				// No label, just the block type
				return blockType
			}
			// Parse "label" from the rest
			if idx := strings.Index(rest, `"`); idx != -1 {
				end := strings.Index(rest[idx+1:], `"`)
				if end != -1 {
					return rest[idx+1 : idx+1+end]
				}
			}
			return blockType
		}
	}
	return ""
}

func extractAttr(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key+" ") || strings.HasPrefix(line, key+"=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				val := strings.TrimSpace(parts[1])
				val = strings.Trim(val, `"`)
				return val
			}
		}
	}
	return ""
}

func isValidSemver(v string) bool {
	v = strings.TrimPrefix(v, "v")
	// Strip pre-release and build metadata for basic validation
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}
