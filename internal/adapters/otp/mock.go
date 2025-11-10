package otp

import (
	"context"

	"github.com/hidden-life/secrum-server/internal/ports"
	"go.uber.org/zap"
)

type MockProvider struct {
	logger *zap.Logger
	env    string
}

func NewMockProvider(logger *zap.Logger, env string) ports.OTPProvider {
	return &MockProvider{
		logger: logger,
		env:    env,
	}
}

func (m *MockProvider) Deliver(_ context.Context, destination, code string) error {
	if m.env != "production" {
		m.logger.Info("Mock OTP delivery", zap.String("destination", destination), zap.String("code", code))
	}
	return nil
}
