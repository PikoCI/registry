package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the registry via GitHub OAuth",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := newRegistryClient()

		// Start a local server to receive the callback
		tokenCh := make(chan string, 1)
		errCh := make(chan error, 1)

		mux := http.NewServeMux()
		mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			token := r.URL.Query().Get("token")
			if token == "" {
				errCh <- fmt.Errorf("no token received")
				fmt.Fprintf(w, "Login failed. You can close this window.")
				return
			}
			tokenCh <- token
			fmt.Fprintf(w, "Login successful! You can close this window.")
		})

		srv := &http.Server{Addr: "127.0.0.1:0", Handler: mux}
		ln, err := (&net.ListenConfig{}).Listen(cmd.Context(), "tcp", "127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("failed to start local server: %w", err)
		}
		defer ln.Close()

		port := ln.Addr().(*net.TCPAddr).Port

		go srv.Serve(ln)
		defer srv.Close()

		url := fmt.Sprintf("%s/api/auth/github?cli=true&port=%d", client.baseURL, port)
		fmt.Printf("Opening browser for login...\n")
		openBrowser(url)

		select {
		case token := <-tokenCh:
			tokenPath := tokenFilePath()
			if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
			fmt.Println("\u2713 Logged in successfully")
			return nil
		case err := <-errCh:
			return err
		}
	},
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
