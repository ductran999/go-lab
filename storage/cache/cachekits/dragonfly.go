package cachekits

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewKeyDB() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6399", // Dragonfly is redis compatible
		Password: "redis",
		DB:       2,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
