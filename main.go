package main

import (
	"fmt"
	"html"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"Erebus/internal/bable"
)

func main() {
	var httpPort int
	httpPort = 8080

	err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), logRequest(http.DefaultServeMux))
	if err != nil {
		slog.Error(err.Error())
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip noisy automatic requests
		if r.URL.Path == "/favicon.ico" || r.URL.Path == "/robots.txt" {
			handler.ServeHTTP(w, r)
			return
		}

		ua := r.Header.Get("User-Agent")
		secChUa := r.Header.Get("Sec-CH-UA")
		accept := r.Header.Get("Accept")
		lang := r.Header.Get("Accept-Language")
		encoding := r.Header.Get("Accept-Encoding")

		// --- simple scraper detection ---
		isBot := false
		lowerUA := strings.ToLower(ua)
		if ua == "" ||
			strings.Contains(lowerUA, "curl") ||
			strings.Contains(lowerUA, "python") ||
			strings.Contains(lowerUA, "scrapy") ||
			strings.Contains(lowerUA, "httpclient") ||
			strings.Contains(lowerUA, "go-http-client") ||
			strings.Contains(lowerUA, "httpx") ||
			strings.Contains(lowerUA, "bot") ||
			secChUa == "" ||
			!strings.HasPrefix(ua, "Mozilla/5.0") ||
			len(r.Header) < 5 {
			isBot = true
		}

		tag := ""
		if isBot {
			tag = "[BOT]"
		}

		log.Printf("%s %s %s %s %s UA=%q Accept=%q Lang=%q Enc=%q Headers=%d",
			tag,
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			r.Proto,
			ua,
			accept,
			lang,
			encoding,
			len(r.Header),
		)

		handler.ServeHTTP(w, r)
	})
}
