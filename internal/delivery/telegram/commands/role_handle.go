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
		h.logger.Debug(ctx, "RoleHandle: no args. how?. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.ReplyToMessage.From, msg.Chat)
		return
	}
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.ID == msg.ReplyToMessage.MessageThreadID || args[0] != "/role" {
		h.logger.Debug(ctx, "RoleHandle: message dont reply. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
		return
	}
	err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"superadmin"})
	if err != nil {
		h.logger.Debug(ctx, "RoleHandle: user not superadmin. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
		return
	}
	if len(args) < 2 {
		h.logger.Debug(ctx, "RoleHandle: missing args. text: %s | user: %v | for: %v | chat: %v", msg.Text, msg.From, msg.ReplyToMessage.From, msg.Chat)
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
		h.logger.Debug(ctx, "RoleHandle: missing role. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
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
		h.logger.Info(ctx, "RoleHandle: not found user. text: %s | user: %v | for: %v | chat: %v", msg.Text, msg.From, msg.ReplyToMessage.From, msg.Chat)
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
	h.logger.Info(ctx, "RoleHandle: update role. text: %s | user: %v | for: %v | chat: %v", msg.Text, msg.From, msg.ReplyToMessage.From, msg.Chat)
	h.UserUseCase.UpdateRole(ctx, msg.ReplyToMessage.From.ID, args[1])
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   "В-в-всё получилось! Наконец-то я могу отдохнуть...",
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
	})
}
