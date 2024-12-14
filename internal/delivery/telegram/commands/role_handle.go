package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) RoleHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	args := strings.Fields(msg.Text)
	if len(args) == 0 {
		return
	}
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.ID == msg.ReplyToMessage.MessageThreadID || args[0] != "/role" {
		return
	}
	err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"superadmin"})
	if err != nil {
		return
	}
	if len(args) < 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "О-о-ох, нет! Укажи роль пожалуйста, на меня и так Рик уже ругается...",
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
		})
		return
	}
	roles := []string{
		"user",
		"moder",
		"admin",
		"superadmin",
	}
	validate := false
	for _, r := range roles {
		if r == args[1] {
			validate = true
			break
		}
	}
	if !validate {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "О-о-ох, нет! Я не понимаю о чем ты... Попробуй иначе [user, moder, admin, superadmin]",
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
		})
		return
	}
	exists, err := h.UserUseCase.Exists(ctx, msg.ReplyToMessage.From.ID)
	if err != nil {
		return
	}
	if !exists {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("О\\-ох\\, %s\\, кажется\\, я не знаю кто это\\!", telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	h.UserUseCase.UpdateRole(ctx, msg.ReplyToMessage.From.ID, args[1])
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   "В-в-всё получилось! Наконец-то я могу отдохнуть...",
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
	})
}
