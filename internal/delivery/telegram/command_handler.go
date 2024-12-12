package telegram

import (
	"context"
	"log"
	"morty-smith-34-c/internal/app/entity"
	"morty-smith-34-c/internal/app/usecase"
	"morty-smith-34-c/internal/storage/cache"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CommandHandler struct {
	ChatUseCase *usecase.ChatUseCase // Логика работы с чатами
	UserUseCase *usecase.UserUseCase // Логика проверки пользователей
	chatCache   *cache.ChatCache
}

func NewCommandHandler(chatUseCase *usecase.ChatUseCase, userUseCase *usecase.UserUseCase, chatCache *cache.ChatCache) *CommandHandler {
	return &CommandHandler{
		ChatUseCase: chatUseCase,
		UserUseCase: userUseCase,
		chatCache:   chatCache,
	}
}

func (h *CommandHandler) HandleCommand(ctx context.Context, b *bot.Bot, msg *models.Message) {
	args := strings.Fields(msg.Text)
	if len(args) == 0 {
		return
	}

	// Проверяем роль пользователя
	if args[0] == "/morty_come_here" || args[0] == "/morty_id_topic_here" {
		err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"superadmin"})
		if err != nil {
			return
		}
	}

	switch args[0] {
	case "/morty_come_here":
		if len(args) < 2 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "О, боже мой! Укажите название кампуса, пожалуйста. Пример: /morty_come_here msk. Это ведь не так сложно, да?",
			})
			return
		}
		campusName := args[1]
		err := h.ChatUseCase.Create(ctx, &entity.Chat{
			ChatID:     msg.Chat.ID,
			CampusName: campusName,
		})
		if err != nil {
			log.Printf("Failed to activate chat: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "Эй, эм... кажется, произошла ошибка! Может, чат уже активирован? О-о-о, я весь в поту!",
			})
			return
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Эй, поздравляю! Чат успешно активирован для кампуса: " + campusName + ". Ура-а-а!",
		})

	case "/morty_id_topic_here":
		err := h.ChatUseCase.UpdateThreadID(ctx, msg.Chat.ID, msg.MessageThreadID)
		if err != nil {
			log.Printf("Failed to set ID topic: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "О-о-о, о нет! Я пытался назначить топик, но что-то пошло не так. Простите, ребята!",
			})
			return
		}
		h.chatCache.SetThreadID(msg.Chat.ID, msg.MessageThreadID)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Эй, топик ID успешно установлен! Круто, да? Теперь я могу немного расслабиться...",
		})
	}
}
