# Загружаем переменные из .env файла
include .env
export $(shell sed 's/=.*//' .env)

# Указываем путь до Go
GO := go

# Переменные для миграций
MIGRATE := migrate
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)
MIGRATIONS_DIR := migrations

# Запуск бота
.PHONY: run
run:
	@echo "Запускаем бота..."
	$(GO) run cmd/bot/main.go

# Создание новой миграции
.PHONY: migration-create
migration-create:
ifndef name
	@echo "Укажите имя миграции с параметром name="; exit 1
endif
	@echo "Создаём миграцию $(name)..."
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Применение миграций
.PHONY: migrate-up
migrate-up:
	@echo "Применяем миграции..."
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) up

# Откат миграций
.PHONY: migrate-down
migrate-down:
	@echo "Откатываем миграции..."
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) down

# Полный откат миграций
.PHONY: migrate-reset
migrate-reset:
	@echo "Полный откат миграций..."
	$(MIGRATE) -database "$(DB_URL)" -path $(MIGRATIONS_DIR) drop -f

# Помощь
.PHONY: help
help:
	@echo "Доступные команды:"
	@echo "  run                - Запуск бота"
	@echo "  migration-create   - Создание новой миграции (пример: make migration-create name=create_users_table)"
	@echo "  migrate-up         - Применение всех миграций"
	@echo "  migrate-down       - Откат последней миграции"
	@echo "  migrate-reset      - Полный откат миграций"
