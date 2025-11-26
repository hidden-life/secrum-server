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

	IsPinned   bool `json:"is_pinned"`
	IsArchived bool `json:"is_archived"`
	IsMuted    bool `json:"is_muted"`
}

type Service struct {
	log           *zap.Logger
	messageRepo   ports.MessageRepository
	userRepo      ports.UserRepository
	chatStateRepo ports.ChatStateRepository
}

func NewService(log *zap.Logger, msgRepo ports.MessageRepository, userRepo ports.UserRepository, chatStateRepo ports.ChatStateRepository) *Service {
	return &Service{
		log:           log,
		messageRepo:   msgRepo,
		userRepo:      userRepo,
		chatStateRepo: chatStateRepo,
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

		state, err := s.chatStateRepo.GetState(ctx, userID, cs.PeerUserID)
		if err != nil {
			s.log.Warn("failed to load state", zap.Error(err), zap.String("peer_user_id", cs.PeerUserID.String()))
		}

		item.IsPinned = state.Pinned
		item.IsArchived = state.Archived
		item.IsMuted = state.Muted

		result = append(result, item)
	}

	return result, nil
}

func (s *Service) SetPinned(ctx context.Context, userID, peerID string, isPinned bool) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	pID, err := uuid.Parse(peerID)
	if err != nil {
		return fmt.Errorf("invalid peer id: %w", err)
	}

	return s.chatStateRepo.SetPinned(ctx, uID, pID, isPinned)
}

func (s *Service) SetMuted(ctx context.Context, userID, peerID string, isMuted bool) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	pID, err := uuid.Parse(peerID)
	if err != nil {
		return fmt.Errorf("invalid peer id: %w", err)
	}

	return s.chatStateRepo.SetMuted(ctx, uID, pID, isMuted)
}

func (s *Service) SetArchived(ctx context.Context, userID, peerID string, isArchived bool) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	pID, err := uuid.Parse(peerID)
	if err != nil {
		return fmt.Errorf("invalid peer id: %w", err)
	}

	return s.chatStateRepo.SetArchived(ctx, uID, pID, isArchived)
}
