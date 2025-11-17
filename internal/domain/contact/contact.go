package contact

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID            uuid.UUID
	OwnerUserID   uuid.UUID
	ContactUserID uuid.UUID
	CreatedAt     time.Time
}

func New(owner, target uuid.UUID) *Contact {
	return &Contact{
		ID:            uuid.New(),
		OwnerUserID:   owner,
		ContactUserID: target,
		CreatedAt:     time.Now(),
	}
}
