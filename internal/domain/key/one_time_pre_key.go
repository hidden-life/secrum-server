package key

import (
	"time"

	"github.com/google/uuid"
)

type OneTimePreKey struct {
	ID        uuid.UUID
	DeviceID  uuid.UUID
	PublicKey string
	CreatedAt time.Time
	UsedAt    *time.Time
}

func New(deviceID uuid.UUID, pub string) *OneTimePreKey {
	return &OneTimePreKey{
		ID:        uuid.New(),
		DeviceID:  deviceID,
		PublicKey: pub,
		CreatedAt: time.Now().UTC(),
	}
}
