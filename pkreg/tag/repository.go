package tag

import "context"

//go:generate go tool mockgen -destination=../mock/tag_repository.go -mock_names=Repository=TagRepository -package mock github.com/pikoci/registry/pkreg/tag Repository

type Repository interface {
	SetForRegType(ctx context.Context, typeID string, tags []string) error
	FilterByRegType(ctx context.Context, typeID string) ([]string, error)
	ListWithCounts(ctx context.Context) ([]*TagCount, error)
	ListWithCountsFiltered(ctx context.Context, kind, query string, limit int) ([]*TagCount, error)
	FilterRegTypesByTag(ctx context.Context, tag string) ([]string, error)
}
