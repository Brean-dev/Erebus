package rediscache

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func newClient() *RedisClient {
	redis, redisError := NewRedisClient()
	if redisError != nil {
		slog.Error("error creating new redis client: %w", "error", redisError)
	}
	return redis
}

// ttlSet is the time to live for the set operation.
var ttlSet = time.Duration(180) * time.Second

// SetIP will write CF-Connecting-IP into memory for 180 seconds(3 minutes)
// of redis. We will use this to track how long scrapers are stuck.
func SetIP(r *http.Request) {
	redis := newClient()
	ip := r.Header.Get("CF-Connecting-IP")
	keyName := fmt.Sprintf("real-ip:%s", ip)
	setValueError := redis.Rdb.SetEx(redis.Ctx, keyName, 1, ttlSet).Err()
	if setValueError != nil {
		slog.Error("error setting value in redis: %w", "error", setValueError)
	}
	_ = redis.Close()
}

// GetKey returns a key from redis as a string.
func GetKey(key string) (string, error) {
	redis := newClient()
	getValueResult, getValueError := redis.Get(key)
	if getValueError != nil {
		slog.Error("error getting value from redis: %w", "error", getValueError)
		return "", getValueError
	}
	_ = redis.Close()
	return getValueResult, nil
}

// GetAllConnectedIPs  will get all the values of a key.
func GetAllConnectedIPs() ([]string, error) {
	redis := newClient()
	var cursor uint64
	var ips []string

	for {
		keys, newCursor, err := redis.Rdb.Scan(redis.Ctx, cursor, "real-ip:*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			ip := strings.TrimPrefix(key, "real-ip:")
			ips = append(ips, ip)
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return ips, nil
}
