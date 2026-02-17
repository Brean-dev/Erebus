// Package rediscache will do things with redis
package rediscache

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

// RedisClient stores the client pointer.
type RedisClient struct {
	Rdb *redis.Client
	Ctx context.Context
}

// NewRedisClient creates a new struct RedisClient.
func NewRedisClient() (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", "localhost",
			os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		PoolSize: 10,
	})
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		Rdb: rdb,
		Ctx: ctx,
	}, nil

}

// TestRedisConnection tests the connection with the Redis instnace.
func (rc *RedisClient) TestRedisConnection() (bool, error) {
	_, err := rc.Rdb.Ping(rc.Ctx).Result()
	if err != nil {
		return false, err
	}
	return true, nil
}

// Get will get the value of key.
func (rc *RedisClient) Get(key string) (string, error) {
	val, err := rc.Rdb.Get(rc.Ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("key '%s' does not exist", key)
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// Set will set the value of key.
func (rc *RedisClient) Set(key string, value any, expiration time.Duration) error {
	return rc.Rdb.Set(rc.Ctx, key, value, expiration).Err()
}

// Delete will delete the value of key.
func (rc *RedisClient) Delete(keys ...string) (int64, error) {
	deleted, err := rc.Rdb.Del(rc.Ctx, keys...).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// Exists will check if key exists.
func (rc *RedisClient) Exists(key string) (bool, error) {
	count, err := rc.Rdb.Exists(rc.Ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close will close the connection with redis.
func (rc *RedisClient) Close() error {
	return rc.Rdb.Close()
}
