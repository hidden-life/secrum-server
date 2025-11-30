package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/device"
	"github.com/hidden-life/secrum-server/internal/domain/user"
	"github.com/hidden-life/secrum-server/internal/pkg/crypto"
	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log         *zap.Logger
	otpStore    ports.OTPStore
	otpSender   ports.OTPProvider
	userRepo    ports.UserRepository
	deviceRepo  ports.DeviceRepository
	tokenIssuer ports.TokenManager
	refreshTTL  time.Duration

	env string
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

func NewService(
	log *zap.Logger,
	env string,
	otpStore ports.OTPStore,
	otpSender ports.OTPProvider,
	userRepo ports.UserRepository,
	deviceRepo ports.DeviceRepository,
	tokenIssuer ports.TokenManager,
) *Service {
	return &Service{
		log:         log,
		otpStore:    otpStore,
		otpSender:   otpSender,
		userRepo:    userRepo,
		deviceRepo:  deviceRepo,
		tokenIssuer: tokenIssuer,
		env:         env,
		refreshTTL:  30 * 24 * time.Hour, // 30 days token lifetime
	}
}

// BeginRegistrationRequest represents the request to begin user registration (OTP flow).
type BeginRegistrationRequest struct {
	Phone string `json:"phone"`
}

// BeginRegistrationResponse represents the response after initiating user registration (OTP flow).
type BeginRegistrationResponse struct {
	RequestID string `json:"request_id"`
	Code      string `json:"code,omitempty"` // Included only in non-production environments
}

// VerifyRegistrationRequest represents the request to verify the OTP code during user registration.
type VerifyRegistrationRequest struct {
	RequestID  string `json:"request_id"`
	Code       string `json:"code"`
	Phone      string `json:"phone"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
	Platform   string `json:"platform,omitempty"`
}

// VerifyRegistrationResponse represents the response after verifying the OTP code during user registration.
type VerifyRegistrationResponse struct {
	Message      string `json:"message"`
	UserID       string `json:"user_id"`
	DeviceID     string `json:"device_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *Service) BeginRegistration(ctx context.Context, req BeginRegistrationRequest) (*BeginRegistrationResponse, error) {
	phone := normalizePhone(req.Phone)
	if phone == "" {
		return nil, fmt.Errorf("invalid phone number: %s", req.Phone)
	}

	code, err := generateOTP(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	requestID, err := s.otpStore.SaveChallenge(ctx, phone, code)
	if err != nil {
		return nil, fmt.Errorf("failed to save OTP challenge: %w", err)
	}

	if err := s.otpSender.Deliver(ctx, phone, code); err != nil {
		s.log.Warn("failed to deliver OTP", zap.Error(err))
	}

	resp := &BeginRegistrationResponse{
		RequestID: requestID,
	}

	if s.env != "prod" {
		resp.Code = code
	}

	return resp, nil
}

func (s *Service) VerifyRegistration(ctx context.Context, req VerifyRegistrationRequest) (*VerifyRegistrationResponse, error) {
	if req.RequestID == "" || req.Code == "" || req.Phone == "" {
		return nil, fmt.Errorf("request_id, code and phone are required fields")
	}

	phoneHash, isValid, err := s.otpStore.VerifyAndConsume(ctx, req.RequestID, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	if !isValid {
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	// Ensure user exists or create new
	u, err := s.userRepo.GetByPhoneHash(ctx, phoneHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if u == nil {
		u = user.New(phoneHash)
		if err := s.userRepo.Create(ctx, u); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		s.log.Info("created new user", zap.String("user_id", u.ID.String()))
	}

	deviceName := req.DeviceName
	if deviceName == "" {
		deviceName = "default"
	}

	platform := strings.ToLower(req.Platform)
	if platform == "" {
		platform = runtime.GOOS
	}

	var dev *device.Device
	if req.DeviceID != "" {
		devUUID, err := uuid.Parse(req.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("invalid device_id: %w", err)
		}

		existing, err := s.deviceRepo.GetById(ctx, devUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get device: %w", err)
		}

		if existing == nil {
			dev = device.New(u.ID, deviceName, platform)
			dev.ID = devUUID

			if err := s.deviceRepo.Create(ctx, dev); err != nil {
				return nil, fmt.Errorf("failed to create device: %w", err)
			}
		} else {
			if existing.UserID != u.ID {
				return nil, fmt.Errorf("device belongs to another user")
			}

			existing.LastSeen = time.Now().UTC()
			existing.Name = deviceName
			existing.Platform = platform

			_ = s.deviceRepo.UpdateLastSeen(ctx, existing.ID, existing.LastSeen)
			dev = existing
		}
	} else {
		dev = device.New(u.ID, deviceName, platform)
		if err := s.deviceRepo.Create(ctx, dev); err != nil {
			return nil, fmt.Errorf("failed to create device: %w", err)
		}
	}

	// Issue tokens
	tokens, err := s.tokenIssuer.Generate(ctx, u.ID.String(), dev.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	hash := crypto.Sha256Hex(tokens.RefreshToken)
	exp := time.Now().UTC().Add(s.refreshTTL)
	if err := s.deviceRepo.UpdateRefreshToken(ctx, dev.ID, hash, &exp); err != nil {
		s.log.Warn("failed to persist refresh token", zap.Error(err))
	}

	s.log.Info("User authenticated", zap.String("user_id", u.ID.String()), zap.String("device_id", dev.ID.String()))

	return &VerifyRegistrationResponse{
		Message:      "User authenticated successfully",
		UserID:       u.ID.String(),
		DeviceID:     dev.ID.String(),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*RefreshResponse, error) {
	if req.RefreshToken == "" {
		return nil, fmt.Errorf("refresh_token is required")
	}

	userID, deviceId, err := s.tokenIssuer.ValidateRefresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	devUUID, _ := uuid.Parse(deviceId)
	dev, err := s.deviceRepo.GetById(ctx, devUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	if dev == nil || !dev.IsActive {
		return nil, fmt.Errorf("device not active or missing")
	}

	want := crypto.Sha256Hex(req.RefreshToken)
	if dev.RefreshTokenHash != want {
		return nil, fmt.Errorf("refresh token mismatch")
	}
	if dev.RefreshTokenExpiresAt != nil && time.Now().UTC().After(*dev.RefreshTokenExpiresAt) {
		return nil, fmt.Errorf("refresh token expired")
	}

	pair, err := s.tokenIssuer.Generate(ctx, userID, deviceId)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	newHash := crypto.Sha256Hex(pair.RefreshToken)
	newExp := time.Now().UTC().Add(s.refreshTTL)
	_ = s.deviceRepo.UpdateRefreshToken(ctx, dev.ID, newHash, &newExp)
	_ = s.deviceRepo.UpdateLastSeen(ctx, devUUID, time.Now().UTC())

	return &RefreshResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, deviceID string) error {
	devUUID, err := uuid.Parse(deviceID)
	if err != nil {
		return fmt.Errorf("failed to parse device UUID: %w", err)
	}

	if err := s.deviceRepo.ClearRefreshToken(ctx, devUUID); err != nil {
		return fmt.Errorf("failed to clear refresh token: %w", err)
	}

	_ = s.deviceRepo.UpdateLastSeen(ctx, devUUID, time.Now().UTC())
	return nil
}

func normalizePhone(phone string) string {
	r := regexp.MustCompile(`\D`)
	digits := r.ReplaceAllString(phone, "")
	if len(digits) < 8 {
		return ""
	}

	if phone[0] != '+' {
		return "+" + digits
	}

	return phone
}

func generateOTP(length int) (string, error) {
	const digits = "0123456789"
	otp := make([]byte, length)
	if _, err := rand.Read(otp); err != nil {
		return "", nil
	}

	for i := 0; i < length; i++ {
		otp[i] = digits[int(otp[i])%len(digits)]
	}

	return string(otp), nil
}
