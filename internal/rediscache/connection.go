// Package rediscache will do things with redis
package rediscache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

var (
	// Ctx is the global context.Background() var for the Redis connection.
	Ctx = context.Background()
	// Rdb is the gloabal *redis.Client pointer.
	Rdb *redis.Client
)

// DisconnectRedis will close the connection to Redis.
func DisconnectRedis() {
	defer func() {
		err := Rdb.Close()
		if err != nil {
			_ = fmt.Errorf("error closing redis: %w", err)
		}
	}()
}

// ConnectRedis will connect to our Redis instance.
func ConnectRedis() *redis.Client {
	Rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_PORT")),
		Password:     os.Getenv("REDIS_PASSWORD"),
		PoolSize:     10,
		DialTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		DB:           0,
	})

	ping, err := TestRedisConnection(Rdb)
	if err != nil {
		return nil
	} else if ping {
		return Rdb
	}
	return Rdb
}

// TestRedisConnection will ping the redis instance.
func TestRedisConnection(r *redis.Client) (bool, error) {
	_, err := r.Ping(Ctx).Result()
	if err != nil {
		return false, err
	}
	return true, nil
}
