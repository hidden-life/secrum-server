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
}
