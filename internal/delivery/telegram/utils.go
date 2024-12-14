package telegram

import (
	"fmt"
	"strings"

	"github.com/go-telegram/bot/models"
)

func EscapeMarkdown(text string) string {
	specialChars := "_*[]()~`>#+-=|{}.!"
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, string(char), "\\"+string(char))
	}
	return text
}

func GenerateMention(user *models.User) string {
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
