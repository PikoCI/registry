package db

import (
	"context"
	"fmt"

	"github.com/cycloidio/sqlr"
	"github.com/pikoci/registry/pkreg/tag"
)

type TagRepository struct {
	querier sqlr.Querier
}

func NewTagRepository(db sqlr.Querier) *TagRepository {
	return &TagRepository{querier: db}
}

func (r *TagRepository) SetForRegType(ctx context.Context, typeID string, tags []string) error {
	// Delete existing tags
	_, err := r.querier.ExecContext(ctx, `DELETE FROM registry_tags WHERE type_id = ?`, typeID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Insert new tags
	for _, t := range tags {
		_, err := r.querier.ExecContext(ctx, `
			INSERT INTO registry_tags(type_id, tag) VALUES (?, ?)
		`, typeID, t)
		if err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return fmt.Errorf("failed to insert tag: %w", err)
		}
	}
	return nil
}

func (r *TagRepository) FilterByRegType(ctx context.Context, typeID string) ([]string, error) {
	rows, err := r.querier.QueryContext(ctx, `
		SELECT tag FROM registry_tags WHERE type_id = ? ORDER BY tag
	`, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (r *TagRepository) ListWithCounts(ctx context.Context) ([]*tag.TagCount, error) {
	rows, err := r.querier.QueryContext(ctx, `
		SELECT tag, COUNT(*) as cnt FROM registry_tags GROUP BY tag ORDER BY cnt DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var result []*tag.TagCount
	for rows.Next() {
		var tc tag.TagCount
		if err := rows.Scan(&tc.Tag, &tc.Count); err != nil {
			return nil, err
		}
		result = append(result, &tc)
	}
	return result, rows.Err()
}

func (r *TagRepository) ListWithCountsFiltered(ctx context.Context, kind, query string, limit int) ([]*tag.TagCount, error) {
	q := `SELECT t.tag, COUNT(*) as cnt
		FROM registry_tags t
		JOIN registry_types rt ON t.type_id = rt.id
		WHERE 1=1`
	var args []interface{}

	if kind != "" {
		q += ` AND rt.kind = ?`
		args = append(args, kind)
	}
	if query != "" {
		q += ` AND (rt.name LIKE ? OR rt.description LIKE ?)`
		like := "%" + query + "%"
		args = append(args, like, like)
	}

	q += ` GROUP BY t.tag ORDER BY cnt DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.querier.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query filtered tags: %w", err)
	}
	defer rows.Close()

	var result []*tag.TagCount
	for rows.Next() {
		var tc tag.TagCount
		if err := rows.Scan(&tc.Tag, &tc.Count); err != nil {
			return nil, err
		}
		result = append(result, &tc)
	}
	return result, rows.Err()
}

func (r *TagRepository) FilterRegTypesByTag(ctx context.Context, tagName string) ([]string, error) {
	rows, err := r.querier.QueryContext(ctx, `
		SELECT type_id FROM registry_tags WHERE tag = ?
	`, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}
