package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"morty-smith-34-c/internal/app/repository"
	"morty-smith-34-c/internal/app/usecase"
	"morty-smith-34-c/internal/delivery/telegram"
	"morty-smith-34-c/internal/delivery/telegram/commands"
	"morty-smith-34-c/internal/school"
	"morty-smith-34-c/internal/storage/cache"
	"morty-smith-34-c/internal/storage/postgres"
	"morty-smith-34-c/pkg/config"
	"morty-smith-34-c/pkg/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	// Настраиваем Telegram Bot API
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log := logger.NewLogger("bot.log", cfg)

	// Подключаемся к базе данных
	db, err := postgres.NewDatabase(
		ctx, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode, log,
	)
	if err != nil {
		log.Error(ctx, "Failed to connect to database: %v", err)
		os.Exit(1)
	}

	// Инициализируем репозитории и usecases
	chatRepo := repository.NewPostgresChatRepository(db.DB)
	userRepo := repository.NewPostgresUserRepository(db.DB)
	chatUseCase := usecase.NewChatUseCase(chatRepo)
	userUseCase := usecase.NewUserUseCase(userRepo)

	// Очередь запросов к апи
	apiQueue := school.NewAPIQueue(3, time.Second, nil, log)
	apiQueue.Start(ctx) // Теперь запросики крутятся

	// Инициализируем JWTService
	jwtService := school.NewJWTService(
		cfg.SchoolUsername,
		cfg.SchoolPassword,
		cfg.SchoolTokenURL,   // URL для аутентификации
		cfg.SchoolBaseApiURL, // URL для API запросов
		log,
		apiQueue,
	)

	// Выполняем первичную аутентификацию
	if err := jwtService.Authenticate(ctx); err != nil {
		log.Error(ctx, "Failed to authenticate with School API: %v", err)
		os.Exit(1)
	}

	// Создаем кеш для идов ThreadID
	chatCache := cache.NewChatCache()

	// Загружаем данные из базы в кеш
	if err := chatCache.LoadFromDatabase(ctx, chatUseCase); err != nil {
		log.Error(ctx, "Failed to load chats from database: %v", err)
		os.Exit(1)
	}

	// Создаём обработчики
	userHandler := telegram.NewUserHandler(log, chatUseCase, userUseCase, jwtService)
	commandHandler := commands.NewCommandHandler(log, chatUseCase, userUseCase, chatCache, userHandler)

	botOptions := []bot.Option{
		bot.WithDebugHandler(func(format string, args ...interface{}) {
			log.Debug(ctx, format, args...)
		}),
		bot.WithErrorsHandler(func(err error) {
			log.Error(ctx, err.Error())
		}),
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update == nil || update.Message == nil {
				return
			}
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
				return
			}
		}),
	}

	tgBot, err := bot.New(cfg.BotToken, botOptions...)
	if err != nil {
		log.Error(ctx, "Failed to initialize Telegram bot: %v", err)
		os.Exit(1)
	}

	// Регистрируем команды
	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/morty_come_here", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/morty_id_topic_here", bot.MatchTypeExact, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/morty_rules", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/morty_faq", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/mute", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/unmute", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.HandleCommand(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/faq", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.FaqHandle(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/rules", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.RulesHandle(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/save", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.SaveHandle(ctx, b, update.Message)
	})

	tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/role", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		commandHandler.RoleHandle(ctx, b, update.Message)
	})

	if cfg.Debug {
		tgBot.RegisterHandler(bot.HandlerTypeMessageText, "/test", bot.MatchTypePrefix, func(ctx context.Context, b *bot.Bot, update *models.Update) {
			log.Debug(ctx, "Test cmd")
		})
	}

	// Запускаем бота
	defer func() {
		tgBot.Close(ctx)
		log.Info(ctx, "Bot has stopped gracefully")
	}()
	log.Info(ctx, "Morty Smith 34-C: I'm a live!")
	tgBot.Start(ctx)
}
