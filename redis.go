package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

var commonPrefix = "tvito_database_tvito_cache_:"

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

	val, err := rc.client.Get(ctx, commonPrefix+key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (rc *RedisClient) WriteKey(key string, value string) error {
	ctx := context.Background()

	err := rc.client.Set(ctx, commonPrefix+key, value, 0).Err()
	return err
}

func (rc *RedisClient) DeleteKey(key string) error {
	ctx := context.Background()

	err := rc.client.Del(ctx, commonPrefix+key).Err()
	return err
}

func (rc *RedisClient) WriteTime(key string, value time.Time) error {
	redisClient := NewRedisClient()
	defer redisClient.Close()

	var timeString = value.Format("2006-01-02 15:04:05 -0700 MST")
	return redisClient.WriteKey(key, "s:"+strconv.Itoa(len(timeString))+":\""+timeString+"\";")
}
