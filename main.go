// Package main this is the entry point of Erebus
// a tarpit for llm scrapers written in golang
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"Erebus/internal/pages"
	cache "Erebus/internal/rediscache"
	"Erebus/internal/utils"
	"github.com/MatusOllah/slogcolor"
)

func main() {
	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, slogcolor.DefaultOptions)))

	if _, err := cache.NewRedisClient(); err != nil {
		slog.Error("redis connection failed", "err", err)
	} else {
		slog.Info("redis connected")
	}

	http.HandleFunc("/robots.txt", pages.RobotsHandler)
	http.HandleFunc("/sitemap.xml", pages.SitemapHandler)
	http.HandleFunc("/", pages.GenerateHandler)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", 8080),
		Handler:           utils.LogRequest(http.DefaultServeMux),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	slog.Info("server started", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server error", "err", err)
	}
}
