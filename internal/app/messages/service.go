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
	log              *zap.Logger
	msgRepository    ports.MessageRepository
	userRepository   ports.UserRepository
	deviceRepository ports.DeviceRepository
}

// SendRequest is input for sending message
type SendRequest struct {
	RecipientUserID   string  `json:"recipient_user_id"`
	RecipientDeviceID string  `json:"recipient_device_id"`
	CipherText        string  `json:"cipher_text"`
	X3DHOTPKID        *string `json:"x3dh_otpk_id,omitempty"`
	PubKey            *string `json:"pub_key,omitempty"`
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
	PubKey            string `json:"pub_key"`
	CreatedAt         string `json:"created_at"`
}

// AckRequest contains message IDs to mark as delivered
type AckRequest struct {
	Delivered []string `json:"delivered,omitempty"`
	Read      []string `json:"read,omitempty"`
}

func NewService(
	log *zap.Logger,
	msgRepository ports.MessageRepository,
	userRepository ports.UserRepository,
	deviceRepository ports.DeviceRepository,
) *Service {
	return &Service{
		log:              log,
		msgRepository:    msgRepository,
		userRepository:   userRepository,
		deviceRepository: deviceRepository,
	}
}

func (s *Service) Send(ctx context.Context, sUserID, sDeviceID string, req *SendRequest) (*SendResponse, error) {
	if req.CipherText == "" || req.RecipientUserID == "" {
		return nil, fmt.Errorf("missing required parameters (recipient_user_id, cipher_text)")
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

	devices, err := s.deviceRepository.ListActiveByUser(ctx, recipientUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of active recipient devices: %w", err)
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("no active recipient devices found")
	}

	var otpkUUID *uuid.UUID
	if req.X3DHOTPKID != nil && *req.X3DHOTPKID != "" {
		id, err := uuid.Parse(*req.X3DHOTPKID)
		if err != nil {
			return nil, fmt.Errorf("invalid x3dh_otpk_id: %w", err)
		}
		otpkUUID = &id
	}

	var messageIDs []uuid.UUID
	for _, d := range devices {
		msg := message.New(senderUserID, senderDeviceID, recipientUserID, d.ID, req.CipherText)
		msg.X3DHOTPKID = otpkUUID
		if req.PubKey != nil {
			msg.PubKey = *req.PubKey
		}

		if err := s.msgRepository.Save(ctx, msg); err != nil {
			return nil, fmt.Errorf("failed to save message for device %s: %w", d.ID.String(), err)
		}

		messageIDs = append(messageIDs, d.ID)
	}

	s.log.Info(
		"Encrypted message stored (multi-device)",
		zap.String("from_user", senderUserID.String()),
		zap.String("from_device", req.RecipientDeviceID),
		zap.String("to_user", req.RecipientUserID),
		zap.Int("devices_count", len(devices)),
	)

	first := ""
	if len(messageIDs) > 0 {
		first = messageIDs[0].String()
	}

	return &SendResponse{
		MessageID: first,
	}, nil
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
			PubKey:            m.PubKey,
		})
	}

	return res, nil
}

func (s *Service) AckDelivered(ctx context.Context, deviceID string, req AckRequest) error {
	if len(req.Delivered) > 0 {
		var ids []uuid.UUID
		for _, str := range req.Delivered {
			id, err := uuid.Parse(str)
			if err != nil {
				return fmt.Errorf("invalid delivered message id: %w", err)
			}

			ids = append(ids, id)
		}

		if err := s.msgRepository.MarkDelivered(ctx, ids); err != nil {
			return fmt.Errorf("failed to mark delivered messages: %w", err)
		}
	}

	if len(req.Read) > 0 {
		var ids []uuid.UUID
		for _, str := range req.Read {
			id, err := uuid.Parse(str)
			if err != nil {
				return fmt.Errorf("invalid read message id: %w", err)
			}
			ids = append(ids, id)
		}

		if repo, isOk := s.msgRepository.(interface {
			MarkRead(ctx context.Context, ids []uuid.UUID) error
		}); isOk {
			if err := repo.MarkRead(ctx, ids); err != nil {
				return fmt.Errorf("failed to mark read messages: %w", err)
			}
		} else {
			s.log.Warn("msgRepository does not implement MarkRead(...)")
		}
	}

	return nil
}
