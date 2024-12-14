package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) RulesHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	rules, exists := h.chatCache.GetRules(msg.Chat.ID)
	if !exists {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Ой\\-ой\\, %s\\, кажется\\, здесь нет правил\\, анархия\\!", generateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("О\\, я нашел\\, %s\\, вот же правила чата\\: %s", generateMention(msg.From), escapeMarkdown(rules)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}

func (h *CommandHandler) FaqHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	faq, exists := h.chatCache.GetFaq(msg.Chat.ID)
	if !exists {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Ох\\, %s\\. Я не могу тебе помочь \\:\\(", generateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("Н\\-н\\-надеюсь\\, %s\\, тебе это поможет\\: %s", generateMention(msg.From), escapeMarkdown(faq)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}

func escapeMarkdown(text string) string {
	specialChars := "_*[]()~`>#+-=|{}.!"
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, string(char), "\\"+string(char))
	}
	return text
}
