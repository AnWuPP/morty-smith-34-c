package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) RulesHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	rules, exists := h.chatCache.GetRules(msg.Chat.ID)
	if !exists {
		h.logger.Debug(ctx, "RulesHandle: rules not exists",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Ой\\-ой\\, %s\\, кажется\\, здесь нет правил\\, анархия\\!", telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	h.logger.Debug(ctx, "RulesHandle: send rules",
		"text", msg.Text,
		"user", telegram.UserForLogger(msg.From),
		"chat", telegram.ChatForLogger(msg.Chat),
	)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("%s\\, вот же правила чата\\:\n%s", telegram.GenerateMention(msg.From), telegram.EscapeMarkdown(rules)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}
