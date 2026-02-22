// Package utils provides shared HTTP middleware and request helpers.
// nolint:revive // package name intentionally generic for utilities
package utils

import (
	"Erebus/internal/logger"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	guuid "github.com/google/uuid"
)

var (
	log            *logger.MultiLogger
	logFile        *os.File
	currentLogDate string
	logMu          sync.Mutex
)

func init() {
	if err := os.Mkdir("logs", 0750); err != nil && !os.IsExist(err) {
		fmt.Fprintf(os.Stderr, "utils: failed to create logs dir: %v\n", err)
	}
}

// GenerateRequestID returns a new UUID v4 string for request tracing.
func GenerateRequestID() string {
	return guuid.New().String()
}

// openLogFile opens a new daily log file if the date has changed.
// Must be called while logMu is held.
func openLogFile() {
	today := time.Now().Format("2006-01-02")
	if today == currentLogDate {
		return
	}

	if logFile != nil {
		_ = logFile.Close()
	}

	currentLogDate = today

	f, err := os.OpenFile(
		fmt.Sprintf("./logs/app_%s.log", today),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "utils: failed to open log file: %v\n", err)
		return
	}

	logFile = f
	log = logger.NewMultiLogger(
		logger.NewStdoutLogger(os.Stdout, logger.InfoLevel),
		logger.NewFileLogger(logFile, logger.InfoLevel),
	)
}

// requestDetails extracts CloudFlare and standard HTTP headers from a request.
func requestDetails(r *http.Request) map[string]string {
	return map[string]string{
		"country":       r.Header.Get("CF-IPCountry"),
		"ray":           r.Header.Get("CF-Ray"),
		"connecting_ip": r.Header.Get("CF-Connecting-IP"),
		"visitor":       r.Header.Get("CF-Visitor"),
		"ipcity":        r.Header.Get("CF-IPCity"),
		"user_agent":    r.Header.Get("User-Agent"),
		"accept":        r.Header.Get("Accept"),
		"lang":          r.Header.Get("Accept-Language"),
		"encoding":      r.Header.Get("Accept-Encoding"),
		"host":          r.Host,
		"method":        r.Method,
		"remote_addr":   r.RemoteAddr,
		"remote_path":   r.URL.Path,
		"proto":         r.Proto,
	}
}

// shouldSkip reports whether the request should be silently dropped
// (e.g. missing User-Agent).
func shouldSkip(userAgent string) bool {
	if userAgent == "" {
		return true
	}
	return strings.Contains(strings.ToLower(userAgent), "wget")
}

// LogRequest wraps an http.Handler to log each incoming request.
func LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logMu.Lock()
		openLogFile()
		logMu.Unlock()

		details := requestDetails(r)

		if shouldSkip(details["user_agent"]) {
			return
		}

		reqLog := log.WithFields(
			logger.Field{Key: "remote_addr", Value: details["remote_addr"]},
			logger.Field{Key: "method", Value: details["method"]},
			logger.Field{Key: "remote_path", Value: details["remote_path"]},
			logger.Field{Key: "proto", Value: details["proto"]},
			logger.Field{Key: "user_agent", Value: details["user_agent"]},
			logger.Field{Key: "accept", Value: details["accept"]},
			logger.Field{Key: "lang", Value: details["lang"]},
			logger.Field{Key: "encoding", Value: details["encoding"]},
			logger.Field{Key: "header_len", Value: len(r.Header)},
			logger.Field{Key: "real_ip", Value: details["connecting_ip"]},
		)
		reqLog.Info(r.Context(), "")

		handler.ServeHTTP(w, r)
	})
}
