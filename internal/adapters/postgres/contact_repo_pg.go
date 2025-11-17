package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/contact"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContactRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewContactRepository(p *pgxpool.Pool) *ContactRepositoryPG {
	return &ContactRepositoryPG{pool: p}
}

func (r *ContactRepositoryPG) Add(ctx context.Context, c *contact.Contact) error {
	const q = `
		INSERT INTO contacts(id, owner_user_id, contact_user_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (owner_user_id, contact_user_id) DO NOTHING`

	_, err := r.pool.Exec(ctx, q, c.ID, c.OwnerUserID, c.ContactUserID, c.CreatedAt)
	return err
}

func (r *ContactRepositoryPG) Remove(ctx context.Context, owner, target uuid.UUID) error {
	const q = `DELETE FROM contacts WHERE owner_user_id = $1 AND contact_user_id = $2`
	_, err := r.pool.Exec(ctx, q, owner, target)
	return err
}

func (r *ContactRepositoryPG) List(ctx context.Context, owner uuid.UUID) ([]*contact.Contact, error) {
	const q = `SELECT 
		id, 
		owner_user_id, 
		contact_user_id, 
		created_at 
	FROM contacts WHERE owner_user_id = $1
	ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, owner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*contact.Contact
	for rows.Next() {
		var c contact.Contact
		if err := rows.Scan(&c.ID, &c.OwnerUserID, &c.ContactUserID, &c.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, &c)
	}

	return result, nil
}

func (r *ContactRepositoryPG) Exists(ctx context.Context, owner uuid.UUID, target uuid.UUID) (bool, error) {
	const q = `SELECT 1 FROM contacts WHERE owner_user_id = $1 AND contact_user_id = $2`
	row := r.pool.QueryRow(ctx, q, owner, target)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return true, nil
}
