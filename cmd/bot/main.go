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
	"morty-smith-34-c/pkg/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log := logger.NewLogger(cfg)

	// Подключаемся к базе данных
	db, err := postgres.NewDatabase(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	if err != nil {
		log.Error("Failed to connect to database: %v", err)
	}

	// Инициализируем репозитории и usecases
	chatRepo := repository.NewPostgresChatRepository(db.DB)
	userRepo := repository.NewPostgresUserRepository(db.DB)
	chatUseCase := usecase.NewChatUseCase(chatRepo)
	userUseCase := usecase.NewUserUseCase(userRepo)

	// Инициализируем JWTService
	jwtService := jwtservice.NewJWTService(
		cfg.SchoolUsername,
		cfg.SchoolPassword,
		cfg.SchoolTokenURL,   // URL для аутентификации
		cfg.SchoolBaseApiURL, // URL для API запросов
		nil,
		log,
	)

	// Выполняем первичную аутентификацию
	if err := jwtService.Authenticate(); err != nil {
		log.Error("Failed to authenticate with School API: %v", err)
	}

	// Настраиваем Telegram Bot API
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Создаем кеш для идов ThreadID
	chatCache := cache.NewChatCache()

	// Загружаем данные из базы в кеш
	if err := chatCache.LoadFromDatabase(ctx, chatUseCase); err != nil {
		log.Error("Failed to load chats from database: %v", err)
	}

	// Создаём обработчики
	commandHandler := telegram.NewCommandHandler(chatUseCase, userUseCase, chatCache)
	userHandler := telegram.NewUserHandler(chatUseCase, userUseCase, jwtService)

	botOptions := []bot.Option{
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			threadID, ok := chatCache.GetThreadID(update.Message.Chat.ID)
			if !ok {
				return
			}
			if update.Message != nil && update.Message.LeftChatMember != nil {
				b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.ID,
				})
				return
			}
			if update.Message != nil && update.Message.NewChatMembers != nil {
				userHandler.HandleNewMembers(ctx, b, update.Message, threadID)
				return
			}
			if update.Message != nil && threadID != -1 && update.Message.MessageThreadID == threadID {
				userHandler.HandleNickname(ctx, b, update.Message)
			}
		}),
	}

	tgBot, err := bot.New(cfg.BotToken, botOptions...)
	if err != nil {
		log.Error("Failed to initialize Telegram bot: %v", err)
	}

	// Регистрируем команды
	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/morty_come_here", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/morty_id_topic_here", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	// tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/test", bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
	// 	b.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
	// 		ChatID:    update.Message.Chat.ID, // ID чата
	// 		MessageID: update.Message.ID,      // ID сообщения
	// 		Reaction: []models.ReactionType{
	// 			{
	// 				Type:              models.ReactionTypeTypeEmoji,
	// 				ReactionTypeEmoji: &models.ReactionTypeEmoji{Emoji: "👍"},
	// 			},
	// 		},
	// 	})
	// })

	// Запускаем бота
	log.Info("Bot is running...")
	tgBot.Start(ctx)
}
