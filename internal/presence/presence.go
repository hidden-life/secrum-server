package presence

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Service {
	return &Service{rdb: rdb}
}

func (s *Service) SetOnline(ctx context.Context, userID string) error {
	return s.rdb.Set(ctx, "presence:"+userID, "online", time.Minute).Err()
}

func (s *Service) SetOffline(ctx context.Context, userID string) error {
	return s.rdb.Set(ctx, "presence:"+userID, "offline", time.Minute).Err()
}

func (s *Service) Refresh(ctx context.Context, userID string) error {
	return s.rdb.Expire(ctx, "presence:"+userID, time.Minute).Err()
}
