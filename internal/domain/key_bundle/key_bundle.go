package keybundle

import (
	"time"

	"github.com/google/uuid"
)

type KeyBundle struct {
	ID                    uuid.UUID
	DeviceID              uuid.UUID
	IdentityKey           string
	SignedPreKey          string
	SignedPreKeySignature string
	OneTimePreKeys        []string
	CreatedAt             time.Time
}

func New(deviceID uuid.UUID, identityKey, signedPreKey, signedPreKeySig string, oneTimePreKeys []string) *KeyBundle {
	return &KeyBundle{
		ID:                    uuid.New(),
		DeviceID:              deviceID,
		IdentityKey:           identityKey,
		SignedPreKey:          signedPreKey,
		SignedPreKeySignature: signedPreKeySig,
		OneTimePreKeys:        oneTimePreKeys,
		CreatedAt:             time.Now().UTC(),
	}
}
