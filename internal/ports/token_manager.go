package ports

import "context"

type TokePair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenManager interface {
	Generate(context.Context, string, string) (*TokePair, error)
}
