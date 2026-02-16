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

var (
	RedisClient *cache.RedisClient
	redisError  error
)

func main() {
	httpPort := 8080

	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr,
		slogcolor.DefaultOptions)))
	http.HandleFunc("/robots.txt", pages.RobotsHandler)
	http.HandleFunc("/sitemap.xml", pages.SitemapHandler)
	http.HandleFunc("/", pages.GenerateHandler)

	slog.Info("server started on :8080")
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", httpPort),
		Handler:           utils.LogRequest(http.DefaultServeMux),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}
	RedisClient, redisError = cache.NewRedisClient()
	if redisError != nil {
		slog.Error(redisError.Error())
	}
	pong, err := RedisClient.TestRedisConnection()
	if err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("", "ping", pong)
	}

	serveErr := server.ListenAndServe()
	if serveErr != nil {
		slog.Error(serveErr.Error())
	}
}
