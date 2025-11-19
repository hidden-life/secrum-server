package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/group"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupMemberRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewGroupMemberRepository(p *pgxpool.Pool) *GroupMemberRepositoryPG {
	return &GroupMemberRepositoryPG{pool: p}
}

func (r *GroupMemberRepositoryPG) AddMember(ctx context.Context, m *group.Member) error {
	const q = `INSERT INTO group_members (group_id, user_id, role, joined_at, is_active) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (group_id, user_id) DO UPDATE
		SET role = EXCLUDED.role,
		joined_at = EXCLUDED.joined_at,
		is_active = EXCLUDED.is_active;`

	_, err := r.pool.Exec(ctx, q, m.GroupID, m.UserID, m.Role, m.JoinedAt, m.IsActive)
	return err
}

func (r *GroupMemberRepositoryPG) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	const q = `UPDATE group_members SET is_active = false WHERE group_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, q, true, groupID, userID)
	return err
}

func (r *GroupMemberRepositoryPG) List(ctx context.Context, groupID uuid.UUID) ([]*group.Member, error) {
	const q = `SELECT group_id, user_id, role, joined_at, is_active 
		FROM group_members
		WHERE group_id = $1
		ORDER BY joined_at DESC`

	rows, err := r.pool.Query(ctx, q, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*group.Member
	for rows.Next() {
		var m group.Member
		if err := rows.Scan(
			&m.GroupID,
			&m.UserID,
			&m.Role,
			&m.JoinedAt,
			&m.IsActive,
		); err != nil {
			return nil, err
		}
		res = append(res, &m)
	}

	return res, rows.Err()
}

func (r *GroupMemberRepositoryPG) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	const q = `SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2 AND is_active = TRUE`
	row := r.pool.QueryRow(ctx, q, groupID, userID)
	var d int

	if err := row.Scan(&d); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *GroupMemberRepositoryPG) GetRole(ctx context.Context, groupID, userID uuid.UUID) (group.MemberRole, error) {
	const q = `SELECT role FROM group_members WHERE group_id = $1 AND user_id = $2 AND is_active = TRUE`
	row := r.pool.QueryRow(ctx, q, groupID, userID)
	var role group.MemberRole
	if err := row.Scan(&role); err != nil {
		return "", err
	}

	return role, nil
}
