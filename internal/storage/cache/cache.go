package cache

import (
	"context"
	"morty-smith-34-c/internal/app/usecase"
	"sync"
)

type ChatCache struct {
	threadIdCache sync.Map
	rulesCache    sync.Map
	faqCache      sync.Map
}

func NewChatCache() *ChatCache {
	return &ChatCache{}
}

// GetThreadID возвращает threadID для chatID, если он есть в кеше.
func (c *ChatCache) GetThreadID(chatID int64) (int, bool) {
	value, ok := c.threadIdCache.Load(chatID)
	if !ok {
		return -1, false
	}
	return value.(int), true
}

// SetThreadID обновляет или добавляет threadID для chatID в кеш.
func (c *ChatCache) SetThreadID(chatID int64, threadID int) {
	c.threadIdCache.Store(chatID, threadID)
}

func (c *ChatCache) GetRules(chatID int64) (string, bool) {
	value, ok := c.rulesCache.Load(chatID)
	if !ok {
		return "", false
	}
	return value.(string), true
}

func (c *ChatCache) SetRules(chatID int64, rulesLink string) {
	c.rulesCache.Store(chatID, rulesLink)
}

func (c *ChatCache) GetFaq(chatID int64) (string, bool) {
	value, ok := c.faqCache.Load(chatID)
	if !ok {
		return "", false
	}
	return value.(string), true
}

func (c *ChatCache) SetFaq(chatID int64, rulesLink string) {
	c.faqCache.Store(chatID, rulesLink)
}

// LoadFromDatabase загружает данные из базы в кеш.
func (c *ChatCache) LoadFromDatabase(ctx context.Context, chatUseCase *usecase.ChatUseCase) error {
	chats, err := chatUseCase.GetAllChats(ctx)
	if err != nil {
		return err
	}

	for _, chat := range chats {
		c.SetThreadID(chat.ChatID, chat.ThreadID)
		if chat.RulesLink != nil {
			c.SetRules(chat.ChatID, *chat.RulesLink)
		}
		if chat.FaqLink != nil {
			c.SetFaq(chat.ChatID, *chat.FaqLink)
		}
	}
	return nil
}
