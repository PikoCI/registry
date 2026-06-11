package namespace

import "context"

//go:generate go tool mockgen -destination=../mock/namespace_repository.go -mock_names=Repository=NamespaceRepository -package mock github.com/pikoci/registry/pkreg/namespace Repository

type Repository interface {
	Create(ctx context.Context, ns Namespace) (string, error)
	FindByID(ctx context.Context, id string) (*Namespace, error)
	FindByName(ctx context.Context, name string) (*Namespace, error)
	FindByOwnerID(ctx context.Context, ownerID string) ([]*Namespace, error)
	Update(ctx context.Context, ns Namespace) error
	Delete(ctx context.Context, id string) error
}
