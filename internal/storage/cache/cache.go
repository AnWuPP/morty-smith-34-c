package cache

import (
	"context"
	"morty-smith-34-c/internal/app/usecase"
	"sync"
)

type ChatCache struct {
	cache sync.Map
}

func NewChatCache() *ChatCache {
	return &ChatCache{}
}

// GetThreadID возвращает threadID для chatID, если он есть в кеше.
func (c *ChatCache) GetThreadID(chatID int64) (int, bool) {
	value, ok := c.cache.Load(chatID)
	if !ok {
		return -1, false
	}
	return value.(int), true
}

// SetThreadID обновляет или добавляет threadID для chatID в кеш.
func (c *ChatCache) SetThreadID(chatID int64, threadID int) {
	c.cache.Store(chatID, threadID)
}

// LoadFromDatabase загружает данные из базы в кеш.
func (c *ChatCache) LoadFromDatabase(ctx context.Context, chatUseCase *usecase.ChatUseCase) error {
	chats, err := chatUseCase.GetAllChats(ctx)
	if err != nil {
		return err
	}

	for _, chat := range chats {
		c.SetThreadID(chat.ChatID, chat.ThreadID)
	}
	return nil
}
