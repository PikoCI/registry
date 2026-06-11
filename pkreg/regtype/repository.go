package regtype

import (
	"context"

	"github.com/pikoci/registry/pkreg/tag"
)

//go:generate go tool mockgen -destination=../mock/regtype_repository.go -mock_names=Repository=RegTypeRepository -package mock github.com/pikoci/registry/pkreg/regtype Repository

type SearchParams struct {
	Query     string
	Kind      string
	Tag       string
	Namespace string
	Page      int
	PerPage   int
}

type SearchResult struct {
	Types      []*RegType
	TotalCount int
	TopTags    []*tag.TagCount `json:",omitempty"`
}

type Repository interface {
	Create(ctx context.Context, rt RegType) (string, error)
	FindByID(ctx context.Context, id string) (*RegType, error)
	FindByNamespaceAndName(ctx context.Context, nsID, name, kind string) (*RegType, error)
	FilterByNamespace(ctx context.Context, nsID string) ([]*RegType, error)
	Search(ctx context.Context, params SearchParams) (*SearchResult, error)
	Update(ctx context.Context, rt RegType) error
	Delete(ctx context.Context, id string) error
}
