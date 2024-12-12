package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Debug            bool
	BotToken         string
	DBHost           string
	DBPort           int
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	DBMigrationsPath string
	SchoolUsername   string
	SchoolPassword   string
	SchoolTokenURL   string
}

func LoadConfig() (*Config, error) {
	// Загружаем переменные из .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := &Config{}
	var err error

	// Загружаем конфигурацию
	cfg.Debug, err = strconv.ParseBool(getEnv("DEBUG", "false"))
	if err != nil {
		return nil, err
	}

	cfg.BotToken = getEnv("BOT_TOKEN", "")
	cfg.DBHost = getEnv("DB_HOST", "localhost")
	cfg.DBPort, err = strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, err
	}
	cfg.DBUser = getEnv("DB_USER", "postgres")
	cfg.DBPassword = getEnv("DB_PASSWORD", "")
	cfg.DBName = getEnv("DB_NAME", "mortybrain")
	cfg.DBSSLMode = getEnv("DB_SSL_MODE", "disable")
	cfg.DBMigrationsPath = getEnv("DB_MIGRATIONS_PATH", "file://migrations")
	cfg.SchoolUsername = getEnv("SCHOOL_USERNAME", "")
	cfg.SchoolPassword = getEnv("SCHOOL_PASSWORD", "")
	cfg.SchoolTokenURL = getEnv("SCHOOL_TOKEN_URL", "")

	return cfg, nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
