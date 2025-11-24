package presence

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	rdb *redis.Client
	ttl time.Duration
}

func New(rdb *redis.Client) *Service {
	return &Service{
		rdb: rdb,
		ttl: time.Minute, // todo: move to configuration
	}
}

func (s *Service) SetOnline(ctx context.Context, userID string) error {
	return s.rdb.Set(ctx, "presence:"+userID, "online", s.ttl).Err()
}

func (s *Service) SetOffline(ctx context.Context, userID string) error {
	return s.rdb.Del(ctx, s.key(userID)).Err()
}

func (s *Service) Refresh(ctx context.Context, userID string) error {
	return s.rdb.Expire(ctx, "presence:"+userID, s.ttl).Err()
}

func (s *Service) IsOnline(ctx context.Context, userID string) (bool, error) {
	res, err := s.rdb.Exists(ctx, s.key(userID)).Result()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}

func (s *Service) key(userID string) string {
	return "presence:" + userID
}
