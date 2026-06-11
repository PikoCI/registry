package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cycloidio/sqlr"
	"github.com/pikoci/registry/pkreg/orgmember"
)

type OrgMemberRepository struct {
	querier sqlr.Querier
}

func NewOrgMemberRepository(db sqlr.Querier) *OrgMemberRepository {
	return &OrgMemberRepository{querier: db}
}

type dbOrgMember struct {
	OrgID     sql.NullString
	UserID    sql.NullString
	Role      sql.NullString
	CreatedAt sql.NullString
}

func (dbom *dbOrgMember) toDomainEntity() *orgmember.OrgMember {
	return &orgmember.OrgMember{
		OrgID:     dbom.OrgID.String,
		UserID:    dbom.UserID.String,
		Role:      dbom.Role.String,
		CreatedAt: stringToTime(dbom.CreatedAt),
	}
}

func (r *OrgMemberRepository) Create(ctx context.Context, om orgmember.OrgMember) error {
	_, err := r.querier.ExecContext(ctx, `
		INSERT INTO org_members(org_id, user_id, role, created_at)
		VALUES (?, ?, ?, ?)
	`, om.OrgID, om.UserID, om.Role, timeToString(om.CreatedAt))
	if err != nil {
		return fmt.Errorf("failed to create org member: %w", err)
	}
	return nil
}

func (r *OrgMemberRepository) FindByOrgAndUser(ctx context.Context, orgID, userID string) (*orgmember.OrgMember, error) {
	var dbom dbOrgMember
	err := r.querier.QueryRowContext(ctx, `
		SELECT org_id, user_id, role, created_at
		FROM org_members WHERE org_id = ? AND user_id = ?
	`, orgID, userID).Scan(&dbom.OrgID, &dbom.UserID, &dbom.Role, &dbom.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("org member not found: %w", err)
	}
	return dbom.toDomainEntity(), nil
}

func (r *OrgMemberRepository) FilterByOrg(ctx context.Context, orgID string) ([]*orgmember.OrgMember, error) {
	rows, err := r.querier.QueryContext(ctx, `
		SELECT org_id, user_id, role, created_at
		FROM org_members WHERE org_id = ?
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query org members: %w", err)
	}
	defer rows.Close()

	var result []*orgmember.OrgMember
	for rows.Next() {
		var dbom dbOrgMember
		if err := rows.Scan(&dbom.OrgID, &dbom.UserID, &dbom.Role, &dbom.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, dbom.toDomainEntity())
	}
	return result, rows.Err()
}

func (r *OrgMemberRepository) Delete(ctx context.Context, orgID, userID string) error {
	res, err := r.querier.ExecContext(ctx, `
		DELETE FROM org_members WHERE org_id = ? AND user_id = ?
	`, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete org member: %w", err)
	}
	return isEntityFound(res)
}
