package commands

import (
	"context"
	"log"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) handleMortyFaq(ctx context.Context, b *bot.Bot, msg *models.Message, args []string) {
	if len(args) < 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Я-я-я ничтожество, но-но, прошу дай мне просто записать подсказки...",
		})
		return
	}
	if msg.ReplyToMessage != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Ох нет, не отвечай на сообщение п-п-пожалуйста, Рик ругается...",
		})
		return
	}
	faq := strings.Join(args[1:], " ")
	err := h.ChatUseCase.UpdateFaqLink(ctx, msg.Chat.ID, faq)
	if err != nil {
		log.Printf("Failed to set faq link: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "О-о-о, о нет! Что-то пошло не так, в-в-вот чёрт, Рик опять будет ругаться...",
		})
		return
	}
	h.chatCache.SetFaq(msg.Chat.ID, faq)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            "Ура! Я з-з-записал подсказки для чата...",
	})
}
