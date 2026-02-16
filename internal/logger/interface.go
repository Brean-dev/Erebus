// Package logger will be in charge of logging duh...
package logger

import "context"

// Level represents the severity of a log message.
type Level int

// Log severity levels from least to most severe.
const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger defines a structured logging interface with level filtering and field support.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)

	// WithFields returns a new logger with additional context fields
	WithFields(fields ...Field) Logger

	// SetLevel dynamically adjusts the minimum log level
	SetLevel(level Level)
}

// Field represents a key-value pair attached to a log entry.
type Field struct {
	Key   string
	Value interface{}
}
