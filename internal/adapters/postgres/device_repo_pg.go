package postgres

import (
	"context"

	"github.com/hidden-life/secrum-server/internal/domain/device"
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
	const q = `INSERT INTO devices (id, user_id, name, platform, created_at, last_seen, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, q, d.ID, d.UserID, d.Name, d.Platform, d.CreatedAt, d.LastSeen, d.IsActive)

	return err
}
