package telegram

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"morty-smith-34-c/internal/app/usecase"
	"morty-smith-34-c/internal/pkg/jwtservice"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type UserHandler struct {
	ChatUseCase *usecase.ChatUseCase // Указатель на ChatUseCase
	UserUseCase *usecase.UserUseCase // Указатель на UserUseCase
	jwtService  jwtservice.JWTService
	timers      map[int64]*time.Timer // Таймеры для пользователей
	messageIDs  map[int64]int         // Сообщения для удаления
	mu          sync.Mutex            // Защита при доступе к timers/messageIDs
}

func NewUserHandler(chatUseCase *usecase.ChatUseCase, userUseCase *usecase.UserUseCase, jwtService jwtservice.JWTService) *UserHandler {
	return &UserHandler{
		ChatUseCase: chatUseCase,
		UserUseCase: userUseCase,
		jwtService:  jwtService,
		timers:      make(map[int64]*time.Timer),
		messageIDs:  make(map[int64]int),
	}
}

func (h *UserHandler) HandleNewMembers(ctx context.Context, b *bot.Bot, msg *models.Message) {
	chat, err := h.ChatUseCase.GetByChatID(ctx, msg.Chat.ID)
	if err != nil {
		return
	}
	readableChatID := strconv.FormatInt(msg.Chat.ID, 10)[4:]
	for _, user := range msg.NewChatMembers {
		exists, err := h.UserUseCase.Exists(ctx, user.ID)
		if err != nil {
			log.Printf("Failed to check user existence: %v", err)
			continue
		}

		if exists {
			// Приветствуем существующего пользователя
			sendMessage, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:          msg.Chat.ID,
				MessageThreadID: msg.MessageThreadID,
				Text: fmt.Sprintf(
					"Эй\\, [%s](tg://user?id=%d)\\! Ты снова здесь\\? Как круто\\! Добро пожаловать обратно\\, чувак\\! Надеюсь\\, ты готов к тому\\, что Рик опять начнёт шутить про тебя\\!",
					user.FirstName, user.ID,
				),
				ParseMode: models.ParseModeMarkdown,
			})
			if err != nil {
				log.Printf("Failed to send message: %v", err)
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
				"О\\-о\\-ох\\, эй\\, [%s](tg://user?id=%d)\\, ты новенький\\, да\\? Ладно, послушай\\! Тебе нужно сбросить [сюда](https://t.me/c/%s/%d) свой школьный ник\\, м\\-может быть\\, окей\\? Зачем\\? А\\-а\\-а\\-а я не знаю\\, просто правила такие\\! Ну\\, пожалуйста\\, сделай это\\, пока Рик не начал ворчать\\!",
				user.FirstName, user.ID, readableChatID, chat.ThreadID,
			),
			ParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			log.Printf("Failed to send message: %v", err)
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
				if err != nil {
					log.Printf("Failed to ban user: %v", err)
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
		return
	}

	// Проверяем ник через JWTService
	userInfo, err := h.jwtService.CheckUser(msg.Text)
	if err != nil {
		if err.Error() == "user not found" {
			sendMessage, _ := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: msg.Chat.ID,
				Text: fmt.Sprintf(
					"О нет\\, [%s](tg://user?id=%d)\\! Чувак\\, я тут\\.\\.\\. мм\\.\\.\\. я тут посмотрел\\, и ничего не нашёл\\! Может\\, ты неправильно написал\\? О\\-о\\-о\\, как мне теперь быть\\? Попробуй ещё раз\\, п\\-пожалуйста\\!",
					msg.From.FirstName, msg.From.ID,
				),
				ReplyParameters: &models.ReplyParameters{
					MessageID: msg.ID,
				},
				MessageThreadID: msg.MessageThreadID,
				ParseMode:       models.ParseModeMarkdown,
			})
			time.AfterFunc(time.Minute*2, func() {
				b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    msg.Chat.ID,
					MessageID: sendMessage.ID,
				})
			})
			return
		}

		// Ошибка при обращении к API
		log.Printf("Failed to check user: %v", err)
		return
	}

	// Удаляем таймер и приветственное сообщение
	h.RemoveUserFromTimers(ctx, b, msg.Chat.ID, msg.From.ID)

	// Сохраняем ник в базе данных
	err = h.UserUseCase.SaveNickname(ctx, msg.From.ID, msg.Text)
	if err != nil {
		log.Printf("Failed to save user to database: %v", err)
		return
	}

	// Подтверждение для пользователя
	sendMessage, _ := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text: fmt.Sprintf(
			"О\\-о\\-о\\, круто\\, [%s](tg://user?id=%d)\\! Я проверил\\, и всё в порядке\\. Ты — наш человек\\! Э\\-э, ну\\, ладно\\, наверное\\.\\.\\. мм\\, читай [правила](https://t.me/c/1975595161/309127/309128)\\, чтобы не попасть в беду\\, окей\\? А я пойду спрячусь где\\-нибудь\\, пока Рик не начал шутить про меня\\.\\.\\.",
			userInfo.Login, msg.From.ID,
		),
		ReplyParameters: &models.ReplyParameters{
			MessageID: msg.ID,
		},
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: bot.True(),
		},
		MessageThreadID: msg.MessageThreadID,
		ParseMode:       models.ParseModeMarkdown,
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
