package main

import (
	"fmt"
	"html"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

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

func generateHandler(w http.ResponseWriter, r *http.Request) {
	// parse count with same constraints
	wordCount := 500

	// slow/stream flag and delay per word (ms)
	stream := true

	generatedText := markov.BuildChain(wordCount)

	// non-streaming: immediate response (same look-and-feel as before)
	if !stream {
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
        <h1>` + html.EscapeString(requestPath) + `</h1>
    </div>
    <div class="text-1">
        <p>` + html.EscapeString(generatedText) + `</p>
    </div>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, errFprint := fmt.Fprint(w, html)
		if errFprint != nil {
			fmt.Printf("err: %s \n", errFprint)
		}
		return
	}

	// streaming slow response (simulate slow connection)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported by server", http.StatusInternalServerError)
		return
	}

	// Prevent some reverse proxies from buffering (Nginx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Write initial HTML skeleton and flush so the browser starts rendering
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Slow stream</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .text-1 { background-color: #f0f0f0; padding: 20px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="text-1"><p>`)
	flusher.Flush()

	words := strings.Fields(generatedText)

	// Simulate slow connection with variable chunk sizes and delays
	i := 0
	for i < len(words) {
		select {
		case <-r.Context().Done():
			// client disconnected, stop work
			return
		default:
			// Variable chunk size: 1-8 words at a time
			chunkSize := 1 + rand.Intn(8)
			if i+chunkSize > len(words) {
				chunkSize = len(words) - i
			}

			// Send the chunk
			chunk := words[i : i+chunkSize]
			fmt.Fprint(w, html.EscapeString(strings.Join(chunk, " "))+" ")
			flusher.Flush()

			i += chunkSize

			// Variable delay: 20-200ms with occasional longer pauses (300-500ms)
			var delay time.Duration
			if rand.Float32() < 0.15 { // 15% chance of longer pause (network congestion)
				delay = time.Duration(300+rand.Intn(200)) * time.Millisecond
			} else {
				delay = time.Duration(20+rand.Intn(180)) * time.Millisecond
			}

			if i < len(words) {
				time.Sleep(delay)
			}
		}
	}

	// close the HTML
	fmt.Fprint(w, `</p></div></body></html>`)
	flusher.Flush()
}
