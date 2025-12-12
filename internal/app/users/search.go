package users

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/user"
	"github.com/hidden-life/secrum-server/internal/ports"
)

type SearchResult struct {
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Username    *string `json:"username"`
}

type SearchService struct {
	userRepo ports.UserRepository
}

func NewSearchService(userRepo ports.UserRepository) *SearchService {
	return &SearchService{userRepo: userRepo}
}

func (s *SearchService) Search(ctx context.Context, query string) (*SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	// UUID
	if id, err := uuid.Parse(query); err == nil {
		u, err := s.userRepo.GetByID(ctx, id)
		if err != nil || u == nil {
			return nil, err
		}

		return buildResult(u), nil
	}

	// user nickname like @nick
	if strings.HasPrefix(query, "@") {
		nickname := strings.TrimSpace(query[1:])
		if nickname == "" {
			return nil, nil
		}

		u, err := s.userRepo.GetByUsername(ctx, nickname)
		if err != nil || u == nil {
			return nil, err
		}

		return buildResult(u), nil
	}

	// phone hash
	u, err := s.userRepo.GetByPhoneHash(ctx, query)
	if err != nil || u == nil {
		return nil, err
	}

	return buildResult(u), nil
}

func buildResult(u *user.User) *SearchResult {
	name := ""
	if u.DisplayName != nil {
		name = *u.DisplayName
	}

	return &SearchResult{
		UserID:      u.ID.String(),
		DisplayName: name,
		Username:    u.Username,
	}
}
