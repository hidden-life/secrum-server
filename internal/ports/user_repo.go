package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/user"
)

type UserRepository interface {
	GetByID(context.Context, uuid.UUID) (*user.User, error)
	GetByPhoneHash(context.Context, string) (*user.User, error)
	Create(context.Context, *user.User) error
	UpdateProfile(context.Context, *user.User) error

	UpdateAllowedMimeTypes(context.Context, uuid.UUID, []string) error
	GetAllowedMimeTypes(context.Context, uuid.UUID) ([]string, error)

	GetByUsername(context.Context, string) (*user.User, error)
}
