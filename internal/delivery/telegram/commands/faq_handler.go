package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) FaqHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	faq, exists := h.chatCache.GetFaq(msg.Chat.ID)
	if !exists {
		// b.SendMessage(ctx, &bot.SendMessageParams{
		// 	ChatID: msg.Chat.ID,
		// 	Text:   fmt.Sprintf("Ох\\, %s\\. Я не могу тебе помочь \\:\\(", telegram.GenerateMention(msg.From)),
		// 	ReplyParameters: &models.ReplyParameters{
		// 		MessageID: msg.ID,
		// 	},
		// 	ParseMode: models.ParseModeMarkdown,
		// })
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("Н\\-н\\-надеюсь\\, %s\\, тебе это поможет\\: %s", telegram.GenerateMention(msg.From), telegram.EscapeMarkdown(faq)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}
