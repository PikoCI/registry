package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cycloidio/sqlr"
	"github.com/pikoci/registry/pkreg/namespace"
)

type NamespaceRepository struct {
	querier sqlr.Querier
}

func NewNamespaceRepository(db sqlr.Querier) *NamespaceRepository {
	return &NamespaceRepository{querier: db}
}

type dbNamespace struct {
	ID        sql.NullString
	Name      sql.NullString
	Type      sql.NullString
	GitHubID  sql.NullString
	OwnerID   sql.NullString
	Public    sql.NullBool
	CreatedAt sql.NullString
}

func newDBNamespace(ns namespace.Namespace) dbNamespace {
	return dbNamespace{
		ID:        toNullString(ns.ID),
		Name:      toNullString(ns.Name),
		Type:      toNullString(ns.Type),
		GitHubID:  toNullString(ns.GitHubID),
		OwnerID:   toNullString(ns.OwnerID),
		Public:    toNullBool(ns.Public),
		CreatedAt: timeToString(ns.CreatedAt),
	}
}

func (dbns *dbNamespace) toDomainEntity() *namespace.Namespace {
	return &namespace.Namespace{
		ID:        dbns.ID.String,
		Name:      dbns.Name.String,
		Type:      dbns.Type.String,
		GitHubID:  dbns.GitHubID.String,
		OwnerID:   dbns.OwnerID.String,
		Public:    dbns.Public.Bool,
		CreatedAt: stringToTime(dbns.CreatedAt),
	}
}

func (r *NamespaceRepository) Create(ctx context.Context, ns namespace.Namespace) (string, error) {
	dbns := newDBNamespace(ns)
	_, err := r.querier.ExecContext(ctx, `
		INSERT INTO namespaces(id, name, type, github_id, owner_id, public, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, dbns.ID, dbns.Name, dbns.Type, dbns.GitHubID, dbns.OwnerID, dbns.Public, dbns.CreatedAt)
	if err != nil {
		return "", fmt.Errorf("failed to create namespace: %w", err)
	}
	return ns.ID, nil
}

func (r *NamespaceRepository) FindByID(ctx context.Context, id string) (*namespace.Namespace, error) {
	var dbns dbNamespace
	err := r.querier.QueryRowContext(ctx, `
		SELECT id, name, type, github_id, owner_id, public, created_at
		FROM namespaces WHERE id = ?
	`, id).Scan(&dbns.ID, &dbns.Name, &dbns.Type, &dbns.GitHubID, &dbns.OwnerID, &dbns.Public, &dbns.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}
	return dbns.toDomainEntity(), nil
}

func (r *NamespaceRepository) FindByName(ctx context.Context, name string) (*namespace.Namespace, error) {
	var dbns dbNamespace
	err := r.querier.QueryRowContext(ctx, `
		SELECT id, name, type, github_id, owner_id, public, created_at
		FROM namespaces WHERE name = ?
	`, name).Scan(&dbns.ID, &dbns.Name, &dbns.Type, &dbns.GitHubID, &dbns.OwnerID, &dbns.Public, &dbns.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("namespace not found: %w", err)
	}
	return dbns.toDomainEntity(), nil
}

func (r *NamespaceRepository) FindByOwnerID(ctx context.Context, ownerID string) ([]*namespace.Namespace, error) {
	rows, err := r.querier.QueryContext(ctx, `
		SELECT id, name, type, github_id, owner_id, public, created_at
		FROM namespaces WHERE owner_id = ?
	`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query namespaces: %w", err)
	}
	defer rows.Close()

	var result []*namespace.Namespace
	for rows.Next() {
		var dbns dbNamespace
		if err := rows.Scan(&dbns.ID, &dbns.Name, &dbns.Type, &dbns.GitHubID, &dbns.OwnerID, &dbns.Public, &dbns.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, dbns.toDomainEntity())
	}
	return result, rows.Err()
}

func (r *NamespaceRepository) Update(ctx context.Context, ns namespace.Namespace) error {
	dbns := newDBNamespace(ns)
	res, err := r.querier.ExecContext(ctx, `
		UPDATE namespaces SET name = ?, type = ?, github_id = ?, owner_id = ?, public = ?
		WHERE id = ?
	`, dbns.Name, dbns.Type, dbns.GitHubID, dbns.OwnerID, dbns.Public, dbns.ID)
	if err != nil {
		return fmt.Errorf("failed to update namespace: %w", err)
	}
	return isEntityFound(res)
}

func (r *NamespaceRepository) Delete(ctx context.Context, id string) error {
	res, err := r.querier.ExecContext(ctx, `DELETE FROM namespaces WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}
	return isEntityFound(res)
}
