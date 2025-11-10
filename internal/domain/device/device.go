package device

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Platform  string
	CreatedAt time.Time
	LastSeen  time.Time

	IsActive bool
}

func New(userID uuid.UUID, name, platform string) *Device {
	now := time.Now().UTC()

	return &Device{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Platform:  platform,
		CreatedAt: now,
		LastSeen:  now,

		IsActive: true,
	}
}
