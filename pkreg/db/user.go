package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cycloidio/sqlr"
	"github.com/pikoci/registry/pkreg/user"
)

type UserRepository struct {
	querier sqlr.Querier
}

func NewUserRepository(db sqlr.Querier) *UserRepository {
	return &UserRepository{querier: db}
}

type dbUser struct {
	ID        sql.NullString
	GitHubID  sql.NullString
	Username  sql.NullString
	AvatarURL sql.NullString
	CreatedAt sql.NullString
}

func newDBUser(u user.User) dbUser {
	return dbUser{
		ID:        toNullString(u.ID),
		GitHubID:  toNullString(u.GitHubID),
		Username:  toNullString(u.Username),
		AvatarURL: toNullString(u.AvatarURL),
		CreatedAt: timeToString(u.CreatedAt),
	}
}

func (dbu *dbUser) toDomainEntity() *user.User {
	return &user.User{
		ID:        dbu.ID.String,
		GitHubID:  dbu.GitHubID.String,
		Username:  dbu.Username.String,
		AvatarURL: dbu.AvatarURL.String,
		CreatedAt: stringToTime(dbu.CreatedAt),
	}
}

func (r *UserRepository) Create(ctx context.Context, u user.User) (string, error) {
	dbu := newDBUser(u)
	_, err := r.querier.ExecContext(ctx, `
		INSERT INTO users(id, github_id, username, avatar_url, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, dbu.ID, dbu.GitHubID, dbu.Username, dbu.AvatarURL, dbu.CreatedAt)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return u.ID, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	var dbu dbUser
	err := r.querier.QueryRowContext(ctx, `
		SELECT id, github_id, username, avatar_url, created_at
		FROM users WHERE id = ?
	`, id).Scan(&dbu.ID, &dbu.GitHubID, &dbu.Username, &dbu.AvatarURL, &dbu.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return dbu.toDomainEntity(), nil
}

func (r *UserRepository) FindByGitHubID(ctx context.Context, ghID string) (*user.User, error) {
	var dbu dbUser
	err := r.querier.QueryRowContext(ctx, `
		SELECT id, github_id, username, avatar_url, created_at
		FROM users WHERE github_id = ?
	`, ghID).Scan(&dbu.ID, &dbu.GitHubID, &dbu.Username, &dbu.AvatarURL, &dbu.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return dbu.toDomainEntity(), nil
}

func (r *UserRepository) Update(ctx context.Context, u user.User) error {
	dbu := newDBUser(u)
	res, err := r.querier.ExecContext(ctx, `
		UPDATE users SET github_id = ?, username = ?, avatar_url = ?
		WHERE id = ?
	`, dbu.GitHubID, dbu.Username, dbu.AvatarURL, dbu.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return isEntityFound(res)
}
