package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/contact"
)

type ContactRepository interface {
	Add(context.Context, *contact.Contact) error
	Remove(context.Context, uuid.UUID, uuid.UUID) error
	List(context.Context, uuid.UUID) ([]*contact.Contact, error)
	Exists(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}
