package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"erebus/markov"

	"github.com/MatusOllah/slogcolor"
	"github.com/tjarratt/babble"
)

func init() {
	// Initialize with Babbler at startup
	b := babble.NewBabbler()
	b.Separator = " "

	// Use 1000 randomly sampled words from Babbler's dictionary
	if err := markov.InitFromBabbler(b, 2, 1000); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	wordCount := 500
	if count := r.URL.Query().Get("count"); count != "" {
		if c, err := strconv.Atoi(count); err == nil && c > 0 && c <= 100 {
			wordCount = c
		}
	}

	generatedText := markov.BuildChain(wordCount)

	// Display the requested path in the HTML
	requestPath := r.URL.Path
	if requestPath == "" {
		requestPath = "/"
	}

	html := `<!DOCTYPE html>
<html>
<head>
	<title></title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; }
		.header {
			color: #333;
			margin-bottom: 20px;
		}
		.text-1 {
			background-color: #f0f0f0;
			padding: 20px;
			border-radius: 5px;
			margin: 20px 0;
		}
		a { margin: 0 10px; }
	</style>
</head>
<body>
	<div class="header">
		<h1></h1>
	</div>
	<div class="text-1">
		<p>` + generatedText + `</p>
	</div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	_, errFprint := fmt.Fprint(w, html)
	if errFprint != nil {
		fmt.Printf("err: %s \n", errFprint)
	}
}

func main() {
	httpPort := 8080
	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, slogcolor.DefaultOptions)))
	// Use a catch-all handler that responds to all endpoints
	http.HandleFunc("/", generateHandler)

	slog.Info("server started on :8080")
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
