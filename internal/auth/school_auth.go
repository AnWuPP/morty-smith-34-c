package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type SchoolAuth struct {
	TokenURL  string
	Username  string
	Password  string
	AuthToken string
}

type authResponse struct {
	AccessToken string `json:"access_token"`
}

// NewSchoolAuth создает новый экземпляр SchoolAuth
func NewSchoolAuth(tokenURL, username, password string) *SchoolAuth {
	return &SchoolAuth{
		TokenURL: tokenURL,
		Username: username,
		Password: password,
	}
}

// Authenticate получает токен доступа
func (s *SchoolAuth) Authenticate() error {
	// Формируем тело запроса в формате application/x-www-form-urlencoded
	payload := fmt.Sprintf(
		"client_id=s21-open-api&username=%s&password=%s&grant_type=password",
		s.Username,
		s.Password,
	)

	req, err := http.NewRequest("POST", s.TokenURL, bytes.NewBufferString(payload))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}
	defer resp.Body.Close()

	// Логируем статус ответа для отладки
	log.Printf("Response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		// Логируем тело ответа для диагностики
		var responseBody []byte
		resp.Body.Read(responseBody)
		log.Printf("Response body: %s", string(responseBody))
		return errors.New("authentication failed: invalid credentials or server error")
	}

	// Декодируем токен из ответа
	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Сохраняем токен
	s.AuthToken = authResp.AccessToken
	return nil
}
