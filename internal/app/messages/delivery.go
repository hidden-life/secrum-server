package messages

import (
	"context"

	"github.com/google/uuid"
)

type RealtimeDelivery interface {
	Push(context.Context, uuid.UUID, any) error
}
