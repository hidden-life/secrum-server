package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hidden-life/secrum-server/internal/domain/crypto"
	"github.com/hidden-life/secrum-server/internal/ports"
	"github.com/redis/go-redis/v9"
)

const (
	otpKeyPrefix   = "otp:req:"
	otpTTL         = 5 * time.Minute // 5 minutes
	otpMaxAttempts = 3
)

// OTPStoreRedis implements the ports.OTPStore interface using Redis as the backend.
type OTPStoreRedis struct {
	rdb *redis.Client
}

// New creates a new instance of OTPStoreRedis.
func New(rdb *redis.Client) ports.OTPStore {
	return &OTPStoreRedis{rdb: rdb}
}

// SaveChallenge saves an OTP challenge associated with a phone number and returns a request ID.
func (s *OTPStoreRedis) SaveChallenge(ctx context.Context, phone, code string) (string, error) {
	requestID := uuid.New().String()
	key := otpKeyPrefix + requestID

	fields := map[string]interface{}{
		"phone_hash": crypto.Hasher(phone),
		"code_hash":  crypto.Hasher(code),
		"attempts":   0,
	}

	if err := s.rdb.HSet(ctx, key, fields).Err(); err != nil {
		return "", err
	}

	if err := s.rdb.Expire(ctx, key, otpTTL).Err(); err != nil {
		return "", err
	}

	return requestID, nil
}

// VerifyAndConsume verifies the OTP code for the given request ID and consumes it if valid.
func (s *OTPStoreRedis) VerifyAndConsume(ctx context.Context, requestID, code string) (string, bool, error) {
	key := otpKeyPrefix + requestID
	exists, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return "", false, err
	}

	if exists == 0 {
		return "", false, nil // Not found or expired
	}

	attempts, err := s.rdb.HIncrBy(ctx, key, "attempts", 1).Result()
	if err != nil {
		return "", false, err
	}

	if attempts > otpMaxAttempts {
		_ = s.rdb.Del(ctx, key).Err()
		return "", false, nil // Exceeded max attempts
	}

	values, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return "", false, err
	}

	if len(values) == 0 {
		return "", false, nil // Not found
	}

	if values["code_hash"] != crypto.Hasher(code) {
		return "", false, nil // Incorrect code
	}

	phoneHash := values["phone_hash"]
	_ = s.rdb.Del(ctx, key).Err() // Consume OTP

	return phoneHash, true, nil
}
