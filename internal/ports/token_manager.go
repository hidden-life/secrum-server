package ports

import "context"

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenManager interface {
	Generate(context.Context, string, string) (*TokenPair, error)
	ValidateAccess(context.Context, string) (userID string, deviceID string, err error)
}
