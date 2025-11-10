package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"

	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type Service struct {
	log         *zap.Logger
	otpStore    ports.OTPStore
	otpProvider ports.OTPProvider
}

func NewService(log *zap.Logger, otpStore ports.OTPStore, otpProvider ports.OTPProvider) *Service {
	return &Service{
		log:         log,
		otpStore:    otpStore,
		otpProvider: otpProvider,
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
	Message string `json:"message"`
}

func (s *Service) BeginRegistration(ctx context.Context, req BeginRegistrationRequest, devExposedCode bool) (*BeginRegistrationResponse, error) {
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

	if err := s.otpProvider.Deliver(ctx, phone, code); err != nil {
		s.log.Warn("failed to deliver OTP", zap.Error(err))
	}

	resp := &BeginRegistrationResponse{
		RequestID: requestID,
	}

	if devExposedCode {
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

	s.log.Info("User verified successfully", zap.String("phone_hash", phoneHash))

	// TODO:
	// - Find or create user by phoneHash
	// - Create device
	// - Issuer JWT access/refresh_tokens

	return &VerifyRegistrationResponse{
		Message: "OTP verified successfully",
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
