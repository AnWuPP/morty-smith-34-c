package commands

import (
	"context"
	"morty-smith-34-c/internal/app/entity"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) handleMortyComeHere(ctx context.Context, b *bot.Bot, msg *models.Message, args []string) {
	if len(args) < 2 {
		h.logger.Debug(ctx, "handlerMortyComeHere: missing campus. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "О, боже мой! Укажите название кампуса, пожалуйста. Пример: /morty_come_here msk. Это ведь не так сложно, да?",
		})
		return
	}
	campusName := args[1]
	err := h.ChatUseCase.Create(ctx, &entity.Chat{
		ChatID:     msg.Chat.ID,
		CampusName: campusName,
	})
	if err != nil {
		h.logger.Error(ctx, "handlerMortyComeHere: create campus erre. text: %s | user: %v | chat: %v | err: %v", msg.Text, msg.From, msg.Chat, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Эй, эм... кажется, произошла ошибка! Может, чат уже активирован? О-о-о, я весь в поту!",
		})
		return
	}
	h.logger.Debug(ctx, "handlerMortyComeHere: campus created. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            "Эй, поздравляю! Чат успешно активирован для кампуса: " + campusName + ". Ура-а-а!",
	})
}
