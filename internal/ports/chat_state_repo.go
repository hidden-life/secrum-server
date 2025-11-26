package ports

import (
	"context"

	"github.com/google/uuid"
)

type ChatState struct {
	Pinned   bool
	Archived bool
	Muted    bool
}

type ChatStateRepository interface {
	SetPinned(ctx context.Context, userID, peerID uuid.UUID, isPinned bool) error
	SetArchived(ctx context.Context, userID, peerID uuid.UUID, isArchived bool) error
	SetMuted(ctx context.Context, userID, peerID uuid.UUID, isMuted bool) error

	GetState(ctx context.Context, userID, peerID uuid.UUID) (*ChatState, error)
}
