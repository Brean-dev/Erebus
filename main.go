// Package main this is the entry point of Erebus
// a tarpit for llm scrapers written in golang
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/MatusOllah/slogcolor"

	"Erebus/internal/pages"
	"Erebus/internal/utils"
)

func main() {
	httpPort := 8080

	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr,
		slogcolor.DefaultOptions)))
	http.HandleFunc("/robots.txt", pages.RobotsHandler)
	http.HandleFunc("/sitemap.xml", pages.SitemapHandler)
	// Catch-all handler that responds to all other endpoints
	http.HandleFunc("/", pages.GenerateHandler)

	slog.Info("server started on :8080")

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", httpPort),
		Handler:           utils.LogRequest(http.DefaultServeMux),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
	}
}
