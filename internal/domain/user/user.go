package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID uuid.UUID

	Username *string

	DisplayName       *string
	AvatarURL         *string
	StatusMessage     *string
	SafetyFingerprint *string

	PhoneHash string
	CreatedAt time.Time
	UpdatedAt time.Time

	IsActive bool

	AllowedMimeTypes []string // user-level mime-types allowed
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

func (u *User) UpdateProfile(displayName, avatarURL, statusMessage *string) {
	u.DisplayName = displayName
	u.AvatarURL = avatarURL
	u.StatusMessage = statusMessage
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) SetSafetyFingerprint(fp string) {
	u.SafetyFingerprint = &fp
	u.UpdatedAt = time.Now().UTC()
}
