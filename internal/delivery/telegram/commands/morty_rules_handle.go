package commands

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) handleMortyRules(ctx context.Context, b *bot.Bot, msg *models.Message, args []string) {
	if len(args) < 2 {
		h.logger.Debug(ctx, "handleMortyRules: missing args. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Я-я-я ничтожество, но-но, прошу дай мне просто записать правила...",
		})
		return
	}
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.ID != msg.ReplyToMessage.MessageThreadID {
		h.logger.Debug(ctx, "handleMortyRules: message is reply. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Ох нет, не отвечай на сообщение п-п-пожалуйста, Рик ругается...",
		})
		return
	}
	rules := strings.Join(args[1:], " ")
	err := h.ChatUseCase.UpdateRulesLink(ctx, msg.Chat.ID, rules)
	if err != nil {
		h.logger.Error(ctx, "handleMortyRules: save rules error. text: %s | user: %v | chat: %v | err: %v", msg.Text, msg.From, msg.Chat, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "О-о-о, о нет! Что-то пошло не так, в-в-вот чёрт, Рик опять будет ругаться...",
		})
		return
	}
	h.logger.Debug(ctx, "handleMortyRules: save rules. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
	h.chatCache.SetRules(msg.Chat.ID, rules)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            "Ура! Я з-з-записал правила чата...",
	})
}
