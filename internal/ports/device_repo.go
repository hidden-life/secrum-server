package ports

import (
	"context"

	"github.com/hidden-life/secrum-server/internal/domain/device"
)

type DeviceRepository interface {
	Create(context.Context, *device.Device) error
}
