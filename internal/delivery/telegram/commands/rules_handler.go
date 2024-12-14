package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) RulesHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	rules, exists := h.chatCache.GetRules(msg.Chat.ID)
	if !exists {
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
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("О\\, я нашел\\, %s\\, вот же правила чата\\: %s", telegram.GenerateMention(msg.From), telegram.EscapeMarkdown(rules)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}
