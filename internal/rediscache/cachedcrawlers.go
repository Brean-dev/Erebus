package rediscache

import (
	"log/slog"
	"net/http"
	"time"
)

// ttlSet is the time to live for the set operation.
var ttlSet = time.Duration(180)

// SetIP will write CF-Connecting-IP into memory for 180 seconds(3 minutes)
// of redis. We will use this to track how long scrapers are stuck.
func SetIP(r *http.Request) {
	redis, redisError := NewRedisClient()
	if redisError != nil {
		slog.Error("error creating new redis client: %w", "error", redisError)
	}
	setValueError := redis.Set("real-ip", r.Header.Get("CF-Connecting-IP"),
		ttlSet*time.Second)
	if setValueError != nil {
		slog.Error("error setting value in redis: %w", "error", setValueError)
	}
}

// GetKey returns a key from redis as a string.
func GetKey(key string) (string, error) {
	redis, redisError := NewRedisClient()
	if redisError != nil {
		slog.Error("error creating new redis client: %w", "error", redisError)
		return "", redisError
	}
	getValueResult, getValueError := redis.Get(key)
	if getValueError != nil {
		slog.Error("error getting value from redis: %w", "error", getValueError)
		return "", getValueError
	}
	return getValueResult, nil
}
