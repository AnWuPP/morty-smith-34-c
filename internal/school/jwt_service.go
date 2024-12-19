package school

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"morty-smith-34-c/pkg/logger"
	"net/http"
	"strings"
	"sync"
	"time"
)

// JWTService - интерфейс для работы с JWT.
type JWTService interface {
	Authenticate(ctx context.Context) error             // Первичная аутентификация
	RefreshTokens(ctx context.Context) error            // Обновление токенов
	GetAccessToken(ctx context.Context) (string, error) // Получение текущего Access-токена
	CheckUser(ctx context.Context, login string) (*UserResponse, error)
}

// jwtService - основная реализация JWTService.
type jwtService struct {
	username     string
	password     string
	authEndpoint string
	apiBaseURL   string
	tokenCache   *TokenCache
	logger       *logger.Logger
	apiQueue     ApiQueue
}

// TokenCache - структура для хранения токенов в памяти.
type TokenCache struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	sync.RWMutex
}

// JwtResponse - структура для ответа сервера при аутентификации.
type JwtResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// UserResponse - структура ответа от API.
type UserResponse struct {
	Login          string `json:"login"`
	ClassName      string `json:"className"`
	ParallelName   string `json:"parallelName"`
	ExpValue       int    `json:"expValue"`
	Level          int    `json:"level"`
	ExpToNextLevel int    `json:"expToNextLevel"`
	Campus         struct {
		ID        string `json:"id"`
		ShortName string `json:"shortName"`
	} `json:"campus"`
	Status string `json:"status"`
}

// NewJWTService создает новый экземпляр jwtService.
func NewJWTService(username, password, authEndpoint, apiBaseURL string, logger *logger.Logger, apiQueue ApiQueue) JWTService {
	return &jwtService{
		username:     username,
		password:     password,
		authEndpoint: authEndpoint,
		apiBaseURL:   apiBaseURL,
		tokenCache:   &TokenCache{},
		logger:       logger,
		apiQueue:     apiQueue,
	}
}

// Authenticate - выполняет первичную аутентификацию и сохраняет токены.
func (j *jwtService) Authenticate(ctx context.Context) error {
	data := fmt.Sprintf("client_id=s21-open-api&username=%s&password=%s&grant_type=password", j.username, j.password)
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	responseChan := j.apiQueue.AddRequest(ctx, "POST", j.authEndpoint, headers, bytes.NewBuffer([]byte(data)))
	result := <-responseChan

	resp, ok := result.(*http.Response)
	if !ok || resp == nil {
		err := fmt.Errorf("failed to authenticate, invalid response: %v", result)
		j.logger.Error(ctx, err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("authorization failed: %s", string(body))
		j.logger.Error(ctx, err.Error())
		return err
	}

	var jwtResponse JwtResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwtResponse); err != nil {
		err := fmt.Errorf("failed to parse response: %w", err)
		j.logger.Error(ctx, err.Error())
		return err
	}

	j.tokenCache.Lock()
	defer j.tokenCache.Unlock()

	j.tokenCache.AccessToken = jwtResponse.AccessToken
	j.tokenCache.RefreshToken = jwtResponse.RefreshToken
	j.tokenCache.Expiry = time.Now().Add(time.Duration(jwtResponse.ExpiresIn) * time.Second)
	j.logger.Info(ctx, "Successfully authenticated")
	return nil
}

// RefreshTokens - обновляет токены, используя refresh_token.
func (j *jwtService) RefreshTokens(ctx context.Context) error {
	j.tokenCache.RLock()
	refreshToken := j.tokenCache.RefreshToken
	j.tokenCache.RUnlock()

	data := fmt.Sprintf("client_id=s21-open-api&grant_type=refresh_token&refresh_token=%s", refreshToken)
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	responseChan := j.apiQueue.AddRequest(ctx, "POST", j.authEndpoint, headers, bytes.NewBuffer([]byte(data)))
	result := <-responseChan

	resp, ok := result.(*http.Response)
	if !ok || resp == nil {
		err := fmt.Errorf("failed to refresh tokens, invalid response: %v", result)
		j.logger.Error(ctx, err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMessage := string(body)

		// Если RefreshToken недействителен, повторяем авторизацию через логин и пароль
		if resp.StatusCode == http.StatusBadRequest && isInvalidGrant(errMessage) {
			j.logger.Warn(ctx, "Refresh token is invalid, re-authenticating with username and password")
			return j.Authenticate(ctx)
		}

		err := fmt.Errorf("failed to refresh tokens: %s", errMessage)
		j.logger.Error(ctx, err.Error())
		return err
	}

	var jwtResponse JwtResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwtResponse); err != nil {
		err := fmt.Errorf("failed to parse response: %w", err)
		j.logger.Error(ctx, err.Error())
		return err
	}

	j.tokenCache.Lock()
	defer j.tokenCache.Unlock()

	j.tokenCache.AccessToken = jwtResponse.AccessToken
	j.tokenCache.RefreshToken = jwtResponse.RefreshToken
	j.tokenCache.Expiry = time.Now().Add(time.Duration(jwtResponse.ExpiresIn) * time.Second)
	j.logger.Info(ctx, "Tokens successfully refreshed")
	return nil
}

// isInvalidGrant проверяет, содержит ли ошибка описание `invalid_grant`.
func isInvalidGrant(errMessage string) bool {
	return strings.Contains(errMessage, "invalid_grant")
}

// GetAccessToken - возвращает текущий access_token или обновляет его, если он истёк.
func (j *jwtService) GetAccessToken(ctx context.Context) (string, error) {
	j.tokenCache.RLock()
	expiry := j.tokenCache.Expiry
	j.tokenCache.RUnlock()

	if time.Now().After(expiry) {
		j.logger.Info(ctx, "Access token expired, refreshing tokens")
		if err := j.RefreshTokens(ctx); err != nil {
			return "", err
		}
	}

	j.tokenCache.RLock()
	defer j.tokenCache.RUnlock()
	return j.tokenCache.AccessToken, nil
}

// CheckUser - проверяет ник пользователя через API платформы.
func (j *jwtService) CheckUser(ctx context.Context, login string) (*UserResponse, error) {
	// Выполняем запрос к API с обработкой 401 Unauthorized
	return j.checkUserWithRetry(ctx, login, false)
}

func (j *jwtService) checkUserWithRetry(ctx context.Context, login string, retried bool) (*UserResponse, error) {
	// Получаем Access Token
	token, err := j.GetAccessToken(ctx)
	if err != nil {
		j.logger.Error(ctx, fmt.Sprintf("Failed to get access token: %v", err))
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Формируем URL для API запроса
	url := fmt.Sprintf("%s/participants/%s", j.apiBaseURL, login)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	responseChan := j.apiQueue.AddRequest(ctx, "GET", url, headers, nil)
	result := <-responseChan

	resp, ok := result.(*http.Response)
	if !ok || resp == nil {
		err := fmt.Errorf("failed to check user, invalid response: %v", result)
		j.logger.Error(ctx, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	// Обрабатываем ответ
	if resp.StatusCode == http.StatusUnauthorized {
		if retried {
			j.logger.Error(ctx, "Re-authentication failed")
			return nil, errors.New("unauthorized after retry")
		}

		j.logger.Info(ctx, "Access token expired, re-authenticating")
		if err := j.Authenticate(ctx); err != nil {
			j.logger.Error(ctx, fmt.Sprintf("Failed to re-authenticate: %v", err))
			return nil, fmt.Errorf("failed to re-authenticate: %w", err)
		}

		// Повторяем запрос после обновления токена
		return j.checkUserWithRetry(ctx, login, true)
	}

	if resp.StatusCode == http.StatusNotFound {
		err := fmt.Errorf("user not found")
		j.logger.Debug(ctx, err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("failed to check user: %s", string(body))
		j.logger.Error(ctx, err.Error())
		return nil, err
	}

	// Декодируем JSON ответ
	var userResponse UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		err := fmt.Errorf("failed to parse response: %w", err)
		j.logger.Error(ctx, err.Error())
		return nil, err
	}

	// Дополнительные проверки
	if userResponse.ParallelName != "Core program" {
		return nil, fmt.Errorf("not core program")
	}

	if userResponse.Status == "BLOCKED" {
		return nil, fmt.Errorf("profile blocked")
	}

	j.logger.Debug(ctx, fmt.Sprintf("CheckUser: User found: %s", userResponse.Login))
	return &userResponse, nil
}
