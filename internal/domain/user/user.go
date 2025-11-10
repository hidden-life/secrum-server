package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID

	PhoneHash string
	CreatedAt time.Time
	UpdatedAt time.Time

	IsActive bool
}

func New(hash string) *User {
	now := time.Now().UTC()
	return &User{
		ID: uuid.New(),

		PhoneHash: hash,
		CreatedAt: now,
		UpdatedAt: now,

		IsActive: true,
	}
}
