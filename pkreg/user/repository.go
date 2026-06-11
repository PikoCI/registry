package user

import "context"

//go:generate go tool mockgen -destination=../mock/user_repository.go -mock_names=Repository=UserRepository -package mock github.com/pikoci/registry/pkreg/user Repository

type Repository interface {
	Create(ctx context.Context, u User) (string, error)
	FindByID(ctx context.Context, id string) (*User, error)
	FindByGitHubID(ctx context.Context, ghID string) (*User, error)
	Update(ctx context.Context, u User) error
}
