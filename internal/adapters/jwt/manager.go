package jwt

import (
	"context"
	"errors"
	"fmt"
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

// ValidateAccess parses and validates access token, returns userID and deviceID
func (m *Manager) ValidateAccess(_ context.Context, accessToken string) (userID string, deviceID string, err error) {
	token, err := jwtLib.Parse(accessToken, func(token *jwtLib.Token) (interface{}, error) {
		if token.Method != jwtLib.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return m.accessSecret, nil
	})

	if err != nil || !token.Valid {
		return "", "", errors.New("invalid access token")
	}

	claims, isOk := token.Claims.(jwtLib.MapClaims)
	if !isOk {
		return "", "", errors.New("invalid claims")
	}

	userID, _ = claims["sub"].(string)
	deviceID, _ = claims["device"].(string)
	if userID == "" || deviceID == "" {
		return "", "", errors.New("missing required claims")
	}

	return userID, deviceID, nil
}
