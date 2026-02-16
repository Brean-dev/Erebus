// Package rediscache will do things with redis
package rediscache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
)

var ctx = context.Background()

// ConnectRedis will connect to our Redis instance.
func ConnectRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	defer func() {
		err := rdb.Close()
		if err != nil {
			_ = fmt.Errorf("error closing redis: %w", err)
		}
	}()

	ping, err := TestRedisConnection(rdb)
	if err != nil {
		return nil
	} else if ping {
		return rdb
	}
	return rdb
}

// TestRedisConnection will ping the redis instance.
func TestRedisConnection(r *redis.Client) (bool, error) {
	_, err := r.Ping(ctx).Result()
	if err != nil {
		return false, err
	}
	return true, nil
}
