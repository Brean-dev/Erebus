package rediscache

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
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

// ttlSet is the activity timeout for an active session marker.
// If no request is received within this duration, the session is considered expired.
var ttlSet = time.Duration(180) * time.Second

// ttlHistory is how long first-seen and last-seen timestamps persist
// after a session expires, allowing detection of returning IPs.
var ttlHistory = 24 * time.Hour

// SetIP tracks an IP's connection session in Redis.
//
// On first request from an IP, a new session is started by recording the
// first-seen timestamp. On subsequent requests within the same session,
// the last-seen timestamp and active marker TTL are refreshed.
//
// When a returning IP is detected whose previous session has expired,
// the trapped duration (last_seen - first_seen) is logged so an external
// analytics server can aggregate total trapped time per IP.
func SetIP(r *http.Request) {
	rc := newClient()
	if rc == nil {
		return
	}
	defer func() { _ = rc.Close() }()

	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		return
	}

	activeKey := fmt.Sprintf("trap:active:%s", ip)
	firstSeenKey := fmt.Sprintf("trap:first-seen:%s", ip)
	lastSeenKey := fmt.Sprintf("trap:last-seen:%s", ip)
	realIPKey := fmt.Sprintf("real-ip:%s", ip)

	now := time.Now().Unix()
	nowStr := strconv.FormatInt(now, 10)

	active, err := rc.Exists(activeKey)
	if err != nil {
		slog.Error("error checking active session", "error", err)
		return
	}

	if active {
		// Continuing session: refresh active marker and update last-seen.
		pipe := rc.Rdb.TxPipeline()
		pipe.SetEx(rc.Ctx, activeKey, "1", ttlSet)
		pipe.Set(rc.Ctx, lastSeenKey, nowStr, ttlHistory)
		pipe.SetEx(rc.Ctx, realIPKey, "1", ttlSet)
		if _, pipeErr := pipe.Exec(rc.Ctx); pipeErr != nil {
			slog.Error("error refreshing session", "error", pipeErr)
		}
		return
	}

	// No active session. Check for a previous expired session to log its duration.
	readPipe := rc.Rdb.Pipeline()
	firstCmd := readPipe.Get(rc.Ctx, firstSeenKey)
	lastCmd := readPipe.Get(rc.Ctx, lastSeenKey)
	_, _ = readPipe.Exec(rc.Ctx)

	firstSeenStr, firstErr := firstCmd.Result()
	lastSeenStr, lastErr := lastCmd.Result()

	if firstErr == nil && lastErr == nil {
		firstSeen, pErr1 := strconv.ParseInt(firstSeenStr, 10, 64)
		lastSeen, pErr2 := strconv.ParseInt(lastSeenStr, 10, 64)
		if pErr1 == nil && pErr2 == nil && lastSeen >= firstSeen {
			trappedSeconds := lastSeen - firstSeen
			slog.Info("session ended",
				"ip", ip,
				"trapped_seconds", trappedSeconds,
				"first_seen", firstSeen,
				"last_seen", lastSeen,
			)
		}
	}

	// Start a new session.
	writePipe := rc.Rdb.TxPipeline()
	writePipe.Set(rc.Ctx, firstSeenKey, nowStr, ttlHistory)
	writePipe.Set(rc.Ctx, lastSeenKey, nowStr, ttlHistory)
	writePipe.SetEx(rc.Ctx, activeKey, "1", ttlSet)
	writePipe.SetEx(rc.Ctx, realIPKey, "1", ttlSet)
	if _, pipeErr := writePipe.Exec(rc.Ctx); pipeErr != nil {
		slog.Error("error starting new session", "error", pipeErr)
	}
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
