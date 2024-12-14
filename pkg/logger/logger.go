package logger

import (
	"context"
	"fmt"
	"log/slog"
	"morty-smith-34-c/pkg/config"
	"os"
	"time"

	"gorm.io/gorm/logger"
)

// Logger представляет основной логгер для всего приложения, включая поддержку GORM
type Logger struct {
	internalLogger *slog.Logger
	debugMode      bool
	logLevel       logger.LogLevel
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
		logLevel:       logger.Info, // Уровень логирования по умолчанию
	}
}

// LogMode изменяет уровень логирования GORM
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.logLevel = level
	return l
}

// Info логирует сообщение на уровне INFO
func (l *Logger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	if l.logLevel >= logger.Info {
		l.internalLogger.Info(msg, keysAndValues...)
	}
}

// Warn логирует предупреждения
func (l *Logger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	if l.logLevel >= logger.Warn {
		l.internalLogger.Warn(msg, keysAndValues...)
	}
}

// Error логирует ошибки
func (l *Logger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	if l.logLevel >= logger.Error {
		l.internalLogger.Error(msg, keysAndValues...)
	}
}

// Trace логирует информацию о запросах, их длительности и ошибках
func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel == logger.Silent {
		return
	}

	sql, rows := fc()
	elapsed := time.Since(begin)

	switch {
	case err != nil && l.logLevel >= logger.Error:
		l.internalLogger.Error("SQL execution failed",
			"error", err,
			"elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Microseconds())/1000),
			"rows", rows,
			"sql", sql,
		)
	case elapsed > 200*time.Millisecond && l.logLevel >= logger.Warn: // Настройте порог для «медленных» запросов
		l.internalLogger.Warn("Slow SQL query",
			"elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Microseconds())/1000),
			"rows", rows,
			"sql", sql,
		)
	case l.logLevel >= logger.Info:
		l.internalLogger.Info("Executed SQL query",
			"elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Microseconds())/1000),
			"rows", rows,
			"sql", sql,
		)
	}
}

// Debug логирует сообщение на уровне DEBUG
func (l *Logger) Debug(msg string, keysAndValues ...any) {
	if l.debugMode {
		l.internalLogger.Debug(msg, keysAndValues...)
	}
}
