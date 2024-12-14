package postgres

import (
	"context"
	"fmt"
	myLogger "morty-smith-34-c/pkg/logger"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(ctx context.Context, host string, port int, user, password, dbName, sslMode string, log *myLogger.Logger) (*Database, error) {
	// Формируем строку подключения
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode,
	)

	// Настраиваем подключение
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: log.LogMode(logger.Silent), // Уровень логов GORM
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Проверяем соединение
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	log.Info(ctx, "Database connected successfully")
	return &Database{DB: db}, nil
}
