package main

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"go-lab/storage/cache/cachekits"

	"github.com/bradfitz/gomemcache/memcache"
)

var keyPrefix = "benchmark:"
var val = strings.Repeat("a", 100*1024)

func BenchmarkRedis_Set(b *testing.B) {
	c, _ := cachekits.NewRedisClient()
	ctx := context.Background()

	b.ResetTimer()
	i := 0
	for b.Loop() {
		key := keyPrefix + strconv.Itoa(i)

		err := c.Set(ctx, key, val, 0).Err()
		if err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkDragonfly_Set(b *testing.B) {
	c, _ := cachekits.NewDragonfly()
	ctx := context.Background()

	b.ResetTimer()
	i := 0
	for b.Loop() {
		key := keyPrefix + strconv.Itoa(i)

		err := c.Set(ctx, key, val, 0).Err()
		if err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkKeyDB_Set(b *testing.B) {
	c, _ := cachekits.NewKeyDB()
	ctx := context.Background()

	b.ResetTimer()
	i := 0
	for b.Loop() {
		key := keyPrefix + strconv.Itoa(i)

		err := c.Set(ctx, key, val, 0).Err()
		if err != nil {
			b.Fatal(err)
		}
		i++
	}
}

func BenchmarkMemcached_Set(b *testing.B) {
	c, _ := cachekits.NewMemcachedClient()

	b.ResetTimer()
	i := 0
	for b.Loop() {
		key := keyPrefix + strconv.Itoa(i)

		err := c.Set(&memcache.Item{
			Key:   key,
			Value: []byte(val),
		})
		if err != nil {
			b.Fatal(err)
		}
		i++
	}
}
