package users

import (
	"context"
	"strings"

	"github.com/google/uuid"
	app_context "github.com/hidden-life/secrum-server/internal/app/context"
	"github.com/hidden-life/secrum-server/internal/domain/crypto"
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

	currentUserID := app_context.UserIDFromContext(ctx)

	// UUID
	if id, err := uuid.Parse(query); err == nil {
		if id.String() == currentUserID { // prevent to search current user :)
			return nil, nil
		}

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
	//u, err := s.userRepo.GetByPhoneHash(ctx, query)
	//if err != nil || u == nil {
	//	return nil, err
	//}
	candidates := phoneCandidates(query)
	for _, p := range candidates {
		h := crypto.Hasher(p)
		u, err := s.userRepo.GetByPhoneHash(ctx, h)
		if err != nil {
			continue
		}

		if u != nil {
			return buildResult(u), nil
		}
	}

	return nil, nil
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

func phoneCandidates(input string) []string {
	s := strings.TrimSpace(input)
	if s == "" {
		return nil
	}

	// leave only + and digits
	b := strings.Builder{}
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}

		if r == '+' && b.Len() == 0 {
			b.WriteRune(r)
		}
	}

	s = b.String()
	if s == "" {
		return nil
	}

	if strings.HasPrefix(s, "00") {
		s = "+" + strings.TrimPrefix(s, "00")
	}

	unique := make(map[string]struct{})
	add := func(x string) {
		x = strings.TrimSpace(x)
		if x == "" {
			return
		}

		if _, ok := unique[x]; ok {
			return
		}

		unique[x] = struct{}{}
	}

	add(s)

	if !strings.HasPrefix(s, "+") {
		add("+" + s)
	} else {
		add(strings.TrimPrefix(s, "+"))
	}

	out := make([]string, 0, len(unique))
	for k := range unique {
		out = append(out, k)
	}

	return out
}
