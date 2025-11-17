package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/user"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserPGRepository is a PostgreSQL implementation of the UserRepository interface.
type UserPGRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository NewRepository creates a new instance of UserPGRepository.
func NewUserRepository(p *pgxpool.Pool) ports.UserRepository {
	return &UserPGRepository{pool: p}
}

func (r *UserPGRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	const q = `SELECT 
    id,
    phone_hash,
    display_name,
    avatar_url,
    status_message,
    safety_fingerprint,
    created_at,
    updated_at,
    is_active
FROM users 
WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	var u user.User
	if err := row.Scan(
		&u.ID,
		&u.PhoneHash,
		&u.DisplayName,
		&u.AvatarURL,
		&u.StatusMessage,
		&u.SafetyFingerprint,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.IsActive,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserPGRepository) GetByPhoneHash(ctx context.Context, hash string) (*user.User, error) {
	const q = `SELECT 
    id,
    phone_hash,
    display_name,
    avatar_url,
    status_message,
    safety_fingerprint,
    created_at,
    updated_at,
    is_active
FROM users 
WHERE phone_hash = $1`
	row := r.pool.QueryRow(ctx, q, hash)
	var u user.User
	if err := row.Scan(
		&u.ID,
		&u.PhoneHash,
		&u.DisplayName,
		&u.AvatarURL,
		&u.StatusMessage,
		&u.SafetyFingerprint,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.IsActive,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserPGRepository) Create(ctx context.Context, u *user.User) error {
	const q = `INSERT INTO users (
                   id, 
                   phone_hash, 
                   display_name,
                   avatar_url,
                   status_message,
                   safety_fingerprint,
                   created_at, 
                   updated_at, 
                   is_active
                   ) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(ctx, q, u.ID, u.PhoneHash, u.DisplayName, u.AvatarURL, u.StatusMessage, u.SafetyFingerprint, u.CreatedAt, u.UpdatedAt, u.IsActive)

	return err
}

func (r *UserPGRepository) UpdateProfile(ctx context.Context, u *user.User) error {
	if u.ID == uuid.Nil {
		return errors.New("empty user ID")
	}

	const q = `UPDATE users SET
display_name = $2,
avatar_url = $3,
status_message = $4,
safety_fingerprint = $5,
updated_at = $6
WHERE id = $1`

	_, err := r.pool.Exec(ctx, q, u.ID, u.DisplayName, u.AvatarURL, u.StatusMessage, u.SafetyFingerprint, u.UpdatedAt)
	return err
}
