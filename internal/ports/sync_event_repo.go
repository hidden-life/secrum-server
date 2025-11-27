package ports

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SyncEvent struct {
	ID        int64           `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

type SyncEventRepository interface {
	// Append adds event for user and returns cursor ID
	Append(ctx context.Context, userID uuid.UUID, eventType string, payload any) (int64, error)
	// ListSince returns event > after for user
	ListSince(ctx context.Context, userID uuid.UUID, after int64, limit int) ([]SyncEvent, error)
	// GetLastID returns max ID of user events (or 0 if there are no)
	GetLastID(ctx context.Context, userID uuid.UUID) (int64, error)
}
