package keys

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/key"
	keybundle "github.com/hidden-life/secrum-server/internal/domain/key_bundle"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log            *zap.Logger
	repository     ports.KeyRepository
	otpkRepository ports.OneTimePreKeyRepository
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

type PreKeyBundleRequest struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id,omitempty"`
}

type PreKeyBundleResponse struct {
	UserID        string `json:"user_id"`
	DeviceID      string `json:"device_id"`
	IdentityKey   string `json:"identity_key"`
	SignedPreKey  string `json:"signed_prekey"`
	OneTimePreKey *struct {
		ID        string `json:"id"`
		PublicKey string `json:"public_key"`
	} `json:"one_time_prekey,omitempty"`
}

type FetchRequest struct {
	DeviceID string `json:"device_id"`
}

func NewService(log *zap.Logger, repo ports.KeyRepository, otpk ports.OneTimePreKeyRepository) *Service {
	return &Service{
		log:            log,
		repository:     repo,
		otpkRepository: otpk,
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

	// add upload one-time per key
	if len(req.OneTimePreKeys) > 0 {
		var list []*key.OneTimePreKey
		for _, pub := range req.OneTimePreKeys {
			list = append(list, key.New(deviceID, pub))
		}

		if err := s.otpkRepository.BulkInsert(ctx, list); err != nil {
			s.log.Warn("failed to bulk insert one-time per keys", zap.Error(err))
		}
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

func (s *Service) PreKeyBundle(ctx context.Context, req *PreKeyBundleRequest) (*PreKeyBundleResponse, error) {
	if req.DeviceID == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	devID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("invalid device ID: %w", err)
	}

	kb, err := s.repository.GetByDeviceID(ctx, devID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key bundle: %w", err)
	}
	if kb == nil {
		return nil, fmt.Errorf("no keys found for device")
	}

	// take one free one-time prekey
	otpk, err := s.otpkRepository.GetOneUnusedAndMarkUsed(ctx, devID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch one-time pre key bundle: %w", err)
	}

	resp := &PreKeyBundleResponse{
		UserID:       req.UserID,
		DeviceID:     req.DeviceID,
		IdentityKey:  kb.IdentityKey,
		SignedPreKey: kb.SignedPreKey,
	}

	if otpk != nil {
		resp.OneTimePreKey = &struct {
			ID        string `json:"id"`
			PublicKey string `json:"public_key"`
		}{
			ID: otpk.ID.String(), PublicKey: otpk.PublicKey,
		}
	}

	return resp, nil
}
