package telegram

import (
	"context"
	"fmt"
	"log"
	"morty-smith-34-c/internal/app/entity"
	"morty-smith-34-c/internal/app/usecase"
	"morty-smith-34-c/internal/storage/cache"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CommandHandler struct {
	ChatUseCase *usecase.ChatUseCase // Логика работы с чатами
	UserUseCase *usecase.UserUseCase // Логика проверки пользователей
	chatCache   *cache.ChatCache
}

func NewCommandHandler(chatUseCase *usecase.ChatUseCase, userUseCase *usecase.UserUseCase, chatCache *cache.ChatCache) *CommandHandler {
	return &CommandHandler{
		ChatUseCase: chatUseCase,
		UserUseCase: userUseCase,
		chatCache:   chatCache,
	}
}

func (h *CommandHandler) HandleCommand(ctx context.Context, b *bot.Bot, msg *models.Message) {
	args := strings.Fields(msg.Text)
	if len(args) == 0 {
		return
	}

	// Проверяем роль пользователя
	if args[0] == "/morty_come_here" || args[0] == "/morty_id_topic_here" || args[0] == "/morty_rules" || args[0] == "/morty_faq" {
		err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"superadmin"})
		if err != nil {
			return
		}
	}
	if args[0] == "/mute" || args[0] == "/unmute" {
		err := h.UserUseCase.CheckRole(ctx, msg.From.ID, []string{"moderator", "admin", "superadmin"})
		if err != nil {
			return
		}
	}

	switch args[0] {
	case "/mute":
		h.handleMute(ctx, b, msg, args)
	case "/unmute":
		h.handleUnmute(ctx, b, msg)
	case "/morty_faq":
		if len(args) < 2 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "Я-я-я ничтожество, но-но, прошу дай мне просто записать подсказки...",
			})
			return
		}
		if msg.ReplyToMessage != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "Ох нет, не отвечай на сообщение п-п-пожалуйста, Рик ругается...",
			})
			return
		}
		faq := strings.Join(args[1:], " ")
		err := h.ChatUseCase.UpdateFaqLink(ctx, msg.Chat.ID, faq)
		if err != nil {
			log.Printf("Failed to set faq link: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "О-о-о, о нет! Что-то пошло не так, в-в-вот чёрт, Рик опять будет ругаться...",
			})
			return
		}
		h.chatCache.SetFaq(msg.Chat.ID, faq)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Ура! Я з-з-записал подсказки для чата...",
		})
	case "/morty_rules":
		if len(args) < 2 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "Я-я-я ничтожество, но-но, прошу дай мне просто записать правила...",
			})
			return
		}
		if msg.ReplyToMessage != nil {
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
			log.Printf("Failed to set rules link: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "О-о-о, о нет! Что-то пошло не так, в-в-вот чёрт, Рик опять будет ругаться...",
			})
			return
		}
		h.chatCache.SetRules(msg.Chat.ID, rules)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Ура! Я з-з-записал правила чата...",
		})
	case "/morty_come_here":
		if len(args) < 2 {
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
			log.Printf("Failed to activate chat: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "Эй, эм... кажется, произошла ошибка! Может, чат уже активирован? О-о-о, я весь в поту!",
			})
			return
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text:            "Эй, поздравляю! Чат успешно активирован для кампуса: " + campusName + ". Ура-а-а!",
		})

	case "/morty_id_topic_here":
		if msg.ReplyToMessage != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text:            "Ох нет, не отвечай на сообщение п-п-пожалуйста, Рик ругается...",
			})
			return
		}
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
}

// handleMute handles the /mute command
func (h *CommandHandler) handleMute(ctx context.Context, b *bot.Bot, msg *models.Message, args []string) {
	if msg.ReplyToMessage == nil {
		return
	}

	duration, err := parseDuration(strings.Join(args, " "))
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Ой\\-ой\\, %s\\, кажется\\, ты что\\-то напутал с форматом времени\\!", generateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	if duration < 5*time.Minute {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Минимальное время мута — 5 минут\\, как будто у нас есть время на меньшее, %s\\!", generateMention(msg.From)),
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
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("Ох\\, замутить %s не удалось\\, давай попробуем снова\\, %s\\!", generateMention(msg.ReplyToMessage.From), generateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("Бум\\! %s теперь в муте на %s\\, %s\\!", generateMention(msg.ReplyToMessage.From), strings.Join(args[1:], " "), generateMention(msg.From)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}

// handleUnmute handles the /unmute command
func (h *CommandHandler) handleUnmute(ctx context.Context, b *bot.Bot, msg *models.Message) {
	if msg.ReplyToMessage == nil {
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
	})
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   fmt.Sprintf("О нет\\! Не могу размутить %s\\, %s\\, помоги мне\\!", generateMention(msg.ReplyToMessage.From), generateMention(msg.From)),
			ReplyParameters: &models.ReplyParameters{
				MessageID: msg.ID,
			},
			ParseMode: models.ParseModeMarkdown,
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   fmt.Sprintf("Фух\\! %s размучен\\, %s\\, ты герой\\!", generateMention(msg.ReplyToMessage.From), generateMention(msg.From)),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		ParseMode: models.ParseModeMarkdown,
	})
}

// parseDuration parses human-readable time format into time.Duration
func parseDuration(input string) (time.Duration, error) {
	re := regexp.MustCompile(`(?i)(\d+\s*день|\d+\s*дня|\d+\s*дней|\d+\s*час|\d+\s*минут|\d+\s*секунд)`)
	matches := re.FindAllString(strings.ToLower(input), -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("неверный формат времени")
	}

	var duration time.Duration
	for _, match := range matches {
		var value int
		var err error

		if strings.Contains(match, "день") || strings.Contains(match, "дня") || strings.Contains(match, "дней") {
			value, err = extractNumber(match, "день")
			duration += time.Duration(value) * 24 * time.Hour
		} else if strings.Contains(match, "час") {
			value, err = extractNumber(match, "час")
			duration += time.Duration(value) * time.Hour
		} else if strings.Contains(match, "минут") {
			value, err = extractNumber(match, "минут")
			duration += time.Duration(value) * time.Minute
		} else if strings.Contains(match, "секунд") {
			value, err = extractNumber(match, "секунд")
			duration += time.Duration(value) * time.Second
		}

		if err != nil {
			return 0, err
		}
	}

	return duration, nil
}

// extractNumber extracts the number from the time string
func extractNumber(input, unit string) (int, error) {
	value := strings.TrimSpace(strings.Replace(input, unit, "", 1))
	return strconv.Atoi(value)
}

func generateMention(user *models.User) string {
	if user.FirstName != "" && user.LastName != "" {
		return fmt.Sprintf("[%s %s](tg://user?id=%d)", user.FirstName, user.LastName, user.ID)
	} else if user.FirstName != "" {
		return fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, user.ID)
	} else if user.Username != "" {
		return fmt.Sprintf("[@%s](tg://user?id=%d)", user.Username, user.ID)
	} else if user.FirstName != "" || user.LastName != "" {
		return fmt.Sprintf("[%s %s](tg://user?id=%d)", user.FirstName, user.LastName, user.ID)
	}
	return fmt.Sprintf("[User](tg://user?id=%d)", user.ID)
}
