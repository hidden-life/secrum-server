package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/app/chats"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type FullSyncResponse struct {
	Chats  []chats.ChatItem `json:"chats"`
	Cursor int64            `json:"cursor"`
}

type DeltaEvent struct {
	ID        int64           `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt string          `json:"created_at"`
}

type DeltaSyncResponse struct {
	Events []DeltaEvent `json:"events"`
	Cursor int64        `json:"cursor"`
}

type Service struct {
	log           *zap.Logger
	chatSvc       *chats.Service
	syncEventRepo ports.SyncEventRepository
}

func NewService(log *zap.Logger, chatSvc *chats.Service, repo ports.SyncEventRepository) *Service {
	return &Service{
		log:           log,
		chatSvc:       chatSvc,
		syncEventRepo: repo,
	}
}

func (s *Service) Full(ctx context.Context, userID string) (*FullSyncResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	chatsList, err := s.chatSvc.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	cursor, err := s.syncEventRepo.GetLastID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to read sync cursor: %w", err)
	}

	return &FullSyncResponse{
		Chats:  chatsList,
		Cursor: cursor,
	}, nil
}

func (s *Service) Delta(ctx context.Context, userID string, cursor int64, limit int) (*DeltaSyncResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	events, err := s.syncEventRepo.ListSince(ctx, uid, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list sync events: %w", err)
	}

	out := make([]DeltaEvent, 0, len(events))
	newCursor := cursor
	for _, event := range events {
		out = append(out, DeltaEvent{
			ID:        event.ID,
			Type:      event.Type,
			Payload:   event.Payload,
			CreatedAt: event.CreatedAt.Format(time.RFC3339Nano),
		})

		if event.ID > newCursor {
			newCursor = event.ID
		}
	}

	return &DeltaSyncResponse{
		Events: out,
		Cursor: newCursor,
	}, nil
}
