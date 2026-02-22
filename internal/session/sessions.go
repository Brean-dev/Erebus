package session

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrInvalidIP indicates the IP address is missing or invalid.
	ErrInvalidIP = errors.New("invalid or missing IP address")
)

// ttlSet is the activity timeout for an active session marker.
// If no request is received within this duration, the session is considered expired.
const ttlSet = 180 * time.Second

// ttlHistory is how long first-seen and last-seen timestamps persist
// after a session expires, allowing detection of returning IPs.
const ttlHistory = 24 * time.Hour

// SetIP tracks an IP's connection session in Redis.
//
// On first request from an IP, a new session is started by recording the
// first-seen timestamp. On subsequent requests within the same session,
// the last-seen timestamp and active marker TTL are refreshed.
//
// When a returning IP is detected whose previous session has expired,
// the trapped duration (last_seen - first_seen) is logged so an external
// analytics server can aggregate total trapped time per IP.
func (c *Client) SetIP(r *http.Request) error {
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		return ErrInvalidIP
	}

	// Validate IP to prevent key injection attacks
	if strings.ContainsAny(ip, " \t\n\r") {
		slog.Warn("invalid IP address format", "ip", ip)
		return ErrInvalidIP
	}

	activeKey := fmt.Sprintf("trap:active:%s", ip)
	firstSeenKey := fmt.Sprintf("trap:first-seen:%s", ip)
	lastSeenKey := fmt.Sprintf("trap:last-seen:%s", ip)
	realIPKey := fmt.Sprintf("real-ip:%s", ip)

	now := time.Now().Unix()
	nowStr := strconv.FormatInt(now, 10)

	active, err := c.Exists(activeKey)
	if err != nil {
		slog.Error("failed to check active session", "ip", ip, "error", err)
		return fmt.Errorf("check active session: %w", err)
	}

	if active {
		// Continuing session: refresh active marker and update last-seen.
		pipe := c.Rdb.TxPipeline()
		pipe.SetEx(c.Ctx, activeKey, "1", ttlSet)
		pipe.Set(c.Ctx, lastSeenKey, nowStr, ttlHistory)
		pipe.SetEx(c.Ctx, realIPKey, "1", ttlSet)

		if _, pipeErr := pipe.Exec(c.Ctx); pipeErr != nil {
			slog.Error("failed to refresh session", "ip", ip, "error", pipeErr)
			return fmt.Errorf("refresh session: %w", pipeErr)
		}
		return nil
	}

	// No active session. Check for a previous expired session to log its duration.
	ctx := c.Ctx
	firstSeenStr, firstErr := c.Rdb.Get(ctx, firstSeenKey).Result()
	lastSeenStr, lastErr := c.Rdb.Get(ctx, lastSeenKey).Result()

	// Log session end if both timestamps exist
	switch {
	case firstErr == nil && lastErr == nil:
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
	case !errors.Is(firstErr, redis.Nil) && firstErr != nil:
		slog.Error("failed to get first_seen", "ip", ip, "error", firstErr)
	case !errors.Is(lastErr, redis.Nil) && lastErr != nil:
		slog.Error("failed to get last_seen", "ip", ip, "error", lastErr)
	}

	// Start a new session.
	writePipe := c.Rdb.TxPipeline()
	writePipe.Set(c.Ctx, firstSeenKey, nowStr, ttlHistory)
	writePipe.Set(c.Ctx, lastSeenKey, nowStr, ttlHistory)
	writePipe.SetEx(c.Ctx, activeKey, "1", ttlSet)
	writePipe.SetEx(c.Ctx, realIPKey, "1", ttlSet)

	if _, pipeErr := writePipe.Exec(c.Ctx); pipeErr != nil {
		slog.Error("failed to start new session", "ip", ip, "error", pipeErr)
		return fmt.Errorf("start new session: %w", pipeErr)
	}

	return nil
}

// GetKey returns a key from redis as a string.
func (c *Client) GetKey(key string) (string, error) {
	if key == "" {
		return "", errors.New("key cannot be empty")
	}

	result, err := c.Get(key)
	if err != nil {
		slog.Error("failed to get value from redis", "key", key, "error", err)
		return "", fmt.Errorf("get key: %w", err)
	}

	return result, nil
}

// GetAllConnectedIPs returns all currently connected IP addresses.
func (c *Client) GetAllConnectedIPs() ([]string, error) {
	var cursor uint64
	var ips []string
	ctx := c.Ctx

	for {
		keys, newCursor, err := c.Rdb.Scan(ctx, cursor, "real-ip:*", 100).Result()
		if err != nil {
			slog.Error("failed to scan redis keys", "error", err)
			return nil, fmt.Errorf("scan keys: %w", err)
		}

		for _, key := range keys {
			ip := strings.TrimPrefix(key, "real-ip:")
			if ip != "" {
				ips = append(ips, ip)
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
	return ips, nil
}
