package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/key"
)

type OneTimePreKeyRepository interface {
	BulkInsert(context.Context, []*key.OneTimePreKey) error
	GetOneUnusedAndMarkUsed(context.Context, uuid.UUID) (*key.OneTimePreKey, error)
}
