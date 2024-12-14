package repository

import (
	"context"
	"fmt"
	"morty-smith-34-c/internal/app/entity"

	"gorm.io/gorm"
)

type ChatRepository interface {
	Create(ctx context.Context, chat *entity.Chat) error
	UpdateThreadID(ctx context.Context, chatID int64, threadID int) error
	UpdateRulesLink(ctx context.Context, chatID int64, rulesLink string) error
	UpdateFaqLink(ctx context.Context, chatID int64, faqLink string) error
	GetByChatID(ctx context.Context, chatID int64) (*entity.Chat, error)
	GetAllChats(ctx context.Context) ([]*entity.Chat, error)
}

type PostgresChatRepository struct {
	DB *gorm.DB
}

func NewPostgresChatRepository(db *gorm.DB) *PostgresChatRepository {
	return &PostgresChatRepository{DB: db}
}

func (r *PostgresChatRepository) Create(ctx context.Context, chat *entity.Chat) error {
	return r.DB.WithContext(ctx).Create(chat).Error
}

func (r *PostgresChatRepository) UpdateThreadID(ctx context.Context, chatID int64, threadID int) error {
	return r.DB.WithContext(ctx).
		Model(&entity.Chat{}).
		Where("chat_id = ?", chatID).
		Update("thread_id", threadID).Error
}

func (r *PostgresChatRepository) UpdateRulesLink(ctx context.Context, chatID int64, rulesLink string) error {
	return r.DB.WithContext(ctx).
		Model(&entity.Chat{}).
		Where("chat_id = ?", chatID).
		Update("rules_link", rulesLink).Error
}

func (r *PostgresChatRepository) UpdateFaqLink(ctx context.Context, chatID int64, faqLink string) error {
	return r.DB.WithContext(ctx).
		Model(&entity.Chat{}).
		Where("chat_id = ?", chatID).
		Update("faq_link", faqLink).Error
}

func (r *PostgresChatRepository) GetByChatID(ctx context.Context, chatID int64) (*entity.Chat, error) {
	var chat entity.Chat
	if err := r.DB.WithContext(ctx).Where("chat_id = ?", chatID).First(&chat).Error; err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *PostgresChatRepository) GetAllChats(ctx context.Context) ([]*entity.Chat, error) {
	var chats []*entity.Chat
	err := r.DB.WithContext(ctx).Find(&chats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chats: %w", err)
	}
	return chats, nil
}
