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

// LogRequest wraps an http.Handler to log each incoming request.
func LogRequest(handler http.Handler) http.Handler {
	logMu.Lock()
	openLogFile()
	logMu.Unlock()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logMu.Lock()
		openLogFile()
		logMu.Unlock()
		ctx := r.Context()
		userAgent := r.Header.Get("User-Agent")
		accept := r.Header.Get("Accept")
		lang := r.Header.Get("Accept-Language")
		encoding := r.Header.Get("Accept-Encoding")
		//nolint:canonicalheader //cannot change header
		cfConnectingIP := r.Header.Get("CF-Connecting-IP")

		lowerUA := strings.ToLower(userAgent)
		if userAgent == "" || strings.Contains(lowerUA, "wget") {
			return
		}

		reqLog := log.WithFields(
			logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
			logger.Field{Key: "method", Value: r.Method},
			logger.Field{Key: "remote_path", Value: r.URL.Path},
			logger.Field{Key: "proto", Value: r.Proto},
			logger.Field{Key: "user_agent", Value: userAgent},
			logger.Field{Key: "accept", Value: accept},
			logger.Field{Key: "lang", Value: lang},
			logger.Field{Key: "encoding", Value: encoding},
			logger.Field{Key: "header_len", Value: len(r.Header)},
			logger.Field{Key: "real_ip", Value: cfConnectingIP},
		)
		reqLog.Info(ctx, "")

		handler.ServeHTTP(w, r)
	})
}
