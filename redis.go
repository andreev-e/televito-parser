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

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func writeRedisKey(key string, value string) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
	})
	defer rdb.Close()

	ctx := context.Background()

	err := rdb.Set(ctx, key, value, 0).Err()
	return err
}
