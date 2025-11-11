package jwt

import (
	"context"
	"time"

	jwtLib "github.com/golang-jwt/jwt/v5"
	"github.com/hidden-life/secrum-server/internal/ports"
)

type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration

	issuer string
}

func NewManager(accessSecret, refreshSecret, issuer string, accessMin, refreshMin int) ports.TokenManager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     time.Duration(accessMin) * time.Minute,
		refreshTTL:    time.Duration(refreshMin) * time.Minute,
		issuer:        issuer,
	}
}

func (m *Manager) Generate(_ context.Context, userID, deviceID string) (*ports.TokenPair, error) {
	now := time.Now().UTC()

	accessToken := jwtLib.NewWithClaims(jwtLib.SigningMethodHS256, jwtLib.MapClaims{
		"sub":    userID,
		"device": deviceID,
		"iss":    m.issuer,
		"iat":    now.Unix(),
		"exp":    now.Add(m.accessTTL).Unix(),
		"type":   "access_token",
	})
	accessStr, err := accessToken.SignedString(m.accessSecret)
	if err != nil {
		return nil, err
	}

	refreshToken := jwtLib.NewWithClaims(jwtLib.SigningMethodHS256, jwtLib.MapClaims{
		"sub":    userID,
		"device": deviceID,
		"iss":    m.issuer,
		"iat":    now.Unix(),
		"exp":    now.Add(m.refreshTTL).Unix(),
		"type":   "refresh_token",
	})
	refreshStr, err := refreshToken.SignedString(m.refreshSecret)
	if err != nil {
		return nil, err
	}

	return &ports.TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}
