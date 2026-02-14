// Package logger handles the logging of Erebus
// It creates an Encoded JSON object which it then outputs with slog
package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/MatusOllah/slogcolor"
	"log/slog"
	"time"
)

type StandardLogger struct {
	mu          sync.Mutex
	level       Level
	output      io.Writer
	fields      []Field
	encoder     *json.Encoder
	slogHandler slog.Handler
}

func NewStandardLogger(output io.Writer, level Level) *StandardLogger {

	if output == nil {
		output = os.Stdout
	}
	return &StandardLogger{
		level:       level,
		output:      output,
		fields:      make([]Field, 0),
		encoder:     json.NewEncoder(output),
		slogHandler: slogcolor.NewHandler(output, slogcolor.DefaultOptions),
	}
}

func (l *StandardLogger) log(ctx context.Context, level Level, msg string,
	fields ...Field) {
	if level < l.level {
		return // Skipping when the level is below minimum level
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Convert your Level to slog.Level
	var slogLevel slog.Level
	switch level {
	case DebugLevel:
		slogLevel = slog.LevelDebug
	case InfoLevel:
		slogLevel = slog.LevelInfo
	case WarnLevel:
		slogLevel = slog.LevelWarn
	case ErrorLevel:
		slogLevel = slog.LevelError
	}

	// Build attributes
	excludedKeys := map[string]bool{
		"remote_addr": true,
		"remote_path": true,
		"accept":      true,
		"lang":        true,
		"encoding":    true,
		"header_len":  true,
	}

	attrs := []slog.Attr{}
	for _, field := range l.fields {
		if field.Key == "user_agent" {
			strValue, ok := field.Value.(string)
			if !ok {
				continue
			}
			if strValue == "" ||
				strings.Contains(strValue, "www.letsencrypt.org") {
				return
			}
		}
		if !excludedKeys[field.Key] {
			attrs = append(attrs, slog.Any(field.Key, field.Value))
		}
	}
	for _, field := range fields {
		if !excludedKeys[field.Key] {
			attrs = append(attrs, slog.Any(field.Key, field.Value))
		}
	} // Output colored slog to stdout
	record := slog.NewRecord(time.Now(), slogLevel, msg, 0)
	record.AddAttrs(attrs...)
	_ = l.slogHandler.Handle(ctx, record)

	// Output JSON to file
	entry := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     level.String(),
		"message":   msg,
	}
	for _, attr := range attrs {
		entry[attr.Key] = attr.Value
	}
}

func (l *StandardLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, DebugLevel, msg, fields...)
}

func (l *StandardLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, InfoLevel, msg, fields...)
}

func (l *StandardLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, WarnLevel, msg, fields...)
}

func (l *StandardLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, ErrorLevel, msg, fields...)
}

func (l *StandardLogger) SetLevel(level Level) {
	// Lock thread
	l.mu.Lock()
	// Unlock thread at end of function
	l.level = level
}

func (l *StandardLogger) WithFields(fields ...Field) Logger {
	// Locking the variables in this thread so they are immutable
	l.mu.Lock()
	// Unlock the thread at the end of the function
	defer l.mu.Unlock()

	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &StandardLogger{
		level:       l.level,
		output:      l.output,
		fields:      newFields,
		slogHandler: l.slogHandler,
		encoder:     l.encoder,
	}
}
