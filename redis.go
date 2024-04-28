package main

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       1,
		}),
	}
}

func (rc *RedisClient) Close() {
	rc.client.Close()
}

func (rc *RedisClient) ReadKey(key string) (string, error) {
	ctx := context.Background()

	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (rc *RedisClient) WriteKey(key string, value string) error {
	ctx := context.Background()

	err := rc.client.Set(ctx, key, value, 0).Err()
	return err
}

func (rc *RedisClient) DeleteKey(key string) error {
	ctx := context.Background()

	err := rc.client.Del(ctx, key).Err()
	return err
}
