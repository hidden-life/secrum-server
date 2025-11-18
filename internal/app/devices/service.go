package devices

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log              *zap.Logger
	deviceRepository ports.DeviceRepository
}

type DeviceDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Platform  string `json:"platform"`
	IsActive  bool   `json:"is_active"`
	IsCurrent bool   `json:"is_current"`
	CreatedAt string `json:"created_at"`
	LastSeen  string `json:"last_seen"`
}

func NewService(log *zap.Logger, deviceRepository ports.DeviceRepository) *Service {
	return &Service{
		log:              log,
		deviceRepository: deviceRepository,
	}
}

func (s *Service) ListUserDevices(ctx context.Context, userID, currentDeviceID string) ([]DeviceDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	devs, err := s.deviceRepository.ListActiveByUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to list active devices: %w", err)
	}

	res := make([]DeviceDTO, 0, len(devs))
	for _, d := range devs {
		dto := DeviceDTO{
			ID:        d.ID.String(),
			Name:      d.Name,
			Platform:  d.Platform,
			IsActive:  d.IsActive,
			IsCurrent: currentDeviceID != "" && d.ID.String() == currentDeviceID,
			CreatedAt: d.CreatedAt.Format(time.RFC3339Nano),
			LastSeen:  d.LastSeen.Format(time.RFC3339Nano),
		}
		res = append(res, dto)
	}

	return res, nil
}

func (s *Service) DeactivateDevice(ctx context.Context, userID, deviceID, currentDeviceID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	devID, err := uuid.Parse(deviceID)
	if err != nil {
		return fmt.Errorf("invalid device id: %w", err)
	}

	device, err := s.deviceRepository.GetById(ctx, devID)
	if err != nil {
		return fmt.Errorf("failed to load device: %w", err)
	}
	if device == nil {
		return fmt.Errorf("device not found")
	}

	if device.UserID != uid {
		return fmt.Errorf("device does not belong to this user")
	}

	if err := s.deviceRepository.Deactivate(ctx, device.ID); err != nil {
		return fmt.Errorf("failed to deactivate device: %w", err)
	}

	s.log.Info("device deactivated", zap.String("user_id", device.UserID.String()), zap.String("device_id", device.ID.String()))

	return nil
}

func (s *Service) DeleteDevice(ctx context.Context, userID, deviceID, currentDeviceID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	devID, err := uuid.Parse(deviceID)
	if err != nil {
		return fmt.Errorf("invalid device id: %w", err)
	}

	device, err := s.deviceRepository.GetById(ctx, devID)
	if err != nil {
		return fmt.Errorf("failed to load device: %w", err)
	}

	if device == nil {
		return fmt.Errorf("device not found")
	}

	if device.UserID != uid {
		return fmt.Errorf("device does not belong to this user")
	}

	if currentDeviceID != "" && device.ID.String() == currentDeviceID {
		return fmt.Errorf("device is currently active, so deactivate it first")
	}

	if err := s.deviceRepository.Delete(ctx, device.ID); err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	s.log.Info("device deleted", zap.String("user_id", device.UserID.String()), zap.String("device_id", device.ID.String()))

	return nil
}
