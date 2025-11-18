package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/device"
)

type DeviceRepository interface {
	Create(context.Context, *device.Device) error
	GetById(context.Context, uuid.UUID) (*device.Device, error)

	UpdateRefreshToken(context.Context, uuid.UUID, string, *time.Time) error
	ClearRefreshToken(context.Context, uuid.UUID) error

	UpdateLastSeen(context.Context, uuid.UUID, time.Time) error

	ListActiveByUser(context.Context, uuid.UUID) ([]*device.Device, error)
	Deactivate(context.Context, uuid.UUID) error
	Delete(context.Context, uuid.UUID) error
}
