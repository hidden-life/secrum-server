package messages

import (
	"context"

	"github.com/google/uuid"
)

type RealtimeDelivery interface {
	PushToDevice(context.Context, uuid.UUID, []byte) error
	PushToUser(context.Context, uuid.UUID, []byte) error
	PushToGroup(context.Context, uuid.UUID, []byte) error
	Broadcast(context.Context, []byte) error
}
