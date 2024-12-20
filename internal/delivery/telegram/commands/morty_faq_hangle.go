package commands

import (
	"context"
	"morty-smith-34-c/internal/delivery/telegram"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) handleMortyFaq(ctx context.Context, b *bot.Bot, msg *models.Message, args []string) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if len(args) < 2 {
		h.logger.Debug(ctx, "handleMortyFaq: missing args", "text", msg.Text, "user", telegram.UserForLogger(msg.From), "chat", telegram.ChatForLogger(msg.Chat))
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Я-я-я ничтожество, но-но, прошу дай мне просто записать подсказки...",
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
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.ID != msg.ReplyToMessage.MessageThreadID {
		h.logger.Debug(ctx, "handleMortyFaq: message is reply", "text", msg.Text, "user", telegram.UserForLogger(msg.From), "chat", telegram.ChatForLogger(msg.Chat))
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Ох нет, не отвечай на сообщение п-п-пожалуйста, Рик ругается...",
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
	faq := strings.TrimSpace(strings.TrimPrefix(msg.Text, "/morty_faq"))
	err := h.ChatUseCase.UpdateFaqLink(ctx, msg.Chat.ID, faq)
	if err != nil {
		h.logger.Error(ctx, "handleMortyFaq: update link error",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
			"err", err,
		)
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "О-о-о, о нет! Что-то пошло не так, в-в-вот чёрт, Рик опять будет ругаться...",
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
	h.logger.Debug(ctx, "handleMortyFaq: saved faq", "text", msg.Text, "user", telegram.UserForLogger(msg.From), "chat", telegram.ChatForLogger(msg.Chat))
	h.chatCache.SetFaq(msg.Chat.ID, faq)
	sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            "Ура! Я з-з-записал подсказки для чата...",
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
