package chats

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type ChatItem struct {
	PeerUserID      string  `json:"peer_user_id"`
	PeerDisplayName *string `json:"peer_display_name,omitempty"`
	LastCipherText  string  `json:"last_cipher_text"`
	LastMessageAt   string  `json:"last_message_at"`
	UnreadCount     int     `json:"unread_count"`
}

type Service struct {
	log         *zap.Logger
	messageRepo ports.MessageRepository
	userRepo    ports.UserRepository
}

func NewService(log *zap.Logger, msgRepo ports.MessageRepository, userRepo ports.UserRepository) *Service {
	return &Service{
		log:         log,
		messageRepo: msgRepo,
		userRepo:    userRepo,
	}
}

func (s *Service) List(ctx context.Context, uid string) ([]ChatItem, error) {
	if uid == "" {
		return nil, fmt.Errorf("user id is required")
	}

	userID, err := uuid.Parse(uid)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	summaries, err := s.messageRepo.UserChatsList(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	result := make([]ChatItem, 0, len(summaries))
	for _, cs := range summaries {
		var displayName *string
		u, err := s.userRepo.GetByID(ctx, cs.PeerUserID)
		if err != nil {
			s.log.Warn("failed to load peer user", zap.Error(err), zap.String("peer_user_id", cs.PeerUserID.String()))
		} else if u != nil && u.DisplayName != nil {
			displayName = u.DisplayName
		}

		item := ChatItem{
			PeerUserID:      cs.PeerUserID.String(),
			PeerDisplayName: displayName,
			LastCipherText:  cs.LastCipherText,
			LastMessageAt:   cs.LastMessageAt.Format(time.RFC3339Nano),
			UnreadCount:     cs.UnreadCount,
		}

		result = append(result, item)
	}

	return result, nil
}
