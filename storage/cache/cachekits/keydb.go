package cachekits

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewDragonfly() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6389", // Dragonfly is redis compatible
		DB:   1,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
