package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) RoleHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	args := strings.Fields(msg.Text)
	if len(args) == 0 {
		h.logger.Debug(ctx, "RoleHandle: no args",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		return
	}
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.ID == msg.ReplyToMessage.MessageThreadID || args[0] != "/role" {
		h.logger.Debug(ctx, "RoleHandle: message dont reply",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		return
	}
	err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"superadmin"})
	if err != nil {
		h.logger.Debug(ctx, "RoleHandle: user not superadmin",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		return
	}
	if len(args) < 2 {
		h.logger.Debug(ctx, "RoleHandle: missing args",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"for", telegram.UserForLogger(msg.ReplyToMessage.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "О-о-ох, нет! Укажи роль пожалуйста, на меня и так Рик уже ругается...",
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
		})
		if err != nil {
			return
		}
		time.AfterFunc(time.Minute, func() {
			b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    sendMsg.Chat.ID,
				MessageID: sendMsg.ID,
			})
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
		h.logger.Debug(ctx, "RoleHandle: missing role",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"for", telegram.UserForLogger(msg.ReplyToMessage.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "О-о-ох, нет! Я не понимаю о чем ты... Попробуй иначе [user, moder, admin, superadmin]",
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
		})
		if err != nil {
			return
		}
		time.AfterFunc(time.Minute, func() {
			b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    sendMsg.Chat.ID,
				MessageID: sendMsg.ID,
			})
		})
		return
	}
	exists, err := h.UserUseCase.Exists(ctx, msg.ReplyToMessage.From.ID)
	if err != nil {
		return
	}
	if !exists {
		h.logger.Info(ctx, "RoleHandle: not found user",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"for", telegram.UserForLogger(msg.ReplyToMessage.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("О\\-ох\\, %s\\, кажется\\, я не знаю кто это\\!", telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			return
		}
		time.AfterFunc(time.Minute, func() {
			b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    sendMsg.Chat.ID,
				MessageID: sendMsg.ID,
			})
		})
		return
	}
	h.logger.Info(ctx, "RoleHandle: set role",
		"text", msg.Text,
		"user", telegram.UserForLogger(msg.From),
		"for", telegram.UserForLogger(msg.ReplyToMessage.From),
		"chat", telegram.ChatForLogger(msg.Chat),
	)
	h.UserUseCase.UpdateRole(ctx, msg.ReplyToMessage.From.ID, args[1])
	sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   "В-в-всё получилось! Наконец-то я могу отдохнуть...",
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
	})
	if err != nil {
		return
	}
	time.AfterFunc(time.Minute, func() {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    sendMsg.Chat.ID,
			MessageID: sendMsg.ID,
		})
	})
}
