package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/MatusOllah/slogcolor"

	"Erebus/internal/pages"
	"Erebus/internal/utils"
)

func main() {
	httpPort := 8080

	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, slogcolor.DefaultOptions)))
	// Use a catch-all handler that responds to all endpoints
	http.HandleFunc("/", pages.GenerateHandler)

	slog.Info("server started on :8080")

	err := http.ListenAndServe(fmt.Sprintf(":%d", httpPort), utils.LogRequest(http.DefaultServeMux))
	if err != nil {
		slog.Error(err.Error())
	}
}
