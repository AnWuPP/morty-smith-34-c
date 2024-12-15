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
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.ID == msg.ReplyToMessage.MessageThreadID {
		h.logger.Debug(ctx, "handleUnmute: message is not reply. text: %s | user: %v | chat: %v", msg.Text, msg.From, msg.Chat)
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
		h.logger.Debug(ctx, "handleUnmute: message is not reply. text: %s | user: %v | for: %v | chat: %v | err: %v", msg.Text, msg.From, msg.ReplyToMessage.From, msg.Chat, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("О нет\\! Не могу размутить %s\\, %s\\, помоги мне\\!", telegram.GenerateMention(msg.ReplyToMessage.From), telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	h.logger.Info(ctx, "handleUnmute: unmute. text: %s | user: %v | for: %v | chat: %v", msg.Text, msg.From, msg.ReplyToMessage.From, msg.Chat)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("Фух\\! %s размучен\\, %s\\, ты герой\\!", telegram.GenerateMention(msg.ReplyToMessage.From), telegram.GenerateMention(msg.From)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}
