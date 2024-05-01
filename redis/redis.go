package Redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"strings"
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

	var timeString = value.Format("2006-01-02 15:04:05 -0700")
	return redisClient.WriteKey(key, "s:"+strconv.Itoa(len(timeString))+":\""+timeString+"\";")
}

func (rc *RedisClient) ReadTime(key string) (string, error) {
	redisClient := NewRedisClient()
	defer redisClient.Close()

	timestring, err := redisClient.ReadKey(key)

	if err != nil {
		return "", err
	}

	return extractTimeString(timestring)
}

func extractTimeString(input string) (string, error) {
	// Find the start and end indexes of the timeString
	startIndex := strings.Index(input, ":\"") + 2
	endIndex := startIndex + strings.Index(input[startIndex:], "\"")

	// Extract the timeString
	if startIndex < 0 || endIndex < 0 {
		return "", fmt.Errorf("unable to extract timeString")
	}
	timeString := input[startIndex:endIndex]

	// Check if the length matches the length mentioned in the string
	length, err := strconv.Atoi(input[2:strings.Index(input, ":\"")])
	if err != nil {
		return "", err
	}
	if len(timeString) != length {
		return "", fmt.Errorf("length mismatch")
	}

	return timeString, nil
}
