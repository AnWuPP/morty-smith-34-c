package usecase

import (
	"context"
	"morty-smith-34-c/internal/app/entity"
	"morty-smith-34-c/internal/app/repository"
)

type ChatUseCase struct {
	ChatRepo repository.ChatRepository
}

func NewChatUseCase(chatRepo repository.ChatRepository) *ChatUseCase {
	return &ChatUseCase{
		ChatRepo: chatRepo,
	}
}

// Create добавляет новый чат
func (u *ChatUseCase) Create(ctx context.Context, chat *entity.Chat) error {
	return u.ChatRepo.Create(ctx, chat)
}

// UpdateThreadID обновляет ThreadID для чата
func (u *ChatUseCase) UpdateThreadID(ctx context.Context, chatID int64, threadID int) error {
	return u.ChatRepo.UpdateThreadID(ctx, chatID, threadID)
}

// UpdateRulesLink обновляет ссылку на правила для чата
func (u *ChatUseCase) UpdateRulesLink(ctx context.Context, chatID int64, rulesLink string) error {
	return u.ChatRepo.UpdateRulesLink(ctx, chatID, rulesLink)
}

// UpdateRulesLink обновляет ссылку на помощь для чата
func (u *ChatUseCase) UpdateFaqLink(ctx context.Context, chatID int64, faqLink string) error {
	return u.ChatRepo.UpdateFaqLink(ctx, chatID, faqLink)
}

// GetByChatID возвращает информацию о чате
func (u *ChatUseCase) GetByChatID(ctx context.Context, chatID int64) (*entity.Chat, error) {
	return u.ChatRepo.GetByChatID(ctx, chatID)
}

func (c *ChatUseCase) GetAllChats(ctx context.Context) ([]*entity.Chat, error) {
	return c.ChatRepo.GetAllChats(ctx)
}
