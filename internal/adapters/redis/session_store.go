package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	rdb *redis.Client
}

func NewSessionStore(rdb *redis.Client) *SessionStore {
	return &SessionStore{rdb: rdb}
}

func (s *SessionStore) RevokeDevice(ctx context.Context, deviceID string) error {
	return s.rdb.Set(ctx, s.key(deviceID), "1", 0).Err()
}

func (s *SessionStore) IsDeviceRevoked(ctx context.Context, deviceID string) (bool, error) {
	res, err := s.rdb.Exists(ctx, s.key(deviceID)).Result()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}

func (s *SessionStore) key(deviceID string) string {
	return fmt.Sprintf("session:revoked:%s", deviceID)
}
