package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// handleMute handles the /mute command
func (h *CommandHandler) handleMute(ctx context.Context, b *bot.Bot, msg *models.Message, args []string) {
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.ID == msg.ReplyToMessage.MessageThreadID {
		h.logger.Debug(ctx, "handleMute: message is not reply. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
		return
	}

	duration, err := parseDuration(strings.Join(args, " "))
	if err != nil {
		h.logger.Debug(ctx, "handleMute: missing format time. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Ой\\-ой\\, %s\\, кажется\\, ты что\\-то напутал с форматом времени\\!", telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	if duration < 5*time.Minute {
		h.logger.Debug(ctx, "handleMute: time less 5 minute. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Минимальное время мута — 5 минут\\, как будто у нас есть время на меньшее, %s\\!", telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	mutedUserID := msg.ReplyToMessage.From.ID
	until := time.Now().Add(duration).Unix()
	_, err = b.RestrictChatMember(ctx, &bot.RestrictChatMemberParams{
		ChatID: msg.Chat.ID,
		UserID: mutedUserID,
		Permissions: &models.ChatPermissions{
			CanSendMessages:       false,
			CanSendPolls:          false,
			CanSendOtherMessages:  false,
			CanAddWebPagePreviews: false,
			CanSendPhotos:         false,
			CanSendAudios:         false,
			CanSendVideos:         false,
			CanSendDocuments:      false,
			CanSendVideoNotes:     false,
			CanSendVoiceNotes:     false,
		},
		UntilDate: int(until),
	})
	if err != nil {
		h.logger.Debug(ctx, "handleMute: cant mute. text: %s | user: %v | for: %v | chat: %v | debug: %v", msg.Text, msg.From, msg.ReplyToMessage.From, msg.Chat, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Ох\\, замутить %s не удалось\\, давай попробуем снова\\, %s\\!", telegram.GenerateMention(msg.ReplyToMessage.From), telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	h.logger.Info(ctx, "handleMute: mute. text: %s | user: %v | for: %v | chat: %v", msg.Text, msg.From, msg.ReplyToMessage.From, msg.Chat)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("Бум\\! %s теперь в муте на %s\\, %s\\!", telegram.GenerateMention(msg.ReplyToMessage.From), strings.Join(args[1:], " "), telegram.GenerateMention(msg.From)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}
