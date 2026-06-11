package pkreg

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pikoci/registry/pkreg/download"
	"github.com/pikoci/registry/pkreg/manifest"
	"github.com/pikoci/registry/pkreg/namespace"
	"github.com/pikoci/registry/pkreg/orgmember"
	"github.com/pikoci/registry/pkreg/regtype"
	"github.com/pikoci/registry/pkreg/tag"
	"github.com/pikoci/registry/pkreg/token"
	"github.com/pikoci/registry/pkreg/user"
	"github.com/pikoci/registry/pkreg/version"
)

type Registry struct {
	Users      user.Repository
	Namespaces namespace.Repository
	RegTypes   regtype.Repository
	Versions   version.Repository
	Tags       tag.Repository
	Tokens     token.Repository
	OrgMembers orgmember.Repository
	Downloads  download.Repository

	jwtSecret          []byte
	githubClientID     string
	githubClientSecret string
	logger             *slog.Logger
}

func New(
	ur user.Repository,
	nsr namespace.Repository,
	rtr regtype.Repository,
	vr version.Repository,
	tgr tag.Repository,
	tkr token.Repository,
	omr orgmember.Repository,
	dlr download.Repository,
	jwtSecret []byte,
	githubClientID, githubClientSecret string,
	l *slog.Logger,
) *Registry {
	return &Registry{
		Users:              ur,
		Namespaces:         nsr,
		RegTypes:           rtr,
		Versions:           vr,
		Tags:               tgr,
		Tokens:             tkr,
		OrgMembers:         omr,
		Downloads:          dlr,
		jwtSecret:          jwtSecret,
		githubClientID:     githubClientID,
		githubClientSecret: githubClientSecret,
		logger:             l,
	}
}

func (r *Registry) GetGitHubClientID() string {
	return r.githubClientID
}

func (r *Registry) GitHubLogin(ctx context.Context, code string) (*user.User, string, error) {
	// Exchange code for access token
	ghToken, err := r.exchangeGitHubCode(code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange GitHub code: %w", err)
	}

	// Fetch GitHub user
	ghUser, err := r.fetchGitHubUser(ghToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch GitHub user: %w", err)
	}

	// Upsert user
	existing, err := r.Users.FindByGitHubID(ctx, ghUser.GitHubID)
	if err != nil {
		// Create new user
		ghUser.ID = uuid.New().String()
		_, err := r.Users.Create(ctx, *ghUser)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create user: %w", err)
		}

		// Create namespace for user
		_, err = r.Namespaces.Create(ctx, namespace.Namespace{
			ID:        uuid.New().String(),
			Name:      ghUser.Username,
			Type:      "community",
			GitHubID:  ghUser.GitHubID,
			OwnerID:   ghUser.ID,
			Public:    true,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return nil, "", fmt.Errorf("failed to create namespace: %w", err)
		}
	} else {
		// Update existing user
		existing.Username = ghUser.Username
		existing.AvatarURL = ghUser.AvatarURL
		if err := r.Users.Update(ctx, *existing); err != nil {
			return nil, "", fmt.Errorf("failed to update user: %w", err)
		}
		ghUser = existing
	}

	// Sign JWT
	jwtToken, err := r.signJWT(ghUser)
	if err != nil {
		return nil, "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return ghUser, jwtToken, nil
}

func (r *Registry) GetCurrentUser(ctx context.Context, userID string) (*user.User, error) {
	return r.Users.FindByID(ctx, userID)
}

func (r *Registry) GetNamespace(ctx context.Context, name string) (*namespace.Namespace, error) {
	return r.Namespaces.FindByName(ctx, name)
}

func (r *Registry) ListNamespaceTypes(ctx context.Context, ns string) ([]*regtype.RegType, error) {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}
	types, err := r.RegTypes.FilterByNamespace(ctx, nsObj.ID)
	if err != nil {
		return nil, err
	}
	for _, t := range types {
		t.NamespaceName = nsObj.Name
	}
	return types, nil
}

func (r *Registry) GetType(ctx context.Context, ns, name string) (*regtype.RegType, []*version.Version, []string, error) {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("namespace not found: %w", err)
	}

	rt, err := r.RegTypes.FindByNamespaceAndName(ctx, nsObj.ID, name, "")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("type not found: %w", err)
	}

	versions, err := r.Versions.FilterByRegType(ctx, rt.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	tags, err := r.Tags.FilterByRegType(ctx, rt.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	return rt, versions, tags, nil
}

func (r *Registry) SearchTypes(ctx context.Context, query, kind, tagFilter, ns string, page, perPage int) (*regtype.SearchResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	result, err := r.RegTypes.Search(ctx, regtype.SearchParams{
		Query:     query,
		Kind:      kind,
		Tag:       tagFilter,
		Namespace: ns,
		Page:      page,
		PerPage:   perPage,
	})
	if err != nil {
		return nil, err
	}

	topTags, _ := r.Tags.ListWithCountsFiltered(ctx, kind, query, 10)
	result.TopTags = topTags

	return result, nil
}

func (r *Registry) ListMyTypes(ctx context.Context, userID string) ([]*regtype.RegType, error) {
	namespaces, err := r.Namespaces.FindByOwnerID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []*regtype.RegType
	for _, ns := range namespaces {
		types, err := r.RegTypes.FilterByNamespace(ctx, ns.ID)
		if err != nil {
			return nil, err
		}
		for _, t := range types {
			t.NamespaceName = ns.Name
		}
		result = append(result, types...)
	}
	return result, nil
}

func (r *Registry) UpdateTypeTags(ctx context.Context, userID, ns, name string, tags []string) error {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return fmt.Errorf("namespace not found: %w", err)
	}

	if err := r.checkNamespaceAccess(ctx, nsObj, userID); err != nil {
		return err
	}

	rt, err := r.RegTypes.FindByNamespaceAndName(ctx, nsObj.ID, name, "")
	if err != nil {
		return fmt.Errorf("type not found: %w", err)
	}

	return r.Tags.SetForRegType(ctx, rt.ID, tags)
}

func (r *Registry) PublishVersion(ctx context.Context, userID, ns, name, ver string, manifestBytes, readmeBytes []byte) (*version.Version, error) {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}

	if err := r.checkNamespaceAccess(ctx, nsObj, userID); err != nil {
		return nil, err
	}

	// Parse manifest to extract kind, description, params, examples, tags
	parsed, err := manifest.Parse(manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	if name == "" {
		name = parsed.Name
	}
	if ver == "" {
		ver = parsed.Version
	}

	if !isValidSemver(ver) {
		return nil, fmt.Errorf("invalid semver: %s", ver)
	}

	// Find or create RegType
	rt, err := r.RegTypes.FindByNamespaceAndName(ctx, nsObj.ID, name, parsed.Kind)
	if err != nil {
		// Create new type
		rt = &regtype.RegType{
			ID:          uuid.New().String(),
			NamespaceID: nsObj.ID,
			Name:        name,
			Kind:        parsed.Kind,
			Description: parsed.Description,
			Repository:  parsed.Repository,
			Homepage:    parsed.Homepage,
			License:     parsed.License,
			Public:      true,
			CreatedAt:   time.Now(),
		}
		_, err = r.RegTypes.Create(ctx, *rt)
		if err != nil {
			return nil, fmt.Errorf("failed to create type: %w", err)
		}
	} else {
		// Update description etc from latest manifest
		rt.Description = parsed.Description
		rt.Repository = parsed.Repository
		rt.Homepage = parsed.Homepage
		rt.License = parsed.License
		if err := r.RegTypes.Update(ctx, *rt); err != nil {
			return nil, fmt.Errorf("failed to update type: %w", err)
		}
	}

	// Check for duplicate version
	existing, err := r.Versions.FindByRegTypeAndNumber(ctx, rt.ID, ver)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("version %s already exists", ver)
	}

	paramsJSON, _ := json.Marshal(parsed.Params)
	examplesJSON, _ := json.Marshal(parsed.Examples)

	v := version.Version{
		ID:        uuid.New().String(),
		TypeID:    rt.ID,
		Version:   ver,
		Content:   string(manifestBytes),
		Readme:    string(readmeBytes),
		Params:    paramsJSON,
		Examples:  examplesJSON,
		CreatedAt: time.Now(),
		CreatedBy: userID,
	}

	_, err = r.Versions.Create(ctx, v)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Set tags
	if len(parsed.Tags) > 0 {
		if err := r.Tags.SetForRegType(ctx, rt.ID, parsed.Tags); err != nil {
			r.logger.Error("failed to set tags", "error", err)
		}
	}

	return &v, nil
}

func (r *Registry) GetVersion(ctx context.Context, ns, name, ver string) (*version.Version, error) {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}

	rt, err := r.RegTypes.FindByNamespaceAndName(ctx, nsObj.ID, name, "")
	if err != nil {
		return nil, fmt.Errorf("type not found: %w", err)
	}

	return r.Versions.FindByRegTypeAndNumber(ctx, rt.ID, ver)
}

func (r *Registry) FetchVersionContent(ctx context.Context, ns, name, ver, tokenHash, ip string) (*version.Version, error) {
	v, err := r.GetVersion(ctx, ns, name, ver)
	if err != nil {
		return nil, err
	}

	// Record download (deduped)
	ipHash := hashIP(ip)
	isNew, err := r.Downloads.RecordDownload(ctx, v.ID, tokenHash, ipHash, time.Now().Truncate(24*time.Hour))
	if err != nil {
		r.logger.Error("failed to record download", "error", err)
	}

	if isNew {
		if err := r.Versions.IncrementDownloads(ctx, v.ID); err != nil {
			r.logger.Error("failed to increment downloads", "error", err)
		}
		v.Downloads++
	}

	return v, nil
}

func (r *Registry) YankVersion(ctx context.Context, userID, ns, name, ver string) error {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return fmt.Errorf("namespace not found: %w", err)
	}

	if err := r.checkNamespaceAccess(ctx, nsObj, userID); err != nil {
		return err
	}

	rt, err := r.RegTypes.FindByNamespaceAndName(ctx, nsObj.ID, name, "")
	if err != nil {
		return fmt.Errorf("type not found: %w", err)
	}

	v, err := r.Versions.FindByRegTypeAndNumber(ctx, rt.ID, ver)
	if err != nil {
		return fmt.Errorf("version not found: %w", err)
	}

	v.Yanked = true
	return r.Versions.Update(ctx, *v)
}

func (r *Registry) DeprecateVersion(ctx context.Context, userID, ns, name, ver, message, successor string) error {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return fmt.Errorf("namespace not found: %w", err)
	}

	if err := r.checkNamespaceAccess(ctx, nsObj, userID); err != nil {
		return err
	}

	rt, err := r.RegTypes.FindByNamespaceAndName(ctx, nsObj.ID, name, "")
	if err != nil {
		return fmt.Errorf("type not found: %w", err)
	}

	v, err := r.Versions.FindByRegTypeAndNumber(ctx, rt.ID, ver)
	if err != nil {
		return fmt.Errorf("version not found: %w", err)
	}

	v.Deprecated = true
	v.DeprecatedMessage = message
	v.SuccessorVersion = successor
	return r.Versions.Update(ctx, *v)
}

func (r *Registry) DeleteVersion(ctx context.Context, userID, ns, name, ver string) error {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return fmt.Errorf("namespace not found: %w", err)
	}

	if err := r.checkNamespaceAccess(ctx, nsObj, userID); err != nil {
		return err
	}

	rt, err := r.RegTypes.FindByNamespaceAndName(ctx, nsObj.ID, name, "")
	if err != nil {
		return fmt.Errorf("type not found: %w", err)
	}

	v, err := r.Versions.FindByRegTypeAndNumber(ctx, rt.ID, ver)
	if err != nil {
		return fmt.Errorf("version not found: %w", err)
	}

	return r.Versions.Delete(ctx, v.ID)
}

func (r *Registry) DeleteType(ctx context.Context, userID, ns, name string) error {
	nsObj, err := r.Namespaces.FindByName(ctx, ns)
	if err != nil {
		return fmt.Errorf("namespace not found: %w", err)
	}

	if err := r.checkNamespaceAccess(ctx, nsObj, userID); err != nil {
		return err
	}

	rt, err := r.RegTypes.FindByNamespaceAndName(ctx, nsObj.ID, name, "")
	if err != nil {
		return fmt.Errorf("type not found: %w", err)
	}

	// Delete all versions first
	versions, err := r.Versions.FilterByRegType(ctx, rt.ID)
	if err != nil {
		return err
	}
	for _, v := range versions {
		if err := r.Versions.Delete(ctx, v.ID); err != nil {
			return err
		}
	}

	// Delete tags
	_ = r.Tags.SetForRegType(ctx, rt.ID, nil)

	return r.RegTypes.Delete(ctx, rt.ID)
}

func (r *Registry) ListTags(ctx context.Context) ([]*tag.TagCount, error) {
	return r.Tags.ListWithCounts(ctx)
}

func (r *Registry) ListTypesByTag(ctx context.Context, tagName string) ([]*regtype.RegType, error) {
	typeIDs, err := r.Tags.FilterRegTypesByTag(ctx, tagName)
	if err != nil {
		return nil, err
	}

	var result []*regtype.RegType
	for _, id := range typeIDs {
		rt, err := r.RegTypes.FindByID(ctx, id)
		if err != nil {
			continue
		}
		ns, err := r.Namespaces.FindByID(ctx, rt.NamespaceID)
		if err == nil {
			rt.NamespaceName = ns.Name
		}
		result = append(result, rt)
	}
	return result, nil
}

func (r *Registry) CreateToken(ctx context.Context, userID, nsName, name string) (*token.Token, string, error) {
	nsObj, err := r.Namespaces.FindByName(ctx, nsName)
	if err != nil {
		return nil, "", fmt.Errorf("namespace not found: %w", err)
	}

	if err := r.checkNamespaceAccess(ctx, nsObj, userID); err != nil {
		return nil, "", err
	}

	rawValue := uuid.New().String()
	hash := sha256Hash(rawValue)
	prefix := rawValue[:8]

	t := token.Token{
		ID:          uuid.New().String(),
		NamespaceID: nsObj.ID,
		Name:        name,
		TokenHash:   hash,
		Prefix:      prefix,
		CreatedAt:   time.Now(),
	}

	_, err = r.Tokens.Create(ctx, t)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create token: %w", err)
	}

	return &t, rawValue, nil
}

func (r *Registry) ListTokens(ctx context.Context, nsName string) ([]*token.Token, error) {
	nsObj, err := r.Namespaces.FindByName(ctx, nsName)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}
	return r.Tokens.FilterByNamespace(ctx, nsObj.ID)
}

func (r *Registry) RevokeToken(ctx context.Context, userID, tokenID string) error {
	return r.Tokens.Delete(ctx, tokenID)
}

func (r *Registry) ClaimOrgNamespace(ctx context.Context, userID, githubOrgLogin string) (*namespace.Namespace, error) {
	// Verify user is member of the GitHub org
	u, err := r.Users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// In production, this would verify org membership via GitHub API
	_ = u

	ns := namespace.Namespace{
		ID:        uuid.New().String(),
		Name:      githubOrgLogin,
		Type:      "community",
		OwnerID:   userID,
		Public:    true,
		CreatedAt: time.Now(),
	}

	_, err = r.Namespaces.Create(ctx, ns)
	if err != nil {
		return nil, fmt.Errorf("failed to create org namespace: %w", err)
	}

	// Add user as admin
	err = r.OrgMembers.Create(ctx, orgmember.OrgMember{
		OrgID:     ns.ID,
		UserID:    userID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add org admin: %w", err)
	}

	return &ns, nil
}

func (r *Registry) InviteOrgMember(ctx context.Context, userID, org, invitee, role string) error {
	nsObj, err := r.Namespaces.FindByName(ctx, org)
	if err != nil {
		return fmt.Errorf("org not found: %w", err)
	}

	// Verify inviter is admin
	member, err := r.OrgMembers.FindByOrgAndUser(ctx, nsObj.ID, userID)
	if err != nil || member.Role != "admin" {
		return fmt.Errorf("only admins can invite members")
	}

	inviteeUser, err := r.Users.FindByID(ctx, invitee)
	if err != nil {
		return fmt.Errorf("invitee not found: %w", err)
	}

	return r.OrgMembers.Create(ctx, orgmember.OrgMember{
		OrgID:     nsObj.ID,
		UserID:    inviteeUser.ID,
		Role:      role,
		CreatedAt: time.Now(),
	})
}

// checkNamespaceAccess verifies the user owns the namespace or is an org member.
func (r *Registry) checkNamespaceAccess(ctx context.Context, ns *namespace.Namespace, userID string) error {
	if ns.OwnerID == userID {
		return nil
	}

	// Check org membership
	_, err := r.OrgMembers.FindByOrgAndUser(ctx, ns.ID, userID)
	if err != nil {
		return fmt.Errorf("access denied: you do not own namespace %q", ns.Name)
	}
	return nil
}

func (r *Registry) signJWT(u *user.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  u.ID,
		"username": u.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(r.jwtSecret)
}

type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (r *Registry) exchangeGitHubCode(code string) (string, error) {
	body := fmt.Sprintf(`{"client_id":"%s","client_secret":"%s","code":"%s"}`,
		r.githubClientID, r.githubClientSecret, code)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp githubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}

	return tokenResp.AccessToken, nil
}

type githubUserResponse struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func (r *Registry) fetchGitHubUser(accessToken string) (*user.User, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ghUser githubUserResponse
	if err := json.Unmarshal(bodyBytes, &ghUser); err != nil {
		return nil, err
	}

	return &user.User{
		GitHubID:  fmt.Sprintf("%d", ghUser.ID),
		Username:  ghUser.Login,
		AvatarURL: ghUser.AvatarURL,
	}, nil
}

func isValidSemver(v string) bool {
	v = strings.TrimPrefix(v, "v")
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

func hashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(h[:])
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
