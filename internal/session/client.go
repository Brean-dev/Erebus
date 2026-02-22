// Package session manages IP session tracking for the tarpit via Redis.
package session

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client stores the client pointer.
type Client struct {
	Rdb *redis.Client
	Ctx context.Context
}

// New creates a new struct Client.
// It retries the connection up to maxRetries times with a delay between attempts,
// allowing time for the Redis container to become reachable on the network.
func New() (*Client, error) {
	const maxRetries = 10
	const retryDelay = 3 * time.Second

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		PoolSize: 10,
	})
	ctx := context.Background()

	var lastErr error
	for i := range maxRetries {
		if err := rdb.Ping(ctx).Err(); err != nil {
			lastErr = err
			fmt.Printf("Redis connection attempt %d/%d failed: %v\n", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}
		return &Client{
			Rdb: rdb,
			Ctx: ctx,
		}, nil
	}

	return nil, fmt.Errorf("failed to connect to Redis after %d attempts: %w", maxRetries, lastErr)
}

// Ping tests the connection with the Redis instance.
func (rc *Client) Ping() error {
	return rc.Rdb.Ping(rc.Ctx).Err()
}

// Get will get the value of key.
func (rc *Client) Get(key string) (string, error) {
	val, err := rc.Rdb.Get(rc.Ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("key '%s' does not exist", key)
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// Set will set the value of key.
func (rc *Client) Set(key string, value any, expiration time.Duration) error {
	return rc.Rdb.Set(rc.Ctx, key, value, expiration).Err()
}

// Delete will delete the value of key.
func (rc *Client) Delete(keys ...string) (int64, error) {
	deleted, err := rc.Rdb.Del(rc.Ctx, keys...).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// Exists will check if key exists.
func (rc *Client) Exists(key string) (bool, error) {
	count, err := rc.Rdb.Exists(rc.Ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close will close the connection with redis.
func (rc *Client) Close() error {
	return rc.Rdb.Close()
}
