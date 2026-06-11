package token

import "context"

//go:generate go tool mockgen -destination=../mock/token_repository.go -mock_names=Repository=TokenRepository -package mock github.com/pikoci/registry/pkreg/token Repository

type Repository interface {
	Create(ctx context.Context, t Token) (string, error)
	FindByHash(ctx context.Context, hash string) (*Token, error)
	FilterByNamespace(ctx context.Context, nsID string) ([]*Token, error)
	Delete(ctx context.Context, id string) error
}
