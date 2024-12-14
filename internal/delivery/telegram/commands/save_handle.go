package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) SaveHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.ID == msg.ReplyToMessage.MessageThreadID {
		return
	}
	err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"admin", "superadmin"})
	if err != nil {
		return
	}
	exists, err := h.UserUseCase.Exists(ctx, msg.ReplyToMessage.From.ID)
	if err != nil {
		return
	}
	if exists {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("О\\-ох\\, %s\\, кажется\\, я уже знаю его\\!", telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	err = h.UserUseCase.SaveNickname(ctx, msg.ReplyToMessage.From.ID, fmt.Sprintf("%d", msg.ReplyToMessage.From.ID))
	if err != nil {
		h.logger.Error(ctx, "Error save %d with nick %s", msg.ReplyToMessage.From.ID, msg.ReplyToMessage.Text)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "О-о-о, о нет! Кажется что-то пошло не так...!",
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text: fmt.Sprintf("Э\\, %s\\, всё получилось\\! Я записал %s\\, теперь это наш человек\\!",
			telegram.GenerateMention(msg.From),
			telegram.GenerateMention(msg.ReplyToMessage.From),
		),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
	h.userHandler.RemoveUserFromTimers(ctx, b, msg.Chat.ID, msg.ReplyToMessage.From.ID)
}
