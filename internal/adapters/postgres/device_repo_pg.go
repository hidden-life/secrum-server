package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/device"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceRepositoryPG struct {
	pool *pgxpool.Pool
}

// NewDeviceRepository creates a new instance of DeviceRepositoryPG.
func NewDeviceRepository(p *pgxpool.Pool) *DeviceRepositoryPG {
	return &DeviceRepositoryPG{pool: p}
}

// Create inserts a new device record into the database.
func (r *DeviceRepositoryPG) Create(ctx context.Context, d *device.Device) error {
	const q = `
INSERT INTO devices (id, user_id, name, platform, created_at, last_seen, refresh_token_hash, refresh_token_expires_at, is_active, identity_key, signed_prekey, signed_prekey_signature) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.pool.Exec(
		ctx,
		q,
		d.ID,
		d.UserID,
		d.Name,
		d.Platform,
		d.CreatedAt,
		d.LastSeen,
		d.RefreshTokenHash,
		d.RefreshTokenExpiresAt,
		d.IsActive,
		d.IdentityKey,
		d.SignedPreKey,
		d.SignedPreKeySignature,
	)

	return err
}

func (r *DeviceRepositoryPG) GetById(ctx context.Context, deviceID uuid.UUID) (*device.Device, error) {
	const q = `SELECT 
		id, 
		user_id, 
		name, 
		platform, 
		created_at, 
		last_seen, 
		refresh_token_hash, 
		refresh_token_expires_at, 
		is_active,
		identity_key,
		signed_prekey,
		signed_prekey_signature
	FROM devices WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, deviceID)
	var d device.Device

	if err := row.Scan(&d.ID, &d.UserID, &d.Name, &d.Platform, &d.CreatedAt, &d.LastSeen, &d.RefreshTokenHash, &d.RefreshTokenExpiresAt, &d.IsActive, &d.IdentityKey, &d.SignedPreKey, &d.SignedPreKeySignature); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &d, nil
}

func (r *DeviceRepositoryPG) UpdateRefreshToken(ctx context.Context, id uuid.UUID, hash string, expiresAt *time.Time) error {
	const q = `UPDATE devices SET refresh_token_hash = $2, refresh_token_expires_at = $3 WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, hash, expiresAt)

	return err
}

func (r *DeviceRepositoryPG) ClearRefreshToken(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE devices SET refresh_token_hash = NULL, refresh_token_expires_at = NULL WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)

	return err
}

func (r *DeviceRepositoryPG) UpdateLastSeen(ctx context.Context, id uuid.UUID, lastSeen time.Time) error {
	const q = `UPDATE devices SET last_seen = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, lastSeen)

	return err
}

func (r *DeviceRepositoryPG) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*device.Device, error) {
	const q = `SELECT
		id,
		user_id,
		name,
		platform,
		created_at,
		last_seen,
		is_active,
		refresh_token_hash,
		refresh_token_expires_at,
		identity_key,
		signed_prekey,
		signed_prekey_signature
	FROM devices
	WHERE user_id = $1 AND is_active = TRUE`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*device.Device
	for rows.Next() {
		var d device.Device
		if err := rows.Scan(
			&d.ID,
			&d.UserID,
			&d.Name,
			&d.Platform,
			&d.CreatedAt,
			&d.LastSeen,
			&d.IsActive,
			&d.RefreshTokenHash,
			&d.RefreshTokenExpiresAt,
			&d.IdentityKey,
			&d.SignedPreKey,
			&d.SignedPreKeySignature,
		); err != nil {
			return nil, err
		}
		res = append(res, &d)
	}

	return res, nil
}

func (r *DeviceRepositoryPG) Deactivate(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE devices SET is_active = FALSE, refresh_token_hash = NULL, refresh_token_expires_at = NULL WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}

func (r *DeviceRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE devices SET is_active = FALSE WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id)
	return err
}

func (r *DeviceRepositoryPG) UpdateKeys(ctx context.Context, id uuid.UUID, identityKey, signedPreKey, signedPreKeySignature string) error {
	const q = `UPDATE devices SET identity_key = $2, signed_prekey = $3, signed_prekey_signature = $4 WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, identityKey, signedPreKey, signedPreKeySignature)

	return err
}
