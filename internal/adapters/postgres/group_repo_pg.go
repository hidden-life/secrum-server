package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/group"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewGroupRepository(p *pgxpool.Pool) *GroupRepositoryPG {
	return &GroupRepositoryPG{pool: p}
}

func (r *GroupRepositoryPG) Create(ctx context.Context, g *group.Group) error {
	const q = `INSERT INTO groups(
                   id, 
                   name, 
                   avatar_url, 
                   created_by, 
                   created_at, 
                   updated_at, 
                   is_active, 
                   allowed_mime_types
                   ) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, q, g.ID, g.Name, g.AvatarURL, g.CreatedBy, g.CreatedAt, g.UpdatedAt, g.IsActive, g.AllowedMimeTypes)

	return err
}

func (r *GroupRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*group.Group, error) {
	const q = `SELECT id, name, avatar_url, created_by, created_at, updated_at, is_active, allowed_mime_types FROM groups WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	g := group.Group{}
	if err := row.Scan(&g.ID, &g.Name, &g.AvatarURL, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt, &g.IsActive, &g.AllowedMimeTypes); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &g, nil
}

func (r *GroupRepositoryPG) ListByUser(ctx context.Context, userID uuid.UUID) ([]*group.Group, error) {
	const q = `SELECT g.id, g.name, g.avatar_url, g.created_by, g.created_at, g.updated_at, g.is_active, g.allowed_mime_types 
		FROM groups AS g
		JOIN group_members AS gm ON g.id = gm.group_id
		WHERE gm.user_id = $1 AND gm.is_active = TRUE AND g.is_active = TRUE ORDER BY g.created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*group.Group
	for rows.Next() {
		var g group.Group
		if err := rows.Scan(&g.ID, &g.Name, &g.AvatarURL, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt, &g.IsActive, &g.AllowedMimeTypes); err != nil {
			return nil, err
		}
		result = append(result, &g)
	}

	return result, rows.Err()
}

func (r *GroupRepositoryPG) Update(ctx context.Context, g *group.Group) error {
	const q = `UPDATE groups 
		SET name = $2,
		avatar_url = $3,
		updated_at = $4,
		is_active = $5
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, g.ID, g.Name, g.AvatarURL, g.UpdatedAt, g.IsActive)
	return err
}

func (r *GroupRepositoryPG) UpdateAllowedMimeTypes(ctx context.Context, groupID uuid.UUID, mimeTypes []string) error {
	const q = `UPDATE groups SET allowed_mime_types = $2, updated_at = NOW() AT TIME ZONE 'UTC' WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, groupID, mimeTypes)
	return err
}

func (r *GroupRepositoryPG) GetAllowedMimeTypes(ctx context.Context, groupID uuid.UUID) ([]string, error) {
	const q = `SELECT allowed_mime_types FROM groups WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, groupID)
	var mimeTypes []string
	if err := row.Scan(&mimeTypes); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return mimeTypes, nil
}
