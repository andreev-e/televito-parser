package main

import (
	"context"
	"github.com/go-redis/redis/v8"
)

func readRedisKey(key string) (string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
	})
	defer rdb.Close()

	ctx := context.Background()

	// Get the value associated with the key.
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}
