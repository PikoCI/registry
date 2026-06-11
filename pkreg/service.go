package pkreg

import (
	"context"

	"github.com/pikoci/registry/pkreg/namespace"
	"github.com/pikoci/registry/pkreg/regtype"
	"github.com/pikoci/registry/pkreg/tag"
	"github.com/pikoci/registry/pkreg/token"
	"github.com/pikoci/registry/pkreg/user"
	"github.com/pikoci/registry/pkreg/version"
)

//go:generate go tool mockgen -destination=mock/service.go -mock_names=Service=Service -package mock github.com/pikoci/registry/pkreg Service

type Service interface {
	// Auth
	GitHubLogin(ctx context.Context, code string) (*user.User, string, error)
	GetCurrentUser(ctx context.Context, userID string) (*user.User, error)
	GetGitHubClientID() string

	// Namespaces
	GetNamespace(ctx context.Context, name string) (*namespace.Namespace, error)
	ListNamespaceTypes(ctx context.Context, ns string) ([]*regtype.RegType, error)

	// Types
	GetType(ctx context.Context, ns, name string) (*regtype.RegType, []*version.Version, []string, error)
	SearchTypes(ctx context.Context, query, kind, tagFilter, ns string, page, perPage int) (*regtype.SearchResult, error)
	ListMyTypes(ctx context.Context, userID string) ([]*regtype.RegType, error)
	UpdateTypeTags(ctx context.Context, userID, ns, name string, tags []string) error

	// Versions
	PublishVersion(ctx context.Context, userID, ns, name, ver string, manifestBytes, readmeBytes []byte) (*version.Version, error)
	GetVersion(ctx context.Context, ns, name, ver string) (*version.Version, error)
	FetchVersionContent(ctx context.Context, ns, name, ver, tokenHash, ip string) (*version.Version, error)
	YankVersion(ctx context.Context, userID, ns, name, ver string) error
	DeprecateVersion(ctx context.Context, userID, ns, name, ver, message, successor string) error
	DeleteVersion(ctx context.Context, userID, ns, name, ver string) error
	DeleteType(ctx context.Context, userID, ns, name string) error

	// Tags
	ListTags(ctx context.Context) ([]*tag.TagCount, error)
	ListTypesByTag(ctx context.Context, tagName string) ([]*regtype.RegType, error)

	// Tokens
	CreateToken(ctx context.Context, userID, nsName, name string) (*token.Token, string, error)
	ListTokens(ctx context.Context, nsName string) ([]*token.Token, error)
	RevokeToken(ctx context.Context, userID, tokenID string) error

	// Orgs
	ClaimOrgNamespace(ctx context.Context, userID, githubOrgLogin string) (*namespace.Namespace, error)
	InviteOrgMember(ctx context.Context, userID, org, invitee, role string) error
}
