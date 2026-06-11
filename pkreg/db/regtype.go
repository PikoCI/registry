package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cycloidio/sqlr"
	"github.com/pikoci/registry/pkreg/regtype"
)

type RegTypeRepository struct {
	querier sqlr.Querier
}

func NewRegTypeRepository(db sqlr.Querier) *RegTypeRepository {
	return &RegTypeRepository{querier: db}
}

type dbRegType struct {
	ID            sql.NullString
	NamespaceID   sql.NullString
	NamespaceName sql.NullString
	Name          sql.NullString
	Kind        sql.NullString
	Description sql.NullString
	Repository  sql.NullString
	Homepage    sql.NullString
	License     sql.NullString
	Public      sql.NullBool
	CreatedAt   sql.NullString
}

func newDBRegType(rt regtype.RegType) dbRegType {
	return dbRegType{
		ID:          toNullString(rt.ID),
		NamespaceID: toNullString(rt.NamespaceID),
		Name:        toNullString(rt.Name),
		Kind:        toNullString(rt.Kind),
		Description: toNullString(rt.Description),
		Repository:  toNullString(rt.Repository),
		Homepage:    toNullString(rt.Homepage),
		License:     toNullString(rt.License),
		Public:      toNullBool(rt.Public),
		CreatedAt:   timeToString(rt.CreatedAt),
	}
}

func (dbrt *dbRegType) toDomainEntity() *regtype.RegType {
	return &regtype.RegType{
		ID:            dbrt.ID.String,
		NamespaceID:   dbrt.NamespaceID.String,
		NamespaceName: dbrt.NamespaceName.String,
		Name:          dbrt.Name.String,
		Kind:        dbrt.Kind.String,
		Description: dbrt.Description.String,
		Repository:  dbrt.Repository.String,
		Homepage:    dbrt.Homepage.String,
		License:     dbrt.License.String,
		Public:      dbrt.Public.Bool,
		CreatedAt:   stringToTime(dbrt.CreatedAt),
	}
}

func scanRegType(row interface{ Scan(...interface{}) error }) (*dbRegType, error) {
	var dbrt dbRegType
	err := row.Scan(&dbrt.ID, &dbrt.NamespaceID, &dbrt.Name, &dbrt.Kind, &dbrt.Description,
		&dbrt.Repository, &dbrt.Homepage, &dbrt.License, &dbrt.Public, &dbrt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &dbrt, nil
}

func scanRegTypeWithNS(row interface{ Scan(...interface{}) error }) (*dbRegType, error) {
	var dbrt dbRegType
	err := row.Scan(&dbrt.ID, &dbrt.NamespaceID, &dbrt.NamespaceName, &dbrt.Name, &dbrt.Kind, &dbrt.Description,
		&dbrt.Repository, &dbrt.Homepage, &dbrt.License, &dbrt.Public, &dbrt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &dbrt, nil
}

const regTypeColumns = `id, namespace_id, name, kind, description, repository, homepage, license, public, created_at`

func (r *RegTypeRepository) Create(ctx context.Context, rt regtype.RegType) (string, error) {
	dbrt := newDBRegType(rt)
	_, err := r.querier.ExecContext(ctx, `
		INSERT INTO registry_types(id, namespace_id, name, kind, description, repository, homepage, license, public, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, dbrt.ID, dbrt.NamespaceID, dbrt.Name, dbrt.Kind, dbrt.Description,
		dbrt.Repository, dbrt.Homepage, dbrt.License, dbrt.Public, dbrt.CreatedAt)
	if err != nil {
		return "", fmt.Errorf("failed to create regtype: %w", err)
	}
	return rt.ID, nil
}

func (r *RegTypeRepository) FindByID(ctx context.Context, id string) (*regtype.RegType, error) {
	dbrt, err := scanRegType(r.querier.QueryRowContext(ctx,
		`SELECT `+regTypeColumns+` FROM registry_types WHERE id = ?`, id))
	if err != nil {
		return nil, fmt.Errorf("regtype not found: %w", err)
	}
	return dbrt.toDomainEntity(), nil
}

func (r *RegTypeRepository) FindByNamespaceAndName(ctx context.Context, nsID, name, kind string) (*regtype.RegType, error) {
	var query string
	var args []interface{}
	if kind != "" {
		query = `SELECT ` + regTypeColumns + ` FROM registry_types WHERE namespace_id = ? AND name = ? AND kind = ? LIMIT 1`
		args = []interface{}{nsID, name, kind}
	} else {
		query = `SELECT ` + regTypeColumns + ` FROM registry_types WHERE namespace_id = ? AND name = ? ORDER BY created_at ASC LIMIT 1`
		args = []interface{}{nsID, name}
	}
	dbrt, err := scanRegType(r.querier.QueryRowContext(ctx, query, args...))
	if err != nil {
		return nil, fmt.Errorf("regtype not found: %w", err)
	}
	return dbrt.toDomainEntity(), nil
}

func (r *RegTypeRepository) FilterByNamespace(ctx context.Context, nsID string) ([]*regtype.RegType, error) {
	rows, err := r.querier.QueryContext(ctx,
		`SELECT `+regTypeColumns+` FROM registry_types WHERE namespace_id = ? ORDER BY name`, nsID)
	if err != nil {
		return nil, fmt.Errorf("failed to query regtypes: %w", err)
	}
	defer rows.Close()

	var result []*regtype.RegType
	for rows.Next() {
		dbrt, err := scanRegType(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, dbrt.toDomainEntity())
	}
	return result, rows.Err()
}

func (r *RegTypeRepository) Search(ctx context.Context, params regtype.SearchParams) (*regtype.SearchResult, error) {
	query := `SELECT rt.id, rt.namespace_id, n.name, rt.name, rt.kind, rt.description, rt.repository, rt.homepage, rt.license, rt.public, rt.created_at
		FROM registry_types rt
		JOIN namespaces n ON rt.namespace_id = n.id
		WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM registry_types rt JOIN namespaces n ON rt.namespace_id = n.id WHERE 1=1`

	var args []interface{}
	var countArgs []interface{}

	if params.Query != "" {
		filter := ` AND (rt.name LIKE ? OR rt.description LIKE ?)`
		like := "%" + params.Query + "%"
		query += filter
		countQuery += filter
		args = append(args, like, like)
		countArgs = append(countArgs, like, like)
	}
	if params.Kind != "" {
		filter := ` AND rt.kind = ?`
		query += filter
		countQuery += filter
		args = append(args, params.Kind)
		countArgs = append(countArgs, params.Kind)
	}
	if params.Namespace != "" {
		filter := ` AND n.name = ?`
		query += filter
		countQuery += filter
		args = append(args, params.Namespace)
		countArgs = append(countArgs, params.Namespace)
	}
	if params.Tag != "" {
		filter := ` AND rt.id IN (SELECT type_id FROM registry_tags WHERE tag = ?)`
		query += filter
		countQuery += filter
		args = append(args, params.Tag)
		countArgs = append(countArgs, params.Tag)
	}

	// Count
	var totalCount int
	err := r.querier.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Paginate
	offset := (params.Page - 1) * params.PerPage
	query += ` ORDER BY rt.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, params.PerPage, offset)

	rows, err := r.querier.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search regtypes: %w", err)
	}
	defer rows.Close()

	var types []*regtype.RegType
	for rows.Next() {
		dbrt, err := scanRegTypeWithNS(rows)
		if err != nil {
			return nil, err
		}
		types = append(types, dbrt.toDomainEntity())
	}

	return &regtype.SearchResult{
		Types:      types,
		TotalCount: totalCount,
	}, rows.Err()
}

func (r *RegTypeRepository) Update(ctx context.Context, rt regtype.RegType) error {
	dbrt := newDBRegType(rt)
	res, err := r.querier.ExecContext(ctx, `
		UPDATE registry_types SET name = ?, kind = ?, description = ?, repository = ?, homepage = ?, license = ?, public = ?
		WHERE id = ?
	`, dbrt.Name, dbrt.Kind, dbrt.Description, dbrt.Repository, dbrt.Homepage, dbrt.License, dbrt.Public, dbrt.ID)
	if err != nil {
		return fmt.Errorf("failed to update regtype: %w", err)
	}
	return isEntityFound(res)
}

func (r *RegTypeRepository) Delete(ctx context.Context, id string) error {
	res, err := r.querier.ExecContext(ctx, `DELETE FROM registry_types WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete regtype: %w", err)
	}
	return isEntityFound(res)
}
