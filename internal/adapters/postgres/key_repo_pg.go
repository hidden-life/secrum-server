package postgres

import (
	"context"

	"github.com/google/uuid"
	keybundle "github.com/hidden-life/secrum-server/internal/domain/key_bundle"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KeyRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewKeyRepository(pool *pgxpool.Pool) ports.KeyRepository {
	return &KeyRepositoryPG{pool: pool}
}

func (k *KeyRepositoryPG) Save(ctx context.Context, kb *keybundle.KeyBundle) error {
	const q = `
	INSERT INTO key_bundles (id, device_id, identity_key, signed_prekey, one_time_prekeys, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (device_id) DO UPDATE SET
	identity_key = EXCLUDED.identity_key,
	signed_prekey = EXCLUDED.signed_prekey,
	one_time_prekeys = EXCLUDED.one_time_prekeys,
	created_at = EXCLUDED.created_at`

	_, err := k.pool.Exec(ctx, q, kb.ID, kb.DeviceID, kb.IdentityKey, kb.SignedPreKey, kb.OneTimePreKeys, kb.CreatedAt)

	return err
}

func (k *KeyRepositoryPG) GetByDeviceID(ctx context.Context, deviceID uuid.UUID) (*keybundle.KeyBundle, error) {
	const q = `SELECT * FROM key_bundles WHERE device_id = $1`
	row := k.pool.QueryRow(ctx, q, deviceID)
	var kb keybundle.KeyBundle
	if err := row.Scan(&kb.ID, &kb.DeviceID, &kb.IdentityKey, &kb.SignedPreKey, &kb.OneTimePreKeys, &kb.CreatedAt); err != nil {
		return nil, err
	}

	return &kb, nil
}
