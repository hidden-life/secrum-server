package keys

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	keybundle "github.com/hidden-life/secrum-server/internal/domain/key_bundle"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log        *zap.Logger
	repository ports.KeyRepository
}

type UploadRequest struct {
	DeviceID       string   `json:"device_id"`
	IdentityKey    string   `json:"identity_key"`
	SignedPreKey   string   `json:"signed_prekey"`
	OneTimePreKeys []string `json:"one_time_prekeys"`
}

type UploadResponse struct {
	Message string `json:"message"`
}

type FetchRequest struct {
	DeviceID string `json:"device_id"`
}

func NewService(log *zap.Logger, repo ports.KeyRepository) *Service {
	return &Service{
		log:        log,
		repository: repo,
	}
}

func (s *Service) Upload(ctx context.Context, req UploadRequest) (*UploadResponse, error) {
	deviceID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid device ID: %w", err)
	}

	kb := keybundle.New(deviceID, req.IdentityKey, req.SignedPreKey, req.OneTimePreKeys)
	if err := s.repository.Save(ctx, kb); err != nil {
		return nil, fmt.Errorf("failed to save key bundle: %w", err)
	}

	s.log.Info("Device key bundle uploaded", zap.String("device_id", req.DeviceID))

	return &UploadResponse{Message: "Key bundle uploaded successfully"}, nil
}

func (s *Service) Fetch(ctx context.Context, req FetchRequest) (*keybundle.KeyBundle, error) {
	deviceID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid device ID: %w", err)
	}

	kb, err := s.repository.GetByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key bundle: %w", err)
	}

	if kb == nil {
		return nil, fmt.Errorf("no keys found for device")
	}

	return kb, nil
}
