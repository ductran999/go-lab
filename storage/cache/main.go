package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-lab/storage/cache/cachekits"
	"go-lab/storage/cache/onmem"

	"github.com/bradfitz/gomemcache/memcache"
)

var Expiration = time.Second * 15

func RedisExample() {
	appCtx := context.Background()
	exampleKey := "redis"
	redisClient, err := cachekits.NewRedisClient()
	if err != nil {
		log.Fatalln("init redis connection err", err.Error())
	}

	if err := redisClient.Set(appCtx, exampleKey, "redis-value", Expiration).Err(); err != nil {
		log.Fatalln("redis client connection", err)
	}

	val, err := redisClient.Get(context.Background(), exampleKey).Result()
	if err != nil {
		log.Fatalln("redis get key failed", err)
	}

	fmt.Println("redis cached:", val)
}

func MemcachedExample() {
	mcClient, err := cachekits.NewMemcachedClient()
	if err != nil {
		log.Fatalln("memcached client", err)
	}
	err = mcClient.Set(&memcache.Item{
		Key:        "foo",
		Value:      []byte("bar"),
		Expiration: int32(15),
	})
	if err != nil {
		log.Fatal(err)
	}

	item, err := mcClient.Get("foo")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("memcached cached:", string(item.Value))
}

func DragonflyExample() {
	key := "dragonfly"
	dClient, err := cachekits.NewDragonfly()
	if err != nil {
		log.Fatalln("dragonfly client:", err)
	}

	err = dClient.Set(context.Background(), key, "dragonfly", Expiration).Err()
	if err != nil {
		log.Fatalln("dragonfly set cache failed", err)
	}

	val, err := dClient.Get(context.Background(), key).Result()
	if err != nil {
		log.Fatalln("dragonfly get cache miss", err)
	}
	fmt.Println("dragonfly cached:", val)
}

func KeyDBExample() {
	key := "keyDB"
	dClient, err := cachekits.NewKeyDB()
	if err != nil {
		log.Fatalln("keyDB client:", err)
	}

	err = dClient.Set(context.Background(), key, "keyDB", Expiration).Err()
	if err != nil {
		log.Fatalln("keyDB set cache failed", err)
	}

	val, err := dClient.Get(context.Background(), key).Result()
	if err != nil {
		log.Fatalln("keyDB get cache miss", err)
	}
	fmt.Println("keyDB cached:", val)
}

func GoCacheExample() {
	key := "go-cache"
	c := onmem.MewGoCache()
	c.Set(key, "go cache value", Expiration)

	val, existed := c.Get(key)
	if existed {
		fmt.Println("gocache cached:", val)
	}
}

func RistrettoExample() {
	key := "ristretto"
	c, err := onmem.NewRistretto()
	if err != nil {
		log.Fatalln("init ristretto cached failed", err)
	}
	ok := c.SetWithTTL(key, "ristretto val", 1, Expiration)
	if !ok {
		log.Fatalln("ristretto set key error")
	}
	c.Wait()

	val, exist := c.Get(key)
	if exist {
		fmt.Println("ristretto cached:", val)
	}
}

func main() {
	RedisExample()

	MemcachedExample()

	DragonflyExample()

	KeyDBExample()

	GoCacheExample()

	RistrettoExample()
}
