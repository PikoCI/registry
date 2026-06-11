//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pikoci/registry/pkreg"
	"github.com/pikoci/registry/pkreg/db"
	"github.com/pikoci/registry/pkreg/db/migrate"
	"github.com/pikoci/registry/pkreg/namespace"
	"github.com/pikoci/registry/pkreg/user"
	tshttp "github.com/pikoci/registry/pkreg/transport/http"
)

var testJWTSecret = []byte("test-jwt-secret")

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

type testEnv struct {
	server *httptest.Server
	svc    pkreg.Service
	db     *db.UserRepository
	nsDB   *db.NamespaceRepository
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Use a unique SQLite file per test to avoid shared state
	dbFile := fmt.Sprintf("%s/pkreg-test-%s.db", t.TempDir(), uuid.New().String()[:8])
	database, err := db.New("", 0, "", "", db.Options{System: db.SQLite, DBFile: dbFile})
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	if err := migrate.Migrate(database, db.SQLite); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	ur := db.NewUserRepository(database)
	nsr := db.NewNamespaceRepository(database)
	rtr := db.NewRegTypeRepository(database)
	vr := db.NewVersionRepository(database)
	tgr := db.NewTagRepository(database)
	tkr := db.NewTokenRepository(database)
	omr := db.NewOrgMemberRepository(database)
	dlr := db.NewDownloadRepository(database)

	svc := pkreg.New(ur, nsr, rtr, vr, tgr, tkr, omr, dlr, testJWTSecret, "", "", logger)

	handler := tshttp.Handler(svc, testJWTSecret, logger)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &testEnv{server: server, svc: svc, db: ur, nsDB: nsr}
}

// seedUser creates a user + namespace in the DB and returns a JWT token.
func seedUser(t *testing.T, env *testEnv, username string) (userID, token string) {
	t.Helper()
	userID = uuid.New().String()
	nsID := uuid.New().String()

	_, err := env.db.Create(context.Background(), user.User{
		ID:        userID,
		GitHubID:  "gh-" + username,
		Username:  username,
		AvatarURL: "https://example.com/" + username + ".png",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	_, err = env.nsDB.Create(context.Background(), namespace.Namespace{
		ID:        nsID,
		Name:      username,
		Type:      "community",
		GitHubID:  "gh-" + username,
		OwnerID:   userID,
		Public:    true,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to seed namespace: %v", err)
	}

	token = signTestJWT(userID, username)
	return userID, token
}

func signTestJWT(userID, username string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      float64(9999999999),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := tok.SignedString(testJWTSecret)
	return tokenString
}

func signExpiredJWT(userID, username string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      float64(1000000000), // 2001 - long expired
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := tok.SignedString(testJWTSecret)
	return tokenString
}

// publishManifest publishes a version via the multipart API.
func publishManifest(t *testing.T, env *testEnv, token, ns, name, version string, manifest, readme []byte) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, _ := w.CreateFormFile("manifest", "pikoci.hcl")
	part.Write(manifest)

	if readme != nil {
		part, _ = w.CreateFormFile("readme", "README.md")
		part.Write(readme)
	}

	w.Close()

	path := fmt.Sprintf("/api/plugins/%s/%s/%s", ns, name, version)
	req, _ := http.NewRequest("POST", env.server.URL+path, &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s failed: %v", path, err)
	}
	return resp
}

func doGet(t *testing.T, serverURL, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(serverURL + path)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}
	return resp
}

func doAuthGet(t *testing.T, serverURL, path, token string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest("GET", serverURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}
	return resp
}

func doAuthPost(t *testing.T, serverURL, path, token string, body interface{}) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = strings.NewReader(string(b))
	}
	req, _ := http.NewRequest("POST", serverURL+path, reader)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s failed: %v", path, err)
	}
	return resp
}

func doAuthPatch(t *testing.T, serverURL, path, token string, body interface{}) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = strings.NewReader(string(b))
	}
	req, _ := http.NewRequest("PATCH", serverURL+path, reader)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH %s failed: %v", path, err)
	}
	return resp
}

func doAuthDelete(t *testing.T, serverURL, path, token string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest("DELETE", serverURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s failed: %v", path, err)
	}
	return resp
}

func readJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("failed to decode response (status %d): %v (body: %s)", resp.StatusCode, err, string(body))
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}

func sampleManifest(name, version, kind string) []byte {
	return []byte(fmt.Sprintf(`name = "%s"
version = "%s"
description = "A test %s"
repository = "https://github.com/test/%s"
license = "Apache-2.0"
tags = ["test", "%s"]

%s "%s" {
  check {}
}
`, name, version, kind, name, kind, kind, name))
}
