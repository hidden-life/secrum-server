package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/attachment"
)

type AttachmentRepository interface {
	Create(context.Context, *attachment.Attachment) error
	GetByID(context.Context, uuid.UUID) (*attachment.Attachment, error)
	MarkDeleted(context.Context, uuid.UUID) error
}
