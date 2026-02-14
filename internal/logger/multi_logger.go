package logger

import "context"

type MultiLogger struct {
	loggers []Logger
}

func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{loggers: loggers}
}

func (m *MultiLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	for _, logger := range m.loggers {
		logger.Debug(ctx, msg, fields...)
	}
}

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

func (m *MultiLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	for _, logger := range m.loggers {
		logger.Warn(ctx, msg, fields...)
	}
}

func (m *MultiLogger) SetLevel(level Level) {
	for _, logger := range m.loggers {
		logger.SetLevel(level)
	}
}

func (m *MultiLogger) WithFields(fields ...Field) Logger {
	newLoggers := make([]Logger, len(m.loggers))
	for i, logger := range m.loggers {
		newLoggers[i] = logger.WithFields(fields...)
	}

	return &MultiLogger{loggers: newLoggers}
}
