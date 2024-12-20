package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"morty-smith-34-c/internal/app/usecase"
	"morty-smith-34-c/internal/school"
	"morty-smith-34-c/pkg/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UserHandler struct {
	ChatUseCase *usecase.ChatUseCase
	UserUseCase *usecase.UserUseCase
	jwtService  school.JWTService
	timers      map[int64]*time.Timer
	messageIDs  map[int64]int
	mu          sync.Mutex
	logger      *logger.Logger
}

func NewUserHandler(logger *logger.Logger, chatUseCase *usecase.ChatUseCase, userUseCase *usecase.UserUseCase, jwtService school.JWTService) *UserHandler {
	return &UserHandler{
		ChatUseCase: chatUseCase,
		UserUseCase: userUseCase,
		jwtService:  jwtService,
		timers:      make(map[int64]*time.Timer),
		messageIDs:  make(map[int64]int),
		logger:      logger,
	}
}

func (h *UserHandler) HandleNewMembers(ctx context.Context, b *bot.Bot, msg *models.Message, threadID int) {
	readableChatID := strconv.FormatInt(msg.Chat.ID, 10)[4:]
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	})
	for _, user := range msg.NewChatMembers {
		exists, err := h.UserUseCase.Exists(ctx, user.ID)
		if err != nil {
			h.logger.Error(ctx, "HandleNewMembers: Failed to check user existence",
				"user", UserForLogger(msg.From),
				"chat", ChatForLogger(msg.Chat),
				"err", err,
			)
			continue
		}

		if exists {
			// Приветствуем существующего пользователя
			sendMessage, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text: fmt.Sprintf(
					"Эй\\, %s\\! С возвращением!",
					GenerateMention(&user),
				),
				ParseMode: models.ParseModeMarkdown,
			})
			if err != nil {
				h.logger.Debug(ctx, "HandleNewMembers: Failed to send welcome back message",
					"user", UserForLogger(msg.From),
					"chat", ChatForLogger(msg.Chat),
					"err", err,
				)
				continue
			}
			time.AfterFunc(time.Minute*2, func() {
				b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    msg.Chat.ID,
					MessageID: sendMessage.ID,
				})
			})
			continue
		}
		sendMessage, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:          msg.Chat.ID,
			MessageThreadID: msg.MessageThreadID,
			Text: fmt.Sprintf(
				"Добро пожаловать\\, %s\\, у тебя есть **5 минут**\\, чтобы написать свой **школьный ник** в топик [ID](https://t.me/c/%s/%d)\\.",
				GenerateMention(&user), readableChatID, threadID,
			),
			ParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			h.logger.Debug(ctx, "HandleNewMembers: Failed to send welcome message",
				"user", UserForLogger(msg.From),
				"chat", ChatForLogger(msg.Chat),
				"err", err,
			)
			continue
		}

		h.mu.Lock()
		h.messageIDs[user.ID] = sendMessage.ID
		h.mu.Unlock()

		timer := time.AfterFunc(5*time.Minute, func() {
			h.mu.Lock()
			defer h.mu.Unlock()

			if _, exists := h.timers[user.ID]; exists {
				b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    msg.Chat.ID,
					MessageID: sendMessage.ID,
				})

				_, err := b.BanChatMember(ctx, &bot.BanChatMemberParams{
					ChatID:    msg.Chat.ID,
					UserID:    user.ID,
					UntilDate: 0,
				})
				h.logger.Info(ctx, "HandleNewMembers: Ban user after timeout",
					"user", UserForLogger(&user),
					"chat", ChatForLogger(msg.Chat),
				)
				if err != nil {
					h.logger.Debug(ctx, "HandleNewMembers: Failed to ban user",
						"user", UserForLogger(msg.From),
						"chat", ChatForLogger(msg.Chat),
						"err", err,
					)
				}

				delete(h.timers, user.ID)
				delete(h.messageIDs, user.ID)
			}
		})

		h.mu.Lock()
		h.timers[user.ID] = timer
		h.mu.Unlock()
	}
}

func (h *UserHandler) HandleNickname(ctx context.Context, b *bot.Bot, msg *models.Message) {
	if _, exists := h.timers[msg.From.ID]; !exists {
		h.logger.Debug(ctx, "HandleNickname: User does not have timer",
			"text", msg.Text,
			"user", UserForLogger(msg.From),
			"chat", ChatForLogger(msg.Chat),
		)
		return
	}
	msg.Text = strings.ToLower(msg.Text)
	_, err := h.jwtService.CheckUser(ctx, msg.Text)
	if err != nil {
		if err.Error() == "user not found" {
			h.logger.Info(ctx, "HandleNickname: User not found in School API",
				"text", msg.Text,
				"user", UserForLogger(msg.From),
				"chat", ChatForLogger(msg.Chat),
			)
			sendMessage, _ := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text: fmt.Sprintf(
					"Эй\\, %s\\! Не могу найти твой ник в Школе 21\\. Попробуй еще раз\\, без опечаток\\!",
					GenerateMention(msg.From),
				),
				ReplyParameters: &models.ReplyParameters{
					MessageID: msg.ID,
				},
				MessageThreadID: msg.MessageThreadID,
				ParseMode:       models.ParseModeMarkdown,
			})
			b.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
				ChatID:    msg.Chat.ID,
				MessageID: msg.ID,
				Reaction: []models.ReactionType{
					{
						Type:              models.ReactionTypeTypeEmoji,
						ReactionTypeEmoji: &models.ReactionTypeEmoji{Emoji: "🤔"},
					},
				},
			})
			time.AfterFunc(time.Minute*2, func() {
				b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    msg.Chat.ID,
					MessageID: msg.ID,
				})
				b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    msg.Chat.ID,
					MessageID: sendMessage.ID,
				})
			})
			return
		}
		if err.Error() == "not core program" || err.Error() == "profile blocked" {
			h.RemoveUserFromTimers(ctx, b, msg.Chat.ID, msg.From.ID)

			_, err := b.BanChatMember(ctx, &bot.BanChatMemberParams{
				ChatID:    msg.Chat.ID,
				UserID:    msg.From.ID,
				UntilDate: 0,
			})
			h.logger.Info(ctx, "HandleNewMembers: Ban user after check",
				"user", UserForLogger(msg.From),
				"chat", ChatForLogger(msg.Chat),
				"err", err,
			)
			if err != nil {
				h.logger.Debug(ctx, "HandleNickname: Failed to ban user",
					"user", UserForLogger(msg.From),
					"chat", ChatForLogger(msg.Chat),
				)
			}

			b.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
				ChatID:    msg.Chat.ID,
				MessageID: msg.ID,
				Reaction: []models.ReactionType{
					{
						Type:              models.ReactionTypeTypeEmoji,
						ReactionTypeEmoji: &models.ReactionTypeEmoji{Emoji: "🤬"},
					},
				},
			})
			return
		}

		// Ошибка при обращении к API
		h.logger.Error(ctx, "HandleNickname: Bad request to School API",
			"text", msg.Text,
			"user", UserForLogger(msg.From),
			"chat", ChatForLogger(msg.Chat),
			"err", err,
		)
		return
	}

	// Удаляем таймер и приветственное сообщение
	h.RemoveUserFromTimers(ctx, b, msg.Chat.ID, msg.From.ID)

	// Сохраняем ник в базе данных
	err = h.UserUseCase.SaveNickname(ctx, msg.From.ID, msg.Text)
	if err != nil {
		h.logger.Error(ctx, "HandleNickname: Error save to database",
			"text", msg.Text,
			"user", UserForLogger(msg.From),
			"chat", ChatForLogger(msg.Chat),
			"err", err,
		)
		return
	}

	b.SetMessageReaction(ctx, &bot.SetMessageReactionParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
		Reaction: []models.ReactionType{
			{
				Type:              models.ReactionTypeTypeEmoji,
				ReactionTypeEmoji: &models.ReactionTypeEmoji{Emoji: "🔥"},
			},
		},
	})
	h.logger.Info(ctx, "HandleNickname: Validated user in School API",
		"text", msg.Text,
		"user", UserForLogger(msg.From),
		"chat", ChatForLogger(msg.Chat),
	)
	// Подтверждение для пользователя
	sendMessage, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text: fmt.Sprintf(
			"Круто\\, %s\\! Я проверил\\, и всё в порядке\\. Ты — наш человек\\! Соблюдай правила нашего сообщества\\!",
			GenerateMention(msg.From),
		),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: bot.True(),
		},
		ParseMode: models.ParseModeMarkdown,
	})
	time.AfterFunc(time.Minute*2, func() {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    msg.Chat.ID,
			MessageID: sendMessage.ID,
		})
	})
}

func (h *UserHandler) RemoveUserFromTimers(ctx context.Context, b *bot.Bot, chatID int64, userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if timer, exists := h.timers[userID]; exists {
		timer.Stop()
		delete(h.timers, userID)
	}

	if messageID, exists := h.messageIDs[userID]; exists {
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: messageID,
		})
		delete(h.messageIDs, userID)
	}
}
