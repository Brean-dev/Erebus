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
	"Erebus/internal/session"
	"Erebus/internal/utils"
	"github.com/MatusOllah/slogcolor"
)

func main() {
	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, slogcolor.DefaultOptions)))

	rc, err := session.New()
	if err != nil {
		slog.Error("redis connection failed", "err", err)
		os.Exit(1)
	}
	slog.Info("redis connected")

	http.HandleFunc("/robots.txt", pages.RobotsHandler)
	http.HandleFunc("/sitemap.xml", pages.SitemapHandler)
	http.HandleFunc("/", pages.MakeGenerateHandler(rc))

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
