package utils

import (
	guuid "github.com/google/uuid"

	"Erebus/internal/logger"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var log *logger.MultiLogger

func init() {
	consoleLog := logger.NewStandardLogger(os.Stdout, logger.InfoLevel)
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer func(fs *os.File) {
		if err := fs.Close(); err != nil {
			_ = fmt.Errorf("%s", err)
		}
	}(logFile)

	fileLog := logger.NewStandardLogger(logFile, logger.InfoLevel)
	log = logger.NewMultiLogger(consoleLog, fileLog)

}

func GenerateRequestID() string {
	id := guuid.New()
	return id.String()
}

func LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ua := r.Header.Get("User-Agent")
		secChUa := r.Header.Get("Sec-CH-UA")
		accept := r.Header.Get("Accept")
		lang := r.Header.Get("Accept-Language")
		encoding := r.Header.Get("Accept-Encoding")
		cfConnectingIP := r.Header.Get("CF-Connecting-IP")

		// skip noisy automatic requests
		if r.URL.Path == "/favicon.ico" || r.URL.Path == "/robots.txt" {
			handler.ServeHTTP(w, r)
			return
		}

		lowerUA := strings.ToLower(ua)
		if ua == "" || strings.Contains(lowerUA, "wget") {
			return
		}

		// --- simple scraper detection ---
		isBot := ua == "" ||
			strings.Contains(lowerUA, "curl") ||
			strings.Contains(lowerUA, "python") ||
			strings.Contains(lowerUA, "scrapy") ||
			strings.Contains(lowerUA, "httpclient") ||
			strings.Contains(lowerUA, "go-http-client") ||
			strings.Contains(lowerUA, "httpx") ||
			strings.Contains(lowerUA, "bot") ||
			secChUa == "" ||
			!strings.HasPrefix(ua, "Mozilla/5.0") ||
			len(r.Header) < 5

		tag := ""
		if isBot {
			tag = "[BOT]"
		}
		reqLog := log.WithFields(
			logger.Field{Key: "tag", Value: tag},
			logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
			logger.Field{Key: "method", Value: r.Method},
			logger.Field{Key: "remote_path", Value: r.URL.Path},
			logger.Field{Key: "proto", Value: r.Proto},
			logger.Field{Key: "user_agent", Value: ua},
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
