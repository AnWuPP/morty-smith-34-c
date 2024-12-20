package commands

import (
	"context"
	"morty-smith-34-c/internal/app/entity"
	"morty-smith-34-c/internal/delivery/telegram"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) handleMortyComeHere(ctx context.Context, b *bot.Bot, msg *models.Message, args []string) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if len(args) < 2 {
		h.logger.Debug(ctx, "handlerMortyComeHere: missing campus", "text", msg.Text, "user", telegram.UserForLogger(msg.From), "chat", telegram.ChatForLogger(msg.Chat))
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "О, боже мой! Укажите название кампуса, пожалуйста. Пример: /morty_come_here msk. Это ведь не так сложно, да?",
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
	campusName := args[1]
	err := h.ChatUseCase.Create(ctx, &entity.Chat{
		ChatID:     msg.Chat.ID,
		CampusName: campusName,
	})
	if err != nil {
		h.logger.Error(ctx, "handlerMortyComeHere: create campus error",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
			"err", err,
		)
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Эй, эм... кажется, произошла ошибка! Может, чат уже активирован? О-о-о, я весь в поту!",
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
	h.logger.Debug(ctx, "handlerMortyComeHere: campus created", "text", msg.Text, "user", telegram.UserForLogger(msg.From), "chat", telegram.ChatForLogger(msg.Chat))
	sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            "Эй, поздравляю! Чат успешно активирован для кампуса: " + campusName + ". Ура-а-а!",
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
