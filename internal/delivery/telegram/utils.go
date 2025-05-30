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
		return fmt.Sprintf("[%s %s](tg://user?id=%d)", EscapeMarkdown(user.FirstName), EscapeMarkdown(user.LastName), user.ID)
	} else if user.FirstName != "" {
		return fmt.Sprintf("[%s](tg://user?id=%d)", EscapeMarkdown(user.FirstName), user.ID)
	} else if user.Username != "" {
		return fmt.Sprintf("[@%s](tg://user?id=%d)", EscapeMarkdown(user.FirstName), user.ID)
	}
	return fmt.Sprintf("[User](tg://user?id=%d)", user.ID)
}

func UserForLogger(user *models.User) string {
	return fmt.Sprintf("ID: [%d], Username: [%s], FirstName: [%s], LastName: [%s]", user.ID, user.Username, user.FirstName, user.LastName)
}

func ChatForLogger(chat models.Chat) string {
	return fmt.Sprintf("ID: [%d], Title: [%s]", chat.ID, chat.Title)
}
