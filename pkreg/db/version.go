package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cycloidio/sqlr"
	"github.com/pikoci/registry/pkreg/version"
)

type VersionRepository struct {
	querier sqlr.Querier
}

func NewVersionRepository(db sqlr.Querier) *VersionRepository {
	return &VersionRepository{querier: db}
}

type dbVersion struct {
	ID                sql.NullString
	TypeID            sql.NullString
	Version           sql.NullString
	Content           sql.NullString
	Readme            sql.NullString
	Params            sql.NullString
	Examples          sql.NullString
	Yanked            sql.NullBool
	Deprecated        sql.NullBool
	DeprecatedMessage sql.NullString
	SuccessorVersion  sql.NullString
	Downloads         sql.NullInt64
	CreatedAt         sql.NullString
	CreatedBy         sql.NullString
}

func newDBVersion(v version.Version) dbVersion {
	return dbVersion{
		ID:                toNullString(v.ID),
		TypeID:            toNullString(v.TypeID),
		Version:           toNullString(v.Version),
		Content:           toNullString(v.Content),
		Readme:            toNullString(v.Readme),
		Params:            toNullString(string(v.Params)),
		Examples:          toNullString(string(v.Examples)),
		Yanked:            toNullBool(v.Yanked),
		Deprecated:        toNullBool(v.Deprecated),
		DeprecatedMessage: toNullString(v.DeprecatedMessage),
		SuccessorVersion:  toNullString(v.SuccessorVersion),
		Downloads:         toNullInt64(v.Downloads),
		CreatedAt:         timeToString(v.CreatedAt),
		CreatedBy:         toNullString(v.CreatedBy),
	}
}

func (dbv *dbVersion) toDomainEntity() *version.Version {
	return &version.Version{
		ID:                dbv.ID.String,
		TypeID:            dbv.TypeID.String,
		Version:           dbv.Version.String,
		Content:           dbv.Content.String,
		Readme:            dbv.Readme.String,
		Params:            json.RawMessage(dbv.Params.String),
		Examples:          json.RawMessage(dbv.Examples.String),
		Yanked:            dbv.Yanked.Bool,
		Deprecated:        dbv.Deprecated.Bool,
		DeprecatedMessage: dbv.DeprecatedMessage.String,
		SuccessorVersion:  dbv.SuccessorVersion.String,
		Downloads:         int(dbv.Downloads.Int64),
		CreatedAt:         stringToTime(dbv.CreatedAt),
		CreatedBy:         dbv.CreatedBy.String,
	}
}

const versionColumns = `id, type_id, version, content, readme, params, examples, yanked, deprecated, deprecated_message, successor_version, downloads, created_at, created_by`

func scanVersion(row interface{ Scan(...interface{}) error }) (*dbVersion, error) {
	var dbv dbVersion
	err := row.Scan(&dbv.ID, &dbv.TypeID, &dbv.Version, &dbv.Content, &dbv.Readme,
		&dbv.Params, &dbv.Examples, &dbv.Yanked, &dbv.Deprecated, &dbv.DeprecatedMessage,
		&dbv.SuccessorVersion, &dbv.Downloads, &dbv.CreatedAt, &dbv.CreatedBy)
	if err != nil {
		return nil, err
	}
	return &dbv, nil
}

func (r *VersionRepository) Create(ctx context.Context, v version.Version) (string, error) {
	dbv := newDBVersion(v)
	_, err := r.querier.ExecContext(ctx, `
		INSERT INTO registry_versions(id, type_id, version, content, readme, params, examples, yanked, deprecated, deprecated_message, successor_version, downloads, created_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, dbv.ID, dbv.TypeID, dbv.Version, dbv.Content, dbv.Readme,
		dbv.Params, dbv.Examples, dbv.Yanked, dbv.Deprecated, dbv.DeprecatedMessage,
		dbv.SuccessorVersion, dbv.Downloads, dbv.CreatedAt, dbv.CreatedBy)
	if err != nil {
		return "", fmt.Errorf("failed to create version: %w", err)
	}
	return v.ID, nil
}

func (r *VersionRepository) FindByID(ctx context.Context, id string) (*version.Version, error) {
	dbv, err := scanVersion(r.querier.QueryRowContext(ctx,
		`SELECT `+versionColumns+` FROM registry_versions WHERE id = ?`, id))
	if err != nil {
		return nil, fmt.Errorf("version not found: %w", err)
	}
	return dbv.toDomainEntity(), nil
}

func (r *VersionRepository) FindByRegTypeAndNumber(ctx context.Context, typeID, ver string) (*version.Version, error) {
	dbv, err := scanVersion(r.querier.QueryRowContext(ctx,
		`SELECT `+versionColumns+` FROM registry_versions WHERE type_id = ? AND version = ?`, typeID, ver))
	if err != nil {
		return nil, fmt.Errorf("version not found: %w", err)
	}
	return dbv.toDomainEntity(), nil
}

func (r *VersionRepository) FilterByRegType(ctx context.Context, typeID string) ([]*version.Version, error) {
	rows, err := r.querier.QueryContext(ctx,
		`SELECT `+versionColumns+` FROM registry_versions WHERE type_id = ? ORDER BY created_at DESC`, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var result []*version.Version
	for rows.Next() {
		dbv, err := scanVersion(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, dbv.toDomainEntity())
	}
	return result, rows.Err()
}

func (r *VersionRepository) Update(ctx context.Context, v version.Version) error {
	dbv := newDBVersion(v)
	res, err := r.querier.ExecContext(ctx, `
		UPDATE registry_versions SET content = ?, readme = ?, params = ?, examples = ?,
		yanked = ?, deprecated = ?, deprecated_message = ?, successor_version = ?, downloads = ?
		WHERE id = ?
	`, dbv.Content, dbv.Readme, dbv.Params, dbv.Examples,
		dbv.Yanked, dbv.Deprecated, dbv.DeprecatedMessage, dbv.SuccessorVersion, dbv.Downloads, dbv.ID)
	if err != nil {
		return fmt.Errorf("failed to update version: %w", err)
	}
	return isEntityFound(res)
}

func (r *VersionRepository) Delete(ctx context.Context, id string) error {
	res, err := r.querier.ExecContext(ctx, `DELETE FROM registry_versions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}
	return isEntityFound(res)
}

func (r *VersionRepository) IncrementDownloads(ctx context.Context, id string) error {
	res, err := r.querier.ExecContext(ctx, `
		UPDATE registry_versions SET downloads = downloads + 1 WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("failed to increment downloads: %w", err)
	}
	return isEntityFound(res)
}
