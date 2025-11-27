package ports

import (
	"context"
)

type SessionStore interface {
	IsDeviceRevoked(context.Context, string) (bool, error)
	RevokeDevice(context.Context, string) error
}
