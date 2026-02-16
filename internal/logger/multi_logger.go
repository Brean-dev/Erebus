package logger

import "context"

// MultiLogger fans out log messages to multiple loggers.
type MultiLogger struct {
	loggers []Logger
}

// NewMultiLogger creates a MultiLogger that dispatches to all provided loggers.
func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{loggers: loggers}
}

// Debug logs a message at debug level to all loggers.
func (m *MultiLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	for _, logger := range m.loggers {
		logger.Debug(ctx, msg, fields...)
	}
}

// Info logs a message at info level to all loggers.
func (m *MultiLogger) Info(ctx context.Context, msg string, fields ...Field) {
	for _, logger := range m.loggers {
		logger.Info(ctx, msg, fields...)
	}
}

func (m *MultiLogger) Error(ctx context.Context, msg string, fields ...Field) {
	for _, logger := range m.loggers {
		logger.Error(ctx, msg, fields...)
	}
}

// Warn logs a message at warn level to all loggers.
func (m *MultiLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	for _, logger := range m.loggers {
		logger.Warn(ctx, msg, fields...)
	}
}

// SetLevel adjusts the minimum log level for all loggers.
func (m *MultiLogger) SetLevel(level Level) {
	for _, logger := range m.loggers {
		logger.SetLevel(level)
	}
}

// WithFields returns a new MultiLogger
// with the given fields attached to every logger.
func (m *MultiLogger) WithFields(fields ...Field) Logger {
	newLoggers := make([]Logger, len(m.loggers))
	for i, logger := range m.loggers {
		newLoggers[i] = logger.WithFields(fields...)
	}

	return &MultiLogger{loggers: newLoggers}
}
