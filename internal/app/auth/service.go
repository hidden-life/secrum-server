package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"

	"github.com/hidden-life/secrum-server/internal/domain/device"
	"github.com/hidden-life/secrum-server/internal/domain/user"
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

	env string
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
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
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

	if s.env != "production" {
		resp.Code = code
	}

	return resp, nil
}

func (s *Service) VerifyRegistration(ctx context.Context, req VerifyRegistrationRequest) (*VerifyRegistrationResponse, error) {
	if req.RequestID == "" || req.Code == "" {
		return nil, fmt.Errorf("request_id and code are required")
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

	// Register device (for now single default device)
	dev := device.New(u.ID, "default", "unknown")
	if err := s.deviceRepo.Create(ctx, dev); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Issue tokens
	tokens, err := s.tokenIssuer.Generate(ctx, u.ID.String(), dev.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
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
