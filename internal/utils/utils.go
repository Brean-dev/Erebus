// Package utils provides shared HTTP middleware and request helpers.
// nolint:revive // package name intentionally generic for utilities
package utils

import (
	guuid "github.com/google/uuid"

	"Erebus/internal/logger"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	log            *logger.MultiLogger
	logFile        *os.File
	logFileError   error
	currentLogDate string
	logMu          sync.Mutex
)

func init() {
	dirErr := os.Mkdir("logs", 0750)
	if dirErr != nil {
		_ = fmt.Errorf("%w", dirErr)
	}
}

// GenerateRequestID returns a new UUID v4 string for request tracing.
func GenerateRequestID() string {
	id := guuid.New()
	return id.String()
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
	logFile, logFileError = os.OpenFile(
		fmt.Sprintf("/logs/app_%s.log", today),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if logFileError != nil {
		return
	}

	consoleLog := logger.NewStdoutLogger(os.Stdout, logger.InfoLevel)
	fileLog := logger.NewFileLogger(logFile, logger.InfoLevel)
	log = logger.NewMultiLogger(consoleLog, fileLog)
}

func getCFDetails(r *http.Request) map[string]string {
	return map[string]string{
		"country":      r.Header.Get("CF-IPCountry"),
		"ray":          r.Header.Get("CF-Ray"),
		"conneting_ip": r.Header.Get("CF-Connecting-IP"),
		"visitor":      r.Header.Get("CF-Visitor"),
		"ipcity":       r.Header.Get("CF-IPCity"),
		"user_agent":   r.Header.Get("User-Agent"),
		"accept":       r.Header.Get("Accept"),
		"lang":         r.Header.Get("Accept-Language"),
		"encoding":     r.Header.Get("Accept-Encoding"),
		"host":         r.Host,
		"method":       r.Method,
		"remote_addr":  r.RemoteAddr,
		"remote_path":  r.URL.Path,
		"proto":        r.Proto,
	}
}

// LogRequest wraps an http.Handler to log each incoming request.
func LogRequest(handler http.Handler) http.Handler {
	logMu.Lock()
	openLogFile()
	logMu.Unlock()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fetch header
		header := getCFDetails(r)

		logMu.Lock()
		openLogFile()
		logMu.Unlock()
		ctx := r.Context()

		lowerUA := strings.ToLower(header["user_agent"])
		if header["user_agent"] == "" || strings.Contains(lowerUA, "wget") {
			return
		}

		reqLog := log.WithFields(
			logger.Field{Key: "remote_addr", Value: header["remote_addr"]},
			logger.Field{Key: "method", Value: header["method"]},
			logger.Field{Key: "remote_path", Value: header["remote_path"]},
			logger.Field{Key: "proto", Value: header["proto"]},
			logger.Field{Key: "user_agent", Value: header["user_agent"]},
			logger.Field{Key: "accept", Value: header["accept"]},
			logger.Field{Key: "lang", Value: header["lang"]},
			logger.Field{Key: "encoding", Value: header["encoding"]},
			logger.Field{Key: "header_len", Value: len(r.Header)},
			logger.Field{Key: "real_ip", Value: header["real_ip"]},
		)
		reqLog.Info(ctx, "")

		handler.ServeHTTP(w, r)
	})
}
