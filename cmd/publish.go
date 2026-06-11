package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/pikoci/registry/pkreg/manifest"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish [directory|file]",
	Short: "Publish a type to the registry",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}

		var manifestPath string
		var dir string

		info, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("cannot access %s: %w", target, err)
		}

		if info.IsDir() {
			dir = target
			hclFiles, _ := filepath.Glob(filepath.Join(dir, "*.hcl"))
			if len(hclFiles) == 0 {
				return fmt.Errorf("no .hcl file found in %s", dir)
			}
			if len(hclFiles) > 1 {
				return fmt.Errorf("multiple .hcl files found in %s, expected exactly one", dir)
			}
			manifestPath = hclFiles[0]
		} else {
			manifestPath = target
			dir = filepath.Dir(target)
		}

		manifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", manifestPath, err)
		}
		fmt.Printf("\u2713 %s found\n", filepath.Base(manifestPath))

		// Parse manifest to get name and version
		m, err := manifest.Parse(manifestData)
		if err != nil {
			return fmt.Errorf("invalid manifest: %w", err)
		}

		versionOverride, _ := cmd.Flags().GetString("version")
		version := m.Version
		if versionOverride != "" {
			version = versionOverride
		}
		if version == "" {
			return fmt.Errorf("version not specified in manifest; use --version flag")
		}

		var readmeData []byte
		readmePath := filepath.Join(dir, "README.md")
		if data, err := os.ReadFile(readmePath); err == nil {
			readmeData = data
			fmt.Println("\u2713 README.md found \u2014 bundling documentation")
		} else {
			fmt.Println("\u2139 no README found \u2014 publishing without documentation")
		}

		client := newRegistryClient()

		// Get namespace from JWT token
		namespace, err := usernameFromToken(client.token)
		if err != nil {
			return fmt.Errorf("not logged in (run 'pkreg login' first): %w", err)
		}

		// Build multipart request
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)

		part, err := w.CreateFormFile("manifest", filepath.Base(manifestPath))
		if err != nil {
			return err
		}
		part.Write(manifestData)

		if readmeData != nil {
			part, err = w.CreateFormFile("readme", "README.md")
			if err != nil {
				return err
			}
			part.Write(readmeData)
		}

		w.Close()

		publishPath := fmt.Sprintf("/api/plugins/%s/%s/%s", namespace, m.Name, version)
		resp, err := client.doRaw("POST", publishPath, &buf, w.FormDataContentType())
		if err != nil {
			return fmt.Errorf("publish failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("\u2717 Publish failed: %s\n", string(body))
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		fmt.Printf("\u2713 Published %s/%s@%s\n", namespace, m.Name, version)
		return nil
	},
}

func init() {
	publishCmd.Flags().String("version", "", "Override manifest version")
}

// usernameFromToken extracts the username from a JWT token without verification.
func usernameFromToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	payload := parts[1]
	// Add padding if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	var claims struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", fmt.Errorf("failed to parse token claims: %w", err)
	}

	if claims.Username == "" {
		return "", fmt.Errorf("no username in token")
	}

	return claims.Username, nil
}
