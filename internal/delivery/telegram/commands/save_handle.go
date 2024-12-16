package commands

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/delivery/telegram"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (h *CommandHandler) SaveHandle(ctx context.Context, b *bot.Bot, msg *models.Message) {
	args := strings.Fields(msg.Text)
	if len(args) == 0 {
		return
	}
	if msg.ReplyToMessage == nil || msg.ReplyToMessage.ID == msg.ReplyToMessage.MessageThreadID {
		h.logger.Debug(ctx, "SaveHandle: message is not reply",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		return
	}
	err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"admin", "superadmin"})
	if err != nil {
		h.logger.Debug(ctx, "SaveHandle: user is not admin or superadmin",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"for", telegram.UserForLogger(msg.ReplyToMessage.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		return
	}
	var schoolNick string
	if len(args) > 1 {
		schoolNick = strings.TrimSpace(strings.TrimPrefix(msg.Text, "/save"))
	} else {
		schoolNick = fmt.Sprintf("%d", msg.ReplyToMessage.From.ID)
	}
	exists, err := h.UserUseCase.Exists(ctx, msg.ReplyToMessage.From.ID)
	if err != nil {
		h.logger.Error(ctx, "SaveHandle: check user exists",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"for", telegram.UserForLogger(msg.ReplyToMessage.From),
			"chat", telegram.ChatForLogger(msg.Chat),
			"err", err,
		)
		return
	}
	if exists {
		rename, err := h.UserUseCase.UpdateSchoolNick(ctx, msg.ReplyToMessage.From.ID, schoolNick)
		if err != nil {
			h.logger.Error(ctx, "SaveHandle: update nick error",
				"text", msg.Text,
				"user", telegram.UserForLogger(msg.From),
				"for", telegram.UserForLogger(msg.ReplyToMessage.From),
				"chat", telegram.ChatForLogger(msg.Chat),
				"err", err,
			)
			return
		}
		if rename {
			h.logger.Debug(ctx, "SaveHandle: rename user",
				"text", msg.Text,
				"user", telegram.UserForLogger(msg.From),
				"for", telegram.UserForLogger(msg.ReplyToMessage.From),
				"chat", telegram.ChatForLogger(msg.Chat),
			)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text: fmt.Sprintf(
					"О\\-ох\\, %s\\, кажется\\, я уже знаю его\\, но не под этим именем\\, теперь я запомнил [%s](tg://user?id=%d)\\!",
					telegram.GenerateMention(msg.From),
					schoolNick,
					msg.ReplyToMessage.ID,
				),
				ReplyParameters: &models.ReplyParameters{
					MessageID: msg.ID,
				},
				ParseMode: models.ParseModeMarkdown,
			})
			return
		}
		h.logger.Debug(ctx, "SaveHandle: user exists with this nick",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"for", telegram.UserForLogger(msg.ReplyToMessage.From),
			"chat", telegram.ChatForLogger(msg.Chat),
		)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("О\\-ох\\, %s\\, кажется\\, я уже знаю его\\!", telegram.GenerateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}
	err = h.UserUseCase.SaveNickname(ctx, msg.ReplyToMessage.From.ID, schoolNick)
	if err != nil {
		h.logger.Error(ctx, "SaveHandle: dont save user",
			"text", msg.Text,
			"user", telegram.UserForLogger(msg.From),
			"for", telegram.UserForLogger(msg.ReplyToMessage.From),
			"chat", telegram.ChatForLogger(msg.Chat),
			"err", err,
		)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "О-о-о, о нет! Кажется что-то пошло не так...!",
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
		})
		return
	}
	h.userHandler.RemoveUserFromTimers(ctx, b, msg.Chat.ID, msg.ReplyToMessage.From.ID)
	h.logger.Info(ctx, "SaveHandle: save user",
		"text", msg.Text,
		"user", telegram.UserForLogger(msg.From),
		"for", telegram.UserForLogger(msg.ReplyToMessage.From),
		"chat", telegram.ChatForLogger(msg.Chat),
	)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text: fmt.Sprintf("Э\\, %s\\, всё получилось\\! Я записал %s\\, теперь это наш человек\\!",
			telegram.GenerateMention(msg.From),
			telegram.GenerateMention(msg.ReplyToMessage.From),
		),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}
