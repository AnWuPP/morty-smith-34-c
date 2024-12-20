package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// handleUnmute handles the /unmute command
func (h *CommandHandler) handleUnmute(ctx context.Context, b *bot.Bot, msg *models.Message) {
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.ID == msg.ReplyToMessage.MessageThreadID {
		h.logger.Debug(ctx, "handleUnmute: message is not reply",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		return
	}

	if msg.ReplyToMessage.From.ID == msg.From.ID {
		h.logger.Debug(ctx, "handleMute: unmuted yourself try", "text", msg.Text, "user", telegram.UserForLogger(msg.From), "chat", telegram.ChatForLogger(msg.Chat))
		return
	}

	unmutedUserID := msg.ReplyToMessage.From.ID
	_, err := b.RestrictChatMember(ctx, &bot.RestrictChatMemberParams{
		ChatID: msg.Chat.ID,
		UserID: unmutedUserID,
		Permissions: &models.ChatPermissions{
			CanSendMessages:       true,
			CanSendPolls:          true,
			CanSendOtherMessages:  true,
			CanAddWebPagePreviews: true,
			CanSendPhotos:         true,
			CanSendAudios:         true,
			CanSendVideos:         true,
			CanSendDocuments:      true,
			CanSendVideoNotes:     true,
			CanSendVoiceNotes:     true,
		},
		UntilDate: int(time.Now().Add(time.Second * 30).Unix()),
	})
	if err != nil {
		h.logger.Debug(ctx, "handleUnmute: cant unmute",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"for", telegram.UserForLogger(msg.ReplyToMessage.From),
			"chat", telegram.ChatForLogger(msg.Chat),
			"err", err,
		)
		sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("О нет\\! Не могу размутить %s\\, %s\\, помоги мне\\!", telegram.GenerateMention(msg.ReplyToMessage.From), telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
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

	h.logger.Info(ctx, "handleUnmute: unmute",
		"text", msg.Text,
		"user", telegram.UserForLogger(msg.From),
		"for", telegram.UserForLogger(msg.ReplyToMessage.From),
		"chat", telegram.ChatForLogger(msg.Chat),
	)
	sendMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("Фух\\! %s размучен\\, %s\\, ты - герой\\!", telegram.GenerateMention(msg.ReplyToMessage.From), telegram.GenerateMention(msg.From)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
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
