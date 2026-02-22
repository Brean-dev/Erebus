# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Erebus is a **tarpit/honeypot web server** written in Go that traps LLM scrapers and automated crawlers. It generates realistic-looking streaming HTML content to keep bots engaged while tracking their behavior via Redis.

## Commands

```bash
make build    # Compile binary to ./tmp/erebus
make run      # Build and run
make dev      # Start full dev environment (Redis + App) via Docker Compose
make test     # Run all tests: go test -v ./...
make lint     # Run golangci-lint
make deps     # go mod download && go mod tidy
make clean    # Remove ./tmp artifacts
```

For live-reload development (requires [air](https://github.com/cosmtrek/air)):
```bash
air
```

Dev environment uses `dev/docker-compose.redis.yml` for Redis; production uses `docker-compose.yml`.

Required environment variables (see `.env`):
```
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=<password>
```

## Architecture

```
main.go
  → registers /robots.txt, /sitemap.xml, and / (catch-all) handlers
  → initializes Redis client with retry logic (10 attempts, 3s delay)
  → starts HTTP server on :8080 with long WriteTimeout (120s) for streaming
```

### Key Packages

**`internal/pages/`** — Core HTTP handlers and page generation:
- `GenerateHandler` streams chunked HTML to clients. It calls `streamWords` which writes words in random-sized chunks (1–8 words) with configurable delays (`StreamInterval` from `config.toml`, default 6s). Uses `http.Flusher` to push data without buffering. Sets `X-Accel-Buffering: no` to disable Nginx buffering.
- `structure.go` generates page sections, breadcrumbs, sidebars, footers
- `links.go` generates realistic-looking URLs (e.g., `/articles/2023/12/slug`, `/tag/keyword`)
- `meta.go` generates SEO metadata, Open Graph tags, JSON-LD schema
- `robots.go` intentionally disallows bait paths (`/admin/`, `/confidential/`, etc.) to attract aggressive scrapers

**`internal/bable/`** — Markov chain text generator trained on `manifest` and `words_manifesto` files. `NewChain(prefixLen)` → `Build(text)` → `GenerateSentences(n)`.

**`internal/rediscache/`** — IP session tracking:
- Uses CloudFlare header `CF-Connecting-IP` for real IP extraction
- Keys: `trap:active:{ip}` (180s TTL), `trap:first-seen:{ip}`, `trap:last-seen:{ip}` (24h TTL)
- Logs session duration when an IP returns after the active timeout expires

**`internal/erebusconfig/`** — Loads `config.toml` (TOML format). `StreamInterval float64` controls seconds between word chunks.

**`internal/logger/`** — `MultiLogger` writing colored output to stdout and JSON to daily log files in `./logs/app_YYYY-MM-DD.log`. Filters out bot user agents (wget, letsencrypt, etc.).

**`internal/utils/`** — `LogRequest` middleware extracts CloudFlare headers and logs each request with a UUID correlation ID.

## Go Version & Patterns

The project uses **Go 1.25.6** (`go.mod`). Use the new `sync.WaitGroup.Go` method (available since 1.25):
```go
var wg sync.WaitGroup
wg.Go(task1)
wg.Go(task2)
wg.Wait()
```

Use the enhanced `net/http` `ServeMux` with pattern-based routing (Go 1.22+).

## Critical Go Rules (from AGENT.md)

- **Never duplicate `package` declarations.** Each `.go` file must have exactly one `package` line. Before creating or replacing a file, check existing files in the same directory for the package name.
- Accept interfaces, return concrete types. Keep interfaces small (1–3 methods).
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`.
- The `gosec` G404 rule (weak random) is suppressed in `.golangci.yml` — using `math/rand` for content generation is intentional.

## Linting

`.golangci.yml` enables: errcheck, govet, staticcheck, revive, gosec, gocyclo, misspell, and others. Max line length: 120. Min cyclomatic complexity: 15. Run `make lint` before committing; pre-commit hooks also run golangci-lint on modified files.
