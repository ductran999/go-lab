package fallback

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore keeps the shared last-good copy: L2 behind the
// in-memory L1. TTL carries the staleness bound (expired key =
// no copy), 100ms timeouts keep a sick Redis off the hot path,
// and every error degrades to "miss" — the fallback of fallback
// is the next ladder step, never a new failure.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore dials lazily (first use connects); a dead Redis
// surfaces as misses, never startup failure.
func NewRedisStore(addr string) *RedisStore {
	return &RedisStore{client: redis.NewClient(&redis.Options{Addr: addr})}
}

// Get returns the copy, or false on miss, expiry, timeout, or a
// dead server — all four mean "nothing shared", identically.
func (s *RedisStore) Get(key string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

	defer cancel()

	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}

	return val, true
}

// Set stores with TTL (best effort: a failed write only means the
// next process starts without a shared copy).
func (s *RedisStore) Set(key, val string, ttl time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)

	defer cancel()

	_ = s.client.Set(ctx, key, val, ttl).Err()
}
