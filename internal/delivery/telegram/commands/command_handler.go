package commands

import (
	"context"
	"morty-smith-34-c/internal/app/usecase"
	"morty-smith-34-c/internal/delivery/telegram"
	"morty-smith-34-c/internal/storage/cache"
	"morty-smith-34-c/pkg/logger"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CommandHandler struct {
	ChatUseCase *usecase.ChatUseCase // Логика работы с чатами
	UserUseCase *usecase.UserUseCase // Логика проверки пользователей
	chatCache   *cache.ChatCache
	userHandler *telegram.UserHandler
	logger      *logger.Logger
}

func NewCommandHandler(log *logger.Logger, chatUseCase *usecase.ChatUseCase, userUseCase *usecase.UserUseCase, chatCache *cache.ChatCache, userHandler *telegram.UserHandler) *CommandHandler {
	return &CommandHandler{
		ChatUseCase: chatUseCase,
		UserUseCase: userUseCase,
		chatCache:   chatCache,
		userHandler: userHandler,
		logger:      log,
	}
}

func (h *CommandHandler) HandleCommand(ctx context.Context, b *bot.Bot, msg *models.Message) {
	args := strings.Fields(msg.Text)
	if len(args) == 0 {
		return
	}

	// Проверяем роль пользователя
	if args[0] == "/morty_come_here" || args[0] == "/morty_id_topic_here" || args[0] == "/morty_rules" || args[0] == "/morty_faq" {
		err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"superadmin"})
		if err != nil {
			return
		}
	}
	if args[0] == "/mute" || args[0] == "/unmute" {
		err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"moderator", "admin", "superadmin"})
		if err != nil {
			return
		}
	}

	switch args[0] {
	case "/mute":
		h.handleMute(ctx, b, msg, args)
	case "/unmute":
		h.handleUnmute(ctx, b, msg)
	case "/morty_faq":
		h.handleMortyFaq(ctx, b, msg, args)
	case "/morty_rules":
		h.handleMortyRules(ctx, b, msg, args)
	case "/morty_come_here":
		h.handleMortyComeHere(ctx, b, msg, args)
	case "/morty_id_topic_here":
		h.handleMortyIdTopicHere(ctx, b, msg)
	}
}
