// Package pages will have all the logic for serving the pages
package pages

import (
	"fmt"
	"html"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"Erebus/internal/bable"
)

func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	wordCount := 50
	prefixLen := 5

	link1 := bable.Bable(1, 1)
	link2 := bable.Bable(1, 1)
	link3 := bable.Bable(1, 1)
	link4 := bable.Bable(1, 1)
	generatedText := bable.Bable(wordCount, prefixLen)

	// streaming slow response (simulate slow connection)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported by server",
			http.StatusInternalServerError)
		return
	}

	// Prevent some reverse proxies from buffering (Nginx)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	stringTitle := strings.Fields(generatedText)[0]
	// Write initial HTML skeleton and flush so the browser starts rendering
	_, _ = fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>`+stringTitle+`</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .text-1 { background-color: #f0f0f0; padding: 20px; border-radius: 5px; margin: 20px 0; }
        ul { margin: 20px 0; }
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
			_, _ = fmt.Fprint(w, html.EscapeString(strings.Join(chunk, " "))+" ")
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

	// close the text div and add links
	_, _ = fmt.Fprint(w, `</p></div>
    <ul>
        <li><a href="/`+link1+`">`+link1+`</a></li>
        <li><a href="/`+link2+`">`+link2+`</a></li>
        <li><a href="/`+link3+`">`+link3+`</a></li>
        <li><a href="/`+link4+`">`+link4+`</a></li>
    </ul>
</body></html>`)
	flusher.Flush()
}
