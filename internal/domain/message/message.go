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
	DeliveredAt *time.Time
	ReadAt      *time.Time

	X3DHOTPKID *uuid.UUID
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
	}
}
