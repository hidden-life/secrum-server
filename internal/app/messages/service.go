package messages

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/message"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log            *zap.Logger
	msgRepository  ports.MessageRepository
	userRepository ports.UserRepository
}

// SendRequest is input for sending message
type SendRequest struct {
	RecipientUserID   string `json:"recipient_user_id"`
	RecipientDeviceID string `json:"recipient_device_id"`
	CipherText        string `json:"cipher_text"`
}

// SendResponse contains created message ID.
type SendResponse struct {
	MessageID string `json:"message_id"`
}

// PendingMessage is DTO for outgoing response.
type PendingMessage struct {
	ID                string `json:"id"`
	SenderUserID      string `json:"sender_user_id"`
	SenderDeviceID    string `json:"sender_device_id"`
	RecipientUserID   string `json:"recipient_user_id"`
	RecipientDeviceID string `json:"recipient_device_id"`
	CipherText        string `json:"cipher_text"`
	CreatedAt         string `json:"created_at"`
}

// AckRequest contains message IDs to mark as delivered
type AckRequest struct {
	MessageIDs []string `json:"message_ids"`
}

func NewService(log *zap.Logger, msgRepository ports.MessageRepository, userRepository ports.UserRepository) *Service {
	return &Service{
		log:            log,
		msgRepository:  msgRepository,
		userRepository: userRepository,
	}
}

func (s *Service) Send(ctx context.Context, sUserID, sDeviceID string, req *SendRequest) (*SendResponse, error) {
	if req.CipherText == "" || req.RecipientUserID == "" || req.RecipientDeviceID == "" {
		return nil, fmt.Errorf("missing required parameters")
	}

	senderUserID, err := uuid.Parse(sUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid sender user id: %w", err)
	}
	senderDeviceID, err := uuid.Parse(sDeviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid sender device id: %w", err)
	}

	recipientUserID, err := uuid.Parse(req.RecipientUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient user id: %w", err)
	}
	recipientDeviceID, err := uuid.Parse(req.RecipientDeviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient device id: %w", err)
	}

	// Optionally: we can check receiver (user) using userRepository (TODO)
	msg := message.New(senderUserID, senderDeviceID, recipientUserID, recipientDeviceID, req.CipherText)

	if err := s.msgRepository.Save(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	s.log.Info("Encrypted message stored", zap.String("msg_id", msg.ID.String()), zap.String("to_user", req.RecipientUserID), zap.String("from_user", senderUserID.String()))

	return &SendResponse{MessageID: msg.ID.String()}, nil
}

// FetchPending returns messages pending for given device
func (s *Service) FetchPending(ctx context.Context, deviceID string, limit int) ([]PendingMessage, error) {
	if limit <= 0 || limit >= 500 {
		limit = 500
	}

	devID, err := uuid.Parse(deviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid device id: %w", err)
	}

	msgs, err := s.msgRepository.GetPendingByRecipientDevice(ctx, devID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending messages: %w", err)
	}

	res := make([]PendingMessage, 0, len(msgs))
	for _, m := range msgs {
		res = append(res, PendingMessage{
			ID:                m.ID.String(),
			SenderDeviceID:    m.SenderDeviceID.String(),
			SenderUserID:      m.SenderUserID.String(),
			RecipientUserID:   m.RecipientUserID.String(),
			CipherText:        m.CipherText,
			RecipientDeviceID: m.RecipientDeviceID.String(),
			CreatedAt:         m.CreatedAt.Format(time.RFC3339Nano),
		})
	}

	return res, nil
}

func (s *Service) AckDelivered(ctx context.Context, deviceID string, req AckRequest) error {
	if len(req.MessageIDs) == 0 {
		return nil
	}

	var ids []uuid.UUID
	for _, idStr := range req.MessageIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("invalid message id: %w", err)
		}
		ids = append(ids, id)
	}

	if err := s.msgRepository.MarkDelivered(ctx, ids); err != nil {
		return fmt.Errorf("failed to mark delivered message: %w", err)
	}

	return nil
}
