package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/message"
)

// MessageRepository defines operations for encrypted messages storage.
type MessageRepository interface {
	Save(ctx context.Context, message *message.Message) error
	GetPendingByRecipientDevice(context.Context, uuid.UUID, int) ([]*message.Message, error)
	MarkDelivered(context.Context, []uuid.UUID) error
	MarkRead(context.Context, []uuid.UUID) error
}
