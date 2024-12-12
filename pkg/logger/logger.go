package logger

import (
	"log/slog"
	"morty-smith-34-c/pkg/config"
	"os"
)

type Logger struct {
	internalLogger *slog.Logger
	debugMode      bool
}

// NewLogger создаёт новый логгер с уровнем из конфигурации
func NewLogger(cfg *config.Config) *Logger {
	debug := cfg.Debug

	var handler slog.Handler
	if debug {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}

	return &Logger{
		internalLogger: slog.New(handler),
		debugMode:      debug,
	}
}

// Debug логирует сообщение на уровне DEBUG
func (l *Logger) Debug(msg string, keysAndValues ...any) {
	if l.debugMode {
		l.internalLogger.Debug(msg, keysAndValues...)
	}
}

// Info логирует сообщение на уровне INFO
func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.internalLogger.Info(msg, keysAndValues...)
}

// Error логирует сообщение на уровне ERROR
func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.internalLogger.Error(msg, keysAndValues...)
}
