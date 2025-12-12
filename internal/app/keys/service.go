package keys

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/key"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log         *zap.Logger
	devicesRepo ports.DeviceRepository
	otpkRepo    ports.OneTimePreKeyRepository
}

type UploadRequest struct {
	IdentityKey           string   `json:"identity_key"`
	SignedPreKey          string   `json:"signed_prekey"`
	SignedPreKeySignature string   `json:"signed_prekey_sig"`
	OneTimePreKeys        []string `json:"one_time_prekeys"`
}

type DeviceBundle struct {
	DeviceID              string `json:"device_id"`
	IdentityKey           string `json:"identity_key"`
	SignedPreKey          string `json:"signed_prekey"`
	SignedPreKeySignature string `json:"signed_prekey_sig"`
	OneTimePreKey         *struct {
		ID        string `json:"id"`
		PublicKey string `json:"public_key"`
	} `json:"one_time_prekey,omitempty"`
}

type BundleResponse struct {
	UserID  string         `json:"user_id"`
	Devices []DeviceBundle `json:"devices"`
}

func NewService(log *zap.Logger, devicesRepo ports.DeviceRepository, otpkRepo ports.OneTimePreKeyRepository) *Service {
	return &Service{
		log:         log,
		devicesRepo: devicesRepo,
		otpkRepo:    otpkRepo,
	}
}

func (s *Service) UploadDeviceKeys(ctx context.Context, devID uuid.UUID, req UploadRequest) error {
	dev, err := s.devicesRepo.GetById(ctx, devID)
	if err != nil {
		return err
	}

	if dev == nil {
		return fmt.Errorf("device not found")
	}

	dev.IdentityKey = req.IdentityKey
	dev.SignedPreKey = req.SignedPreKey
	dev.SignedPreKeySignature = req.SignedPreKeySignature

	if err := s.devicesRepo.UpdateKeys(ctx, devID, dev.IdentityKey, dev.SignedPreKey, dev.SignedPreKeySignature); err != nil {
		return fmt.Errorf("failed to update keys: %w", err)
	}

	if len(req.OneTimePreKeys) > 0 {
		list := make([]*key.OneTimePreKey, 0, len(req.OneTimePreKeys))
		for _, pub := range req.OneTimePreKeys {
			list = append(list, key.New(dev.ID, pub))
		}

		if err := s.otpkRepo.BulkInsert(ctx, list); err != nil {
			s.log.Warn("failed inserting OTPKs", zap.Error(err))
		}
	}

	s.log.Info("device keys uploaded", zap.String("device_id", dev.ID.String()))
	return nil
}

func (s *Service) GetUserBundle(ctx context.Context, userID uuid.UUID) (*BundleResponse, error) {
	devices, err := s.devicesRepo.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result = &BundleResponse{
		UserID:  userID.String(),
		Devices: make([]DeviceBundle, 0, len(devices)),
	}

	for _, d := range devices {
		var otpk *key.OneTimePreKey
		otpk, err := s.otpkRepo.GetOneUnusedAndMarkUsed(ctx, d.ID)
		if err != nil {
			return nil, err
		}

		db := DeviceBundle{
			DeviceID:              d.ID.String(),
			IdentityKey:           d.IdentityKey,
			SignedPreKey:          d.SignedPreKey,
			SignedPreKeySignature: d.SignedPreKeySignature,
		}

		if otpk != nil {
			db.OneTimePreKey = &struct {
				ID        string `json:"id"`
				PublicKey string `json:"public_key"`
			}{
				ID:        otpk.ID.String(),
				PublicKey: otpk.PublicKey,
			}
		}

		result.Devices = append(result.Devices, db)
	}

	return result, nil
}
