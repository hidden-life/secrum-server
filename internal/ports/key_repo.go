package ports

import (
	"context"

	"github.com/google/uuid"
	keybundle "github.com/hidden-life/secrum-server/internal/domain/key_bundle"
)

type KeyRepository interface {
	Save(context.Context, *keybundle.KeyBundle) error
	GetByDeviceID(context.Context, uuid.UUID) (*keybundle.KeyBundle, error)
}
