package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"morty-smith-34-c/internal/app/repository"
	"morty-smith-34-c/internal/app/usecase"
	"morty-smith-34-c/internal/delivery/telegram"
	"morty-smith-34-c/internal/pkg/jwtservice"
	"morty-smith-34-c/internal/storage/cache"
	"morty-smith-34-c/internal/storage/postgres"
	"morty-smith-34-c/pkg/config"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ConsoleLogger struct{}

func (c *ConsoleLogger) Info(msg string) {
	log.Println("[INFO]:", msg)
}

func (c *ConsoleLogger) Error(msg string) {
	log.Println("[ERROR]:", msg)
}

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Подключаемся к базе данных
	db, err := postgres.NewDatabase(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Инициализируем репозитории и usecases
	chatRepo := repository.NewPostgresChatRepository(db.DB)
	userRepo := repository.NewPostgresUserRepository(db.DB)
	chatUseCase := usecase.NewChatUseCase(chatRepo)
	userUseCase := usecase.NewUserUseCase(userRepo)

	// Инициализируем JWTService
	logger := &ConsoleLogger{}
	jwtService := jwtservice.NewJWTService(
		cfg.SchoolUsername,
		cfg.SchoolPassword,
		cfg.SchoolTokenURL, // URL для аутентификации
		"https://edu-api.21-school.ru/services/21-school/api/v1", // URL для API запросов
		nil,
		logger,
	)

	// Выполняем первичную аутентификацию
	if err := jwtService.Authenticate(); err != nil {
		log.Fatalf("Failed to authenticate with School API: %v", err)
	}

	// Настраиваем Telegram Bot API
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Создаем кеш для идов ThreadID
	chatCache := cache.NewChatCache()

	// Загружаем данные из базы в кеш
	if err := chatCache.LoadFromDatabase(ctx, chatUseCase); err != nil {
		log.Fatalf("Failed to load chats from database: %v", err)
	}

	// Создаём обработчики
	commandHandler := telegram.NewCommandHandler(chatUseCase, userUseCase, chatCache)
	userHandler := telegram.NewUserHandler(chatUseCase, userUseCase, jwtService)

	botOptions := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message != nil && update.Message.NewChatMembers != nil {
				userHandler.HandleNewMembers(ctx, b, update.Message)
				return
			}
			threadID, ok := chatCache.GetThreadID(update.Message.Chat.ID)
			if !ok {
				log.Printf("ChatID %d not found in cache", update.Message.Chat.ID)
				return
			}
			if update.Message != nil && threadID != -1 && update.Message.MessageThreadID == threadID {
				userHandler.HandleNickname(ctx, b, update.Message)
			}
		}),
	}

	tgBot, err := bot.New(cfg.BotToken, botOptions...)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Регистрируем команды
	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/morty_come_here", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/morty_id_topic_here", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	// Запускаем бота
	log.Println("Bot is running...")
	tgBot.Start(ctx)
}
