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

// StdoutLogger writes colored log output to stdout using slog.
type StdoutLogger struct {
	mu          sync.Mutex
	level       Level
	output      io.Writer
	fields      []Field
	encoder     *json.Encoder
	slogHandler slog.Handler
}

// FileLogger writes JSON-formatted log output to a file.
type FileLogger struct {
	mu      sync.Mutex
	level   Level
	output  io.Writer
	fields  []Field
	encoder *json.Encoder
}

// NewStdoutLogger creates a StdoutLogger
// that writes colored output to the given writer.
func NewStdoutLogger(output io.Writer, level Level) *StdoutLogger {

	if output == nil {
		output = os.Stdout
	}
	return &StdoutLogger{
		level:       level,
		output:      output,
		fields:      make([]Field, 0),
		encoder:     json.NewEncoder(output),
		slogHandler: slogcolor.NewHandler(output, slogcolor.DefaultOptions),
	}
}

// NewFileLogger creates a FileLogger that
// writes JSON output to the given writer.
func NewFileLogger(output io.Writer, level Level) *FileLogger {

	return &FileLogger{
		level:   level,
		output:  output,
		fields:  make([]Field, 0),
		encoder: json.NewEncoder(output),
	}
}

// StdoutLogger methods - implements Logger interface with colored output

func (l *StdoutLogger) log(ctx context.Context, level Level,
	msg string, fields ...Field) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Convert Level to slog.Level
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

	// Build attributes, excluding noisy fields
	excludedKeys := map[string]bool{
		"remote_addr": true,
		"remote_path": true,
		"accept":      true,
		"lang":        true,
		"encoding":    true,
		"header_len":  true,
	}

	attrs := []slog.Attr{}
	allFields := make([]Field, 0, len(l.fields)+len(fields))
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)
	for _, field := range allFields {
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

	// Output colored slog to stdout
	record := slog.NewRecord(time.Now(), slogLevel, msg, 0)
	record.AddAttrs(attrs...)
	_ = l.slogHandler.Handle(ctx, record)
}

// Debug logs a message at debug level.
func (l *StdoutLogger) Debug(ctx context.Context,
	msg string, fields ...Field) {
	l.log(ctx, DebugLevel, msg, fields...)
}

// Info logs a message at info level.
func (l *StdoutLogger) Info(ctx context.Context,
	msg string, fields ...Field) {
	l.log(ctx, InfoLevel, msg, fields...)
}

// Warn logs a message at warn level.
func (l *StdoutLogger) Warn(ctx context.Context,
	msg string, fields ...Field) {
	l.log(ctx, WarnLevel, msg, fields...)
}

func (l *StdoutLogger) Error(ctx context.Context,
	msg string, fields ...Field) {
	l.log(ctx, ErrorLevel, msg, fields...)
}

// SetLevel adjusts the minimum log level for this logger.
func (l *StdoutLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// WithFields returns a new logger
// with the given fields attached to every entry.
func (l *StdoutLogger) WithFields(fields ...Field) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &StdoutLogger{
		level:       l.level,
		output:      l.output,
		fields:      newFields,
		encoder:     l.encoder,
		slogHandler: l.slogHandler,
	}
}

// FileLogger methods - implements Logger interface with JSON output

func (l *FileLogger) log(_ context.Context, level Level,
	msg string, fields ...Field) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Build entry with all fields
	entry := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"level":     level.String(),
		"message":   msg,
	}

	allFields := make([]Field, 0, len(l.fields)+len(fields))
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)

	for _, field := range allFields {
		if field.Key == "user_agent" {
			strValue, ok := field.Value.(string)
			if !ok {
				continue
			}
			if strValue == "" || strings.Contains(strValue, "www.l") {
				return
			}
		}
	}
	for _, field := range allFields {
		entry[field.Key] = field.Value
	}

	// Write JSON to file
	_ = l.encoder.Encode(entry)
}

// Debug logs a message at debug level.
func (l *FileLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, DebugLevel, msg, fields...)
}

// Info logs a message at info level.
func (l *FileLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, InfoLevel, msg, fields...)
}

// Warn logs a message at warn level.
func (l *FileLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, WarnLevel, msg, fields...)
}

// Error logs a message at error level.
func (l *FileLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, ErrorLevel, msg, fields...)
}

// SetLevel adjusts the minimum log level for this logger.
func (l *FileLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// WithFields returns a new logger
// with the given fields attached to every entry.
func (l *FileLogger) WithFields(fields ...Field) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &FileLogger{
		level:   l.level,
		output:  l.output,
		fields:  newFields,
		encoder: l.encoder,
	}
}
