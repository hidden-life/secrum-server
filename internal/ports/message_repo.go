package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/message"
)

type ChatSummary struct {
	PeerUserID     uuid.UUID
	LastCipherText string
	LastMessageAt  time.Time
	UnreadCount    int
}

// MessageRepository defines operations for encrypted messages storage.
type MessageRepository interface {
	Save(ctx context.Context, message *message.Message) error
	SaveMany(context.Context, []*message.Message) error
	GetPendingByRecipientDevice(context.Context, uuid.UUID, int) ([]*message.Message, error)
	MarkDelivered(context.Context, []uuid.UUID) error
	MarkRead(context.Context, []uuid.UUID) error
	UserChatsList(context.Context, uuid.UUID) ([]ChatSummary, error)
	GetGroupMessages(context.Context, uuid.UUID, int, *time.Time) ([]*message.Message, error)
	GetByIDs(context.Context, []uuid.UUID) ([]*message.Message, error)
	GetChatHistory(context.Context, uuid.UUID, uuid.UUID, int, *time.Time) ([]*message.Message, error)
	DeleteForAll(context.Context, uuid.UUID) error
	DeleteForMe(context.Context, uuid.UUID, uuid.UUID) error
	Edit(context.Context, uuid.UUID, string, string, *uuid.UUID) error
	AddReaction(context.Context, uuid.UUID, uuid.UUID, string) error
	RemoveReaction(context.Context, uuid.UUID, uuid.UUID) error

	PinMessage(context.Context, uuid.UUID, uuid.UUID) error
	UnpinMessage(context.Context, uuid.UUID, uuid.UUID) error

	SearchMessages(context.Context, uuid.UUID, string, int, *time.Time) ([]*message.Message, error)

	FindMessageByAttachmentID(context.Context, uuid.UUID) (*message.Message, error)
}
