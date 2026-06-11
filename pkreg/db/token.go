package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cycloidio/sqlr"
	"github.com/pikoci/registry/pkreg/token"
)

type TokenRepository struct {
	querier sqlr.Querier
}

func NewTokenRepository(db sqlr.Querier) *TokenRepository {
	return &TokenRepository{querier: db}
}

type dbToken struct {
	ID          sql.NullString
	NamespaceID sql.NullString
	Name        sql.NullString
	TokenHash   sql.NullString
	Prefix      sql.NullString
	CreatedAt   sql.NullString
	ExpiresAt   sql.NullString
}

func (dbt *dbToken) toDomainEntity() *token.Token {
	t := &token.Token{
		ID:          dbt.ID.String,
		NamespaceID: dbt.NamespaceID.String,
		Name:        dbt.Name.String,
		TokenHash:   dbt.TokenHash.String,
		Prefix:      dbt.Prefix.String,
		CreatedAt:   stringToTime(dbt.CreatedAt),
	}
	if dbt.ExpiresAt.Valid && dbt.ExpiresAt.String != "" {
		parsed := stringToTime(dbt.ExpiresAt)
		if !parsed.IsZero() {
			t.ExpiresAt = &parsed
		}
	}
	return t
}

func (r *TokenRepository) Create(ctx context.Context, t token.Token) (string, error) {
	var expiresAt sql.NullString
	if t.ExpiresAt != nil {
		expiresAt = timeToString(*t.ExpiresAt)
	}
	_, err := r.querier.ExecContext(ctx, `
		INSERT INTO api_tokens(id, namespace_id, name, token_hash, prefix, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, t.ID, t.NamespaceID, t.Name, t.TokenHash, t.Prefix, timeToString(t.CreatedAt), expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}
	return t.ID, nil
}

func (r *TokenRepository) FindByHash(ctx context.Context, hash string) (*token.Token, error) {
	var dbt dbToken
	err := r.querier.QueryRowContext(ctx, `
		SELECT id, namespace_id, name, token_hash, prefix, created_at, expires_at
		FROM api_tokens WHERE token_hash = ?
	`, hash).Scan(&dbt.ID, &dbt.NamespaceID, &dbt.Name, &dbt.TokenHash, &dbt.Prefix, &dbt.CreatedAt, &dbt.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("token not found: %w", err)
	}
	return dbt.toDomainEntity(), nil
}

func (r *TokenRepository) FilterByNamespace(ctx context.Context, nsID string) ([]*token.Token, error) {
	rows, err := r.querier.QueryContext(ctx, `
		SELECT id, namespace_id, name, token_hash, prefix, created_at, expires_at
		FROM api_tokens WHERE namespace_id = ? ORDER BY created_at DESC
	`, nsID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tokens: %w", err)
	}
	defer rows.Close()

	var result []*token.Token
	for rows.Next() {
		var dbt dbToken
		if err := rows.Scan(&dbt.ID, &dbt.NamespaceID, &dbt.Name, &dbt.TokenHash, &dbt.Prefix, &dbt.CreatedAt, &dbt.ExpiresAt); err != nil {
			return nil, err
		}
		result = append(result, dbt.toDomainEntity())
	}
	return result, rows.Err()
}

func (r *TokenRepository) Delete(ctx context.Context, id string) error {
	res, err := r.querier.ExecContext(ctx, `DELETE FROM api_tokens WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return isEntityFound(res)
}
