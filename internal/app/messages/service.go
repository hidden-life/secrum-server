package messages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/message"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/hidden-life/secrum-server/internal/real_time"
	"go.uber.org/zap"
)

type Service struct {
	log              *zap.Logger
	msgRepository    ports.MessageRepository
	userRepository   ports.UserRepository
	deviceRepository ports.DeviceRepository
	realtimeDelivery RealtimeDelivery
	syncRepository   ports.SyncEventRepository

	groupRepository       ports.GroupRepository
	groupMemberRepository ports.GroupMemberRepository
}

type MediaMetadata struct {
	MimeType   string  `json:"mime_type"`
	SizeBytes  int64   `json:"size_bytes"`
	DurationMs *int    `json:"duration_ms,omitempty"`
	Width      *int    `json:"width,omitempty"`
	Height     *int    `json:"height,omitempty"`
	BlurHash   *string `json:"blur_hash,omitempty"`
}

// SendRequest is input for sending message
type SendRequest struct {
	RecipientUserID   string  `json:"recipient_user_id"`
	RecipientDeviceID string  `json:"recipient_device_id"`
	CipherText        string  `json:"cipher_text"`
	X3DHOTPKID        *string `json:"x3dh_otpk_id,omitempty"`
	PubKey            *string `json:"pub_key,omitempty"`

	QuotedMessageID *string        `json:"quoted_message_id,omitempty"`
	Media           *MediaMetadata `json:"media,omitempty"`

	AttachmentID *string `json:"attachment_id,omitempty"`
}

// SendResponse contains created message ID.
type SendResponse struct {
	MessageID string `json:"message_id"`
}

type SendGroupMessageRequest struct {
	CipherText string  `json:"cipher_text"`
	PubKey     *string `json:"pub_key,omitempty"`
	X3DHOTPKID *string `json:"x3dh_otpk_id,omitempty"`

	QuotedMessageID *string        `json:"quoted_message_id,omitempty"`
	Media           *MediaMetadata `json:"media,omitempty"`

	AttachmentID *string `json:"attachment_id,omitempty"`
}

// PendingMessage is DTO for outgoing response.
type PendingMessage struct {
	ID                string `json:"id"`
	SenderUserID      string `json:"sender_user_id"`
	SenderDeviceID    string `json:"sender_device_id"`
	RecipientUserID   string `json:"recipient_user_id"`
	RecipientDeviceID string `json:"recipient_device_id"`

	CipherText string `json:"cipher_text"`
	PubKey     string `json:"pub_key"`
	CreatedAt  string `json:"created_at"`

	DeliveredAt *string `json:"delivered_at,omitempty"`
	ReadAt      *string `json:"read_at,omitempty"`

	IsEdited  bool `json:"is_edited"`
	IsDeleted bool `json:"is_deleted"`

	ForwardedFromMessageID *string `json:"forwarded_from_message_id,omitempty"`
	ForwardedFromUserID    *string `json:"forwarded_from_user_id,omitempty"`
	QuotedMessageID        *string `json:"quoted_message_id,omitempty"`

	HasMedia bool           `json:"has_media"`
	Media    *MediaMetadata `json:"media,omitempty"`
}

// AckRequest contains message IDs to mark as delivered
type AckRequest struct {
	Delivered []string `json:"delivered,omitempty"`
	Read      []string `json:"read,omitempty"`
}

type EditMessageRequest struct {
	CipherText string  `json:"text"`
	PubKey     string  `json:"pub_key,omitempty"`
	OTPK       *string `json:"x3dh_otpk_id,omitempty"`
}

func NewService(
	log *zap.Logger,
	msgRepository ports.MessageRepository,
	userRepository ports.UserRepository,
	deviceRepository ports.DeviceRepository,
	realtime RealtimeDelivery,
	syncRepository ports.SyncEventRepository,
	groupRepository ports.GroupRepository,
	groupMemberRepository ports.GroupMemberRepository,
) *Service {
	return &Service{
		log:                   log,
		msgRepository:         msgRepository,
		userRepository:        userRepository,
		deviceRepository:      deviceRepository,
		realtimeDelivery:      realtime,
		syncRepository:        syncRepository,
		groupRepository:       groupRepository,
		groupMemberRepository: groupMemberRepository,
	}
}

// Send allows to send single message
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

	if req.Media != nil && req.Media.MimeType != "" {
		// system policy
		// we will use same policy as in attachments
		sysAllowed := func(mime string) bool {
			switch strings.ToLower(mime) {
			case "image/png", "image/jpeg", "image/webp", "image/gif", "video/mp4", "application/pdf":
				return true
			default:
				return false
			}
		}

		userAllowed, _ := s.userRepository.GetAllowedMimeTypes(ctx, senderUserID)
		var groupAllowed []string
		if !intersectAllowed(sysAllowed, groupAllowed, userAllowed, req.Media.MimeType) {
			return nil, fmt.Errorf("media type '%q' is not allowed for this user", req.Media.MimeType)
		}
	}

	// load all active devices of recipient
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

	var first string
	for _, d := range devices {
		msg := message.New(senderUserID, senderDeviceID, recipientUserID, d.ID, req.CipherText)
		msg.X3DHOTPKID = otpkUUID
		if req.PubKey != nil {
			msg.PubKey = *req.PubKey
		}

		// quoted
		if req.QuotedMessageID != nil && *req.QuotedMessageID != "" {
			if qid, err := uuid.Parse(*req.QuotedMessageID); err == nil {
				msg.QuotedMessageID = &qid
			}
		}

		// media
		if req.Media != nil {
			msg.HasMedia = true
			msg.MediaMimeType = &req.Media.MimeType
			msg.MediaSizeBytes = &req.Media.SizeBytes
			msg.MediaDurationMs = req.Media.DurationMs
			msg.MediaWidth = req.Media.Width
			msg.MediaHeight = req.Media.Height
			msg.MediaBlurHash = req.Media.BlurHash
		}

		if req.AttachmentID != nil {
			id, err := uuid.Parse(*req.AttachmentID)
			if err != nil {
				return nil, fmt.Errorf("invalid attachment id: %w", err)
			}
			msg.AttachmentID = &id
		}

		if err := s.msgRepository.Save(ctx, msg); err != nil {
			return nil, fmt.Errorf("failed to save message for device %s: %w", d.ID.String(), err)
		}

		if first == "" {
			first = msg.ID.String()
		}

		e := real_time.EventMessage{
			ID:                msg.ID.String(),
			SenderUserID:      msg.SenderUserID.String(),
			SenderDeviceID:    msg.SenderDeviceID.String(),
			RecipientUserID:   msg.RecipientUserID.String(),
			RecipientDeviceID: msg.RecipientDeviceID.String(),
			CipherText:        msg.CipherText,
			PubKey:            msg.PubKey,
			CreatedAt:         msg.CreatedAt.Format(time.RFC3339Nano),
		}

		buff, err := real_time.MarshalEvent("message", e)

		if err == nil {
			_ = s.realtimeDelivery.PushToDevice(ctx, msg.RecipientDeviceID, buff)
		}

		if s.syncRepository != nil {
			if _, err = s.syncRepository.Append(ctx, msg.RecipientUserID, "message:new", e); err != nil {
				s.log.Warn("sync append failed", zap.Error(err))
			}
		}
	}

	s.log.Info(
		"Encrypted message stored (multi-device)",
		zap.String("from_user", senderUserID.String()),
		zap.String("from_device", req.RecipientDeviceID),
		zap.String("to_user", req.RecipientUserID),
		zap.Int("devices_count", len(devices)),
	)

	return &SendResponse{
		MessageID: first,
	}, nil
}

// FetchPending returns messages pending for given device
func (s *Service) FetchPending(ctx context.Context, deviceID string, limit int) ([]PendingMessage, error) {
	if limit <= 0 || limit > 500 {
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
		var deliveredAt *string
		if m.DeliveredAt != nil {
			v := m.DeliveredAt.Format(time.RFC3339Nano)
			deliveredAt = &v
		}

		var readAt *string
		if m.ReadAt != nil {
			v := m.ReadAt.Format(time.RFC3339Nano)
			readAt = &v
		}

		res = append(res, PendingMessage{
			ID:                m.ID.String(),
			SenderDeviceID:    m.SenderDeviceID.String(),
			SenderUserID:      m.SenderUserID.String(),
			RecipientUserID:   m.RecipientUserID.String(),
			CipherText:        m.CipherText,
			RecipientDeviceID: m.RecipientDeviceID.String(),
			CreatedAt:         m.CreatedAt.Format(time.RFC3339Nano),
			PubKey:            m.PubKey,

			DeliveredAt: deliveredAt,
			ReadAt:      readAt,

			IsEdited:  m.IsEdited,
			IsDeleted: m.IsDeleted,
		})
	}

	return res, nil
}

func (s *Service) AckDelivered(ctx context.Context, deviceID string, req AckRequest) error {
	var deliveredIDs []uuid.UUID
	var readIDs []uuid.UUID

	// parse delivered
	for _, str := range req.Delivered {
		id, err := uuid.Parse(str)
		if err != nil {
			return fmt.Errorf("invalid delivered message id: %w", err)
		}
		deliveredIDs = append(deliveredIDs, id)
	}

	// parse read
	for _, str := range req.Read {
		id, err := uuid.Parse(str)
		if err != nil {
			return fmt.Errorf("invalid read message id: %w", err)
		}
		readIDs = append(readIDs, id)
	}

	// update delivered_at in database
	if len(deliveredIDs) > 0 {
		if err := s.msgRepository.MarkDelivered(ctx, deliveredIDs); err != nil {
			return fmt.Errorf("failed to mark delivered messages: %w", err)
		}
	}

	// update read_at in database
	if len(readIDs) > 0 {
		if repo, ok := s.msgRepository.(interface {
			MarkRead(ctx context.Context, ids []uuid.UUID) error
		}); ok {
			if err := repo.MarkRead(ctx, readIDs); err != nil {
				return fmt.Errorf("failed to mark read messages: %w", err)
			}
		} else {
			s.log.Warn("msgRepository does not implement MarkRead(...)")
		}
	}

	// WS events to sender (using delivery hub)
	if s.realtimeDelivery != nil && (len(deliveredIDs) > 0 || len(readIDs) > 0) {
		idSet := make(map[uuid.UUID]struct{})

		deliveredSet := make(map[uuid.UUID]struct{})
		for _, id := range deliveredIDs {
			deliveredSet[id] = struct{}{}
			idSet[id] = struct{}{}
		}

		readSet := make(map[uuid.UUID]struct{})
		for _, id := range readIDs {
			readSet[id] = struct{}{}
			idSet[id] = struct{}{}
		}

		allIDs := make([]uuid.UUID, 0, len(idSet))
		for id := range idSet {
			allIDs = append(allIDs, id)
		}

		msgs, err := s.msgRepository.GetByIDs(ctx, allIDs)
		if err != nil {
			// log but not break all ack
			s.log.Warn("failed to load messages for ack events", zap.Error(err))
			return nil
		}

		for _, msg := range msgs {
			// ack delivered
			if _, ok := deliveredSet[msg.ID]; ok {
				e := real_time.EventAckDelivered{
					MessageID:  msg.ID.String(),
					ToUserID:   msg.RecipientUserID.String(),
					ToDeviceID: msg.RecipientDeviceID.String(),
				}

				env := real_time.OutEnvelope{
					Type: "ack_delivered",
					Data: e,
				}

				raw, err := real_time.MarshalEvent("message:delivered", env)
				if err == nil {
					_ = s.realtimeDelivery.PushToDevice(ctx, msg.SenderDeviceID, raw)
				} else {
					s.log.Warn("failed to marshal message:delivered event", zap.Error(err))
				}

				// sync
				if s.syncRepository != nil {
					if _, err := s.syncRepository.Append(ctx, msg.SenderUserID, "message:delivered", e); err != nil {
						s.log.Warn("sync append failed", zap.Error(err))
					}
				}
			}

			// ack read
			if _, ok := readSet[msg.ID]; ok {
				e := real_time.EventAckRead{
					MessageID:  msg.ID.String(),
					ToUserID:   msg.RecipientUserID.String(),
					ToDeviceID: msg.RecipientDeviceID.String(),
				}

				env := real_time.OutEnvelope{
					Type: "ack_read",
					Data: e,
				}

				raw, err := real_time.MarshalEvent("message:read", env)
				if err == nil {
					_ = s.realtimeDelivery.PushToDevice(ctx, msg.SenderDeviceID, raw)
				} else {
					s.log.Warn("failed to marshal message:read event", zap.Error(err))
				}

				// sync
				if s.syncRepository != nil {
					if _, err := s.syncRepository.Append(ctx, msg.SenderUserID, "message:read", e); err != nil {
						s.log.Warn("sync append failed", zap.Error(err))
					}
				}
			}
		}
	}

	return nil
}

func (s *Service) SendGroupMessage(
	ctx context.Context,
	senderUserID, senderDeviceID, groupID string,
	req *SendGroupMessageRequest,
	members []uuid.UUID) (*SendResponse, error) {
	if req.CipherText == "" {
		return nil, fmt.Errorf("cipher_text required")
	}

	sUID, err := uuid.Parse(senderUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid sender user id")
	}
	sDID, err := uuid.Parse(senderDeviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid sender device id")
	}

	gID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	if req.Media != nil && req.Media.MimeType != "" {
		// system policy
		// we will use same policy as in attachments
		sysAllowed := func(mime string) bool {
			switch strings.ToLower(mime) {
			case "image/png", "image/jpeg", "image/webp", "image/gif", "video/mp4", "application/pdf":
				return true
			default:
				return false
			}
		}

		groupAllowed, _ := s.groupRepository.GetAllowedMimeTypes(ctx, gID)
		userAllowed, _ := s.userRepository.GetAllowedMimeTypes(ctx, sUID)
		if !intersectAllowed(sysAllowed, groupAllowed, userAllowed, req.Media.MimeType) {
			return nil, fmt.Errorf("media type '%q' is not allowed for this group/user", req.Media.MimeType)
		}
	}

	var otpkUUID *uuid.UUID
	if req.X3DHOTPKID != nil && *req.X3DHOTPKID != "" {
		id, err := uuid.Parse(*req.X3DHOTPKID)
		if err != nil {
			return nil, fmt.Errorf("invalid x3dh_otpk_id")
		}
		otpkUUID = &id
	}

	var out []*message.Message

	for _, userID := range members {
		devs, err := s.deviceRepository.ListActiveByUser(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to list devices for user %s: %w", userID, err)
		}

		for _, d := range devs {
			msg := message.New(sUID, sDID, userID, d.ID, req.CipherText)
			msg.GroupID = &gID
			msg.X3DHOTPKID = otpkUUID
			if req.PubKey != nil {
				msg.PubKey = *req.PubKey
			}

			// quoted
			if req.QuotedMessageID != nil && *req.QuotedMessageID != "" {
				if qid, err := uuid.Parse(*req.QuotedMessageID); err == nil {
					msg.QuotedMessageID = &qid
				}
			}

			// media
			if req.Media != nil {
				msg.HasMedia = true
				msg.MediaMimeType = &req.Media.MimeType
				msg.MediaSizeBytes = &req.Media.SizeBytes
				msg.MediaDurationMs = req.Media.DurationMs
				msg.MediaWidth = req.Media.Width
				msg.MediaHeight = req.Media.Height
				msg.MediaBlurHash = req.Media.BlurHash
			}

			// attachment
			if req.AttachmentID != nil {
				id, err := uuid.Parse(*req.AttachmentID)
				if err != nil {
					return nil, fmt.Errorf("invalid attachment id")
				}
				msg.AttachmentID = &id
			}

			out = append(out, msg)
		}
	}

	if err := s.msgRepository.SaveMany(ctx, out); err != nil {
		return nil, fmt.Errorf("failed to save group messages: %w", err)
	}

	for _, msg := range out {
		e := real_time.EventGroupMessage{
			ID:                msg.ID.String(),
			SenderUserID:      msg.SenderUserID.String(),
			SenderDeviceID:    msg.SenderDeviceID.String(),
			RecipientUserID:   msg.RecipientUserID.String(),
			RecipientDeviceID: msg.RecipientDeviceID.String(),
			CipherText:        msg.CipherText,
			CreatedAt:         msg.CreatedAt.Format(time.RFC3339Nano),
			PubKey:            msg.PubKey,
			GroupID:           gID.String(),
		}

		buff, err := real_time.MarshalEvent("group_message", e)
		if err == nil {
			_ = s.realtimeDelivery.PushToDevice(ctx, msg.RecipientDeviceID, buff)
		}
		if s.syncRepository != nil {
			if _, err = s.syncRepository.Append(ctx, msg.RecipientUserID, "group:message:new", e); err != nil {
				s.log.Warn("sync append failed", zap.Error(err))
			}
		}
	}

	var first string
	if len(out) > 0 {
		first = out[0].ID.String()
	}

	return &SendResponse{MessageID: first}, nil
}

func (s *Service) FetchGroupHistory(ctx context.Context, gid string, limit int, before *time.Time) ([]PendingMessage, error) {
	groupID, err := uuid.Parse(gid)
	if err != nil {
		return nil, fmt.Errorf("invalid group id")
	}

	msgs, err := s.msgRepository.GetGroupMessages(ctx, groupID, limit, before)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group messages: %w", err)
	}

	res := make([]PendingMessage, 0, len(msgs))
	for _, m := range msgs {
		res = append(res, PendingMessage{
			ID:                m.ID.String(),
			SenderDeviceID:    m.SenderDeviceID.String(),
			SenderUserID:      m.SenderUserID.String(),
			RecipientUserID:   m.RecipientUserID.String(),
			RecipientDeviceID: m.RecipientDeviceID.String(),
			CreatedAt:         m.CreatedAt.Format(time.RFC3339Nano),
			PubKey:            m.PubKey,
			CipherText:        m.CipherText,
		})
	}

	return res, nil
}

func (s *Service) GetChatHistory(ctx context.Context, userID, peerID string, limit int, before *time.Time) ([]PendingMessage, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	pid, err := uuid.Parse(peerID)
	if err != nil {
		return nil, fmt.Errorf("invalid peer id")
	}

	msgs, err := s.msgRepository.GetChatHistory(ctx, uid, pid, limit, before)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat history: %w", err)
	}

	out := make([]PendingMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, PendingMessage{
			ID:                m.ID.String(),
			SenderDeviceID:    m.SenderDeviceID.String(),
			SenderUserID:      m.SenderUserID.String(),
			RecipientUserID:   m.RecipientUserID.String(),
			RecipientDeviceID: m.RecipientDeviceID.String(),
			CipherText:        m.CipherText,
			CreatedAt:         m.CreatedAt.Format(time.RFC3339Nano),
			PubKey:            m.PubKey,
		})
	}

	return out, nil
}

func (s *Service) DeleteForMe(ctx context.Context, userID, msgID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}

	mID, err := uuid.Parse(msgID)
	if err != nil {
		return fmt.Errorf("invalid message id")
	}

	return s.msgRepository.DeleteForMe(ctx, uID, mID)
}

func (s *Service) DeleteForAll(ctx context.Context, actorID, msgID string) error {
	actorUID, err := uuid.Parse(actorID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}

	mID, err := uuid.Parse(msgID)
	if err != nil {
		return fmt.Errorf("invalid message id")
	}

	// here we can add permissions later

	err = s.msgRepository.DeleteForAll(ctx, mID)
	if err != nil {
		return err
	}

	if s.syncRepository != nil {
		_, _ = s.syncRepository.Append(ctx, actorUID, "message:deleted", map[string]string{
			"message_id": msgID,
		})
	}

	return nil
}

func (s *Service) EditMessage(ctx context.Context, userID, msgID string, req EditMessageRequest) error {
	mID, err := uuid.Parse(msgID)
	if err != nil {
		return fmt.Errorf("invalid message id")
	}

	uID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}

	var otpk *uuid.UUID
	if req.OTPK != nil {
		id, _ := uuid.Parse(*req.OTPK)
		otpk = &id
	}

	err = s.msgRepository.Edit(ctx, mID, req.CipherText, req.PubKey, otpk)
	if err != nil {
		return err
	}

	if s.syncRepository != nil {
		_, _ = s.syncRepository.Append(ctx, uID, "message:edited", map[string]string{
			"message_id": mID.String(),
		})
	}

	return nil
}

func (s *Service) AddReaction(ctx context.Context, userID, msgID, emoji string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}

	mID, err := uuid.Parse(msgID)
	if err != nil {
		return fmt.Errorf("invalid message id")
	}

	if s.syncRepository != nil {
		_, _ = s.syncRepository.Append(ctx, uID, "message:reaction:add", map[string]string{
			"message_id": msgID,
			"emoji":      emoji,
		})
	}

	return s.msgRepository.AddReaction(ctx, mID, uID, emoji)
}

func (s *Service) RemoveReaction(ctx context.Context, userID, msgID string) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}

	mID, err := uuid.Parse(msgID)
	if err != nil {
		return fmt.Errorf("invalid message id")
	}

	if s.syncRepository != nil {
		_, _ = s.syncRepository.Append(ctx, uID, "message:reaction:remove", map[string]string{
			"message_id": msgID,
		})
	}

	return s.msgRepository.RemoveReaction(ctx, mID, uID)
}

func (s *Service) ForwardToUser(ctx context.Context, actorUserID, actorDeviceID, srcMessageID, targetUserID string) (*SendResponse, error) {
	actorUID, err := uuid.Parse(actorUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor user id")
	}
	actorDevID, err := uuid.Parse(actorDeviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor device id")
	}
	srcMsgID, err := uuid.Parse(srcMessageID)
	if err != nil {
		return nil, fmt.Errorf("invalid source message id")
	}
	targetUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid target user id")
	}

	msgs, err := s.msgRepository.GetByIDs(ctx, []uuid.UUID{srcMsgID})
	if err != nil || len(msgs) == 0 {
		return nil, fmt.Errorf("source message not found")
	}
	src := msgs[0]

	devices, err := s.deviceRepository.ListActiveByUser(ctx, targetUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active devices: %w", err)
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("active device not found")
	}

	var first string
	for _, d := range devices {
		msg := message.New(actorUID, actorDevID, targetUID, d.ID, src.CipherText)
		msg.PubKey = src.PubKey
		msg.X3DHOTPKID = src.X3DHOTPKID
		msg.ForwardedFromMessageID = &src.ID
		msg.ForwardedFromUserID = &src.SenderUserID

		// move metadata of media
		msg.HasMedia = src.HasMedia
		msg.MediaMimeType = src.MediaMimeType
		msg.MediaSizeBytes = src.MediaSizeBytes
		msg.MediaDurationMs = src.MediaDurationMs
		msg.MediaWidth = src.MediaWidth
		msg.MediaHeight = src.MediaHeight
		msg.MediaBlurHash = src.MediaBlurHash

		if err := s.msgRepository.Save(ctx, msg); err != nil {
			return nil, fmt.Errorf("failed to save forwarded message: %w", err)
		}

		if first == "" {
			first = msg.ID.String()
		}

		// here we can push it over realtimeDelivery hub, as in Send(...)

		if s.syncRepository != nil {
			_, _ = s.syncRepository.Append(ctx, actorUID, "message:forwarded", map[string]string{
				"new_message_id":    msg.ID.String(),
				"source_message_id": src.ID.String(),
			})
		}
	}

	return &SendResponse{MessageID: first}, nil
}

func (s *Service) PinMessage(ctx context.Context, userID, messageID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}
	mid, err := uuid.Parse(messageID)
	if err != nil {
		return fmt.Errorf("invalid message id")
	}

	if s.syncRepository != nil {
		_, _ = s.syncRepository.Append(ctx, uid, "chat:pin", map[string]string{
			"message_id": messageID,
		})
	}

	return s.msgRepository.PinMessage(ctx, mid, uid)
}

func (s *Service) UnpinMessage(ctx context.Context, userID, messageID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id")
	}
	mid, err := uuid.Parse(messageID)
	if err != nil {
		return fmt.Errorf("invalid message id")
	}

	if s.syncRepository != nil {
		_, _ = s.syncRepository.Append(ctx, uid, "chat:unpin", map[string]string{
			"message_id": messageID,
		})
	}

	return s.msgRepository.UnpinMessage(ctx, mid, uid)
}

func (s *Service) SearchMessages(ctx context.Context, userID, query string, limit int, before *time.Time) ([]PendingMessage, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	msgs, err := s.msgRepository.SearchMessages(ctx, uid, query, limit, before)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	res := make([]PendingMessage, 0, len(msgs))
	for _, msg := range msgs {
		pm := PendingMessage{
			ID:                msg.ID.String(),
			SenderDeviceID:    msg.SenderDeviceID.String(),
			SenderUserID:      msg.SenderUserID.String(),
			RecipientUserID:   msg.RecipientUserID.String(),
			RecipientDeviceID: msg.RecipientDeviceID.String(),
			CipherText:        msg.CipherText,
			CreatedAt:         msg.CreatedAt.Format(time.RFC3339Nano),
			PubKey:            msg.PubKey,
		}

		if msg.ForwardedFromMessageID != nil {
			id := msg.ForwardedFromMessageID.String()
			pm.ForwardedFromMessageID = &id
		}

		if msg.ForwardedFromUserID != nil {
			id := msg.ForwardedFromUserID.String()
			pm.ForwardedFromUserID = &id
		}

		if msg.QuotedMessageID != nil {
			id := msg.QuotedMessageID.String()
			pm.QuotedMessageID = &id
		}

		if msg.HasMedia {
			mm := &MediaMetadata{
				MimeType:   "",
				SizeBytes:  0,
				DurationMs: nil,
				Width:      nil,
				Height:     nil,
				BlurHash:   nil,
			}

			if msg.MediaMimeType != nil {
				mm.MimeType = *msg.MediaMimeType
			}

			if msg.MediaSizeBytes != nil {
				mm.SizeBytes = *msg.MediaSizeBytes
			}

			mm.DurationMs = msg.MediaDurationMs
			mm.Width = msg.MediaWidth
			mm.Height = msg.MediaHeight
			mm.BlurHash = msg.MediaBlurHash

			pm.HasMedia = true
			pm.Media = mm
		}

		res = append(res, pm)
	}

	return res, nil
}
