package version

import "context"

//go:generate go tool mockgen -destination=../mock/version_repository.go -mock_names=Repository=VersionRepository -package mock github.com/pikoci/registry/pkreg/version Repository

type Repository interface {
	Create(ctx context.Context, v Version) (string, error)
	FindByID(ctx context.Context, id string) (*Version, error)
	FindByRegTypeAndNumber(ctx context.Context, typeID, version string) (*Version, error)
	FilterByRegType(ctx context.Context, typeID string) ([]*Version, error)
	Update(ctx context.Context, v Version) error
	Delete(ctx context.Context, id string) error
	IncrementDownloads(ctx context.Context, id string) error
}
