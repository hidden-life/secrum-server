package keybundle

import (
	"time"

	"github.com/google/uuid"
)

type KeyBundle struct {
	ID             uuid.UUID
	DeviceID       uuid.UUID
	IdentityKey    string
	SignedPreKey   string
	OneTimePreKeys []string
	CreatedAt      time.Time
}

func New(deviceID uuid.UUID, identityKey, signedPreKey string, oneTimePreKeys []string) *KeyBundle {
	return &KeyBundle{
		ID:             uuid.New(),
		DeviceID:       deviceID,
		IdentityKey:    identityKey,
		SignedPreKey:   signedPreKey,
		OneTimePreKeys: oneTimePreKeys,
		CreatedAt:      time.Now().UTC(),
	}
}
