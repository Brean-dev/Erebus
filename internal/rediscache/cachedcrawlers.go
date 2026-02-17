package rediscache

import (
	"log/slog"
	"net/http"
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
var ttlSet = time.Duration(180)

// SetIP will write CF-Connecting-IP into memory for 180 seconds(3 minutes)
// of redis. We will use this to track how long scrapers are stuck.
func SetIP(r *http.Request) {
	redis := newClient()
	setValueError := redis.Set("real-ip", r.Header.Get("CF-Connecting-IP"),
		ttlSet*time.Second)
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

// GetAllValuesFromKey will get all the values of a key.
func GetAllValuesFromKey(keyName string) {
	redis := newClient()
	ctx := redis.Ctx

	// First, check the type of the key
	keyType, err := redis.Rdb.Type(ctx, keyName).Result()
	if err != nil {
		slog.Error("error getting key type: ", "error", err)
		return
	}

	// Use the appropriate command based on type
	switch keyType {
	case "string":
		val, err := redis.Rdb.Get(ctx, keyName).Result()
		if err != nil {
			slog.Error("error getting string value: ", "error", err)
			return
		}
		slog.Info("value: ", "value", val)

	case "list":
		vals, err := redis.Rdb.LRange(ctx, keyName, 0, -1).Result()
		if err != nil {
			slog.Error("error getting list values: ", "error", err)
			return
		}
		slog.Info("values: ", "values", vals)

	case "set":
		vals, err := redis.Rdb.SMembers(ctx, keyName).Result()
		if err != nil {
			slog.Error("error getting set values: ", "error", err)
			return
		}
		slog.Info("values: ", "values", vals)

	case "zset":
		vals, err := redis.Rdb.ZRange(ctx, keyName, 0, -1).Result()
		if err != nil {
			slog.Error("error getting sorted set values: ", "error", err)
			return
		}
		slog.Info("values: ", "values", vals)

	case "hash":
		result, err := redis.Rdb.HGetAll(ctx, keyName).Result()
		if err != nil {
			slog.Error("error getting hash values: ", "error", err)
			return
		}
		slog.Info("values: ", "values", result)

	case "none":
		slog.Warn("key does not exist", "key", keyName)

	default:
		slog.Error("unknown key type", "type", keyType)
	}

	_ = redis.Close()
}
