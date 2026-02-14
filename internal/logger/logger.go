package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

type StandardLogger struct {
	mu      sync.Mutex
	level   Level
	output  io.Writer
	fields  []Field
	encoder *json.Encoder
}

func NewStandardLogger(output io.Writer, level Level) *StandardLogger {
	if output == nil {
		output = os.Stdout
	}
	return &StandardLogger{
		level:   level,
		output:  output,
		fields:  make([]Field, 0),
		encoder: json.NewEncoder(output),
	}
}

func (l *StandardLogger) log(ctx context.Context, level Level, msg string,
	fields ...Field) {
	if level < l.level {
		return // Skipping when the level is below minimum level
	}

	// Locking the variable so other threads can not touch it
	l.mu.Lock()
	// defer Unlocking to when the function has completed, defer wil
	// Unlock it before exiting the function
	defer l.mu.Unlock()

	//Bulding the log
	entry := make(map[string]interface{})
	entry["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	entry["level"] = level.String()
	entry["message"] = msg

	// Add logger-level fields
	for _, field := range l.fields {
		entry[field.Key] = field.Value
	}

	// Add call-specific fields
	for _, field := range fields {
		entry[field.Key] = field.Value
	}

	// Extract values from context
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry["trace_id"] = traceID
	}

	if requestID := ctx.Value("request_id"); requestID != nil {
		entry["request_id"] = requestID
	}

	// JSON encode the entry
	_ = l.encoder.Encode(entry)
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
		level:   l.level,
		output:  l.output,
		fields:  newFields,
		encoder: l.encoder,
	}
}
