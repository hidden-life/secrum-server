package message

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID                uuid.UUID
	SenderUserID      uuid.UUID
	SenderDeviceID    uuid.UUID
	RecipientUserID   uuid.UUID
	RecipientDeviceID uuid.UUID
	CipherText        string

	CreatedAt   time.Time
	UpdatedAt   *time.Time
	DeliveredAt *time.Time
	ReadAt      *time.Time

	DeletedForAll bool
	DeletedForMe  []uuid.UUID

	X3DHOTPKID *uuid.UUID
	PubKey     string

	GroupID *uuid.UUID

	Reactions map[string]string
}

func New(senderUserID, senderDeviceID, recipientUserID, recipientDeviceID uuid.UUID, msg string) *Message {
	now := time.Now().UTC()

	return &Message{
		ID:                uuid.New(),
		SenderUserID:      senderUserID,
		SenderDeviceID:    senderDeviceID,
		RecipientUserID:   recipientUserID,
		RecipientDeviceID: recipientDeviceID,
		CipherText:        msg,
		CreatedAt:         now,
		Reactions:         map[string]string{},
		DeletedForMe:      []uuid.UUID{},
	}
}
