package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log            *zap.Logger
	userRepository ports.UserRepository
}

// MeResponse is DTO for response like `my profile`
type MeResponse struct {
	UserID            string  `json:"user_id"`
	PhoneHash         string  `json:"phone_hash"`
	DisplayName       *string `json:"display_name,omitempty"`
	AvatarURL         *string `json:"avatar_url,omitempty"`
	StatusMessage     *string `json:"status_message,omitempty"`
	SafetyFingerprint *string `json:"safety_fingerprint,omitempty"`
}

type UpdateProfileRequest struct {
	DisplayName   *string `json:"display_name,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	StatusMessage *string `json:"status_message,omitempty"`
}

func NewService(log *zap.Logger, repo ports.UserRepository) *Service {
	return &Service{
		log:            log,
		userRepository: repo,
	}
}

func (s *Service) GetMe(ctx context.Context, userID string) (*MeResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	u, err := s.userRepository.GetByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if u == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &MeResponse{
		UserID:            u.ID.String(),
		PhoneHash:         u.PhoneHash,
		DisplayName:       u.DisplayName,
		AvatarURL:         u.AvatarURL,
		StatusMessage:     u.StatusMessage,
		SafetyFingerprint: u.SafetyFingerprint,
	}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*MeResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	u, err := s.userRepository.GetByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if u == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Let's update only fields that came in Request
	if req.DisplayName != nil {
		u.DisplayName = req.DisplayName
	}

	if req.AvatarURL != nil {
		u.AvatarURL = req.AvatarURL
	}

	if req.StatusMessage != nil {
		u.StatusMessage = req.StatusMessage
	}
	u.UpdatedAt = time.Now().UTC()

	if err := s.userRepository.UpdateProfile(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return &MeResponse{
		UserID:            u.ID.String(),
		PhoneHash:         u.PhoneHash,
		DisplayName:       u.DisplayName,
		AvatarURL:         u.AvatarURL,
		StatusMessage:     u.StatusMessage,
		SafetyFingerprint: u.SafetyFingerprint,
	}, nil
}
