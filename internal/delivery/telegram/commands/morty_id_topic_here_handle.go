package commands

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) handleMortyIdTopicHere(ctx context.Context, b *bot.Bot, msg *models.Message) {
	err := h.ChatUseCase.UpdateThreadID(ctx, msg.Chat.ID, msg.MessageThreadID)
	if err != nil {
		log.Printf("Failed to set ID topic: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "О-о-о, о нет! Я пытался назначить топик, но что-то пошло не так. Простите, ребята!",
		})
		return
	}
	h.chatCache.SetThreadID(msg.Chat.ID, msg.MessageThreadID)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            "Эй, топик ID успешно установлен! Круто, да? Теперь я могу немного расслабиться...",
	})
}
