package commands

import (
	"context"
	"morty-smith-34-c/internal/delivery/telegram"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) handleMortyIdTopicHere(ctx context.Context, b *bot.Bot, msg *models.Message) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	err := h.ChatUseCase.UpdateThreadID(ctx, msg.Chat.ID, msg.MessageThreadID)
	if err != nil {
		h.logger.Error(ctx, "handleMortyIdTopicHere: save id topic error",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
			"err", err,
		)
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "О-о-о, о нет! Я пытался назначить топик, но что-то пошло не так. Простите, ребята!",
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
	h.logger.Debug(ctx, "handleMortyIdTopicHere: errror save topic", "text", msg.Text, "user", telegram.UserForLogger(msg.From), "chat", telegram.ChatForLogger(msg.Chat))
	h.chatCache.SetThreadID(msg.Chat.ID, msg.MessageThreadID)
	sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            "Эй, топик ID успешно установлен! Круто, да? Теперь я могу немного расслабиться...",
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
