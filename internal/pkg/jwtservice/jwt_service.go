package jwtservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// JWTService - интерфейс для работы с JWT.
type JWTService interface {
	Authenticate() error             // Первичная аутентификация
	RefreshTokens() error            // Обновление токенов
	GetAccessToken() (string, error) // Получение текущего Access-токена
	CheckUser(login string) (*UserResponse, error)
}

// jwtService - основная реализация JWTService.
type jwtService struct {
	username     string
	password     string
	authEndpoint string
	apiBaseURL   string
	httpClient   *http.Client
	tokenCache   *TokenCache
	logger       Logger
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

// Logger - интерфейс для логирования.
type Logger interface {
	Info(msg string)
	Error(msg string)
}

// NewJWTService создает новый экземпляр jwtService.
func NewJWTService(username, password, authEndpoint, apiBaseURL string, httpClient *http.Client, logger Logger) JWTService {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &jwtService{
		username:     username,
		password:     password,
		authEndpoint: authEndpoint,
		apiBaseURL:   apiBaseURL,
		httpClient:   httpClient,
		tokenCache:   &TokenCache{},
		logger:       logger,
	}
}

// Authenticate - выполняет первичную аутентификацию и сохраняет токены.
func (j *jwtService) Authenticate() error {
	data := fmt.Sprintf("client_id=s21-open-api&username=%s&password=%s&grant_type=password", j.username, j.password)
	req, err := http.NewRequest("POST", j.authEndpoint, bytes.NewBuffer([]byte(data)))
	if err != nil {
		j.logger.Error(fmt.Sprintf("Failed to create request: %v", err))
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		j.logger.Error(fmt.Sprintf("Failed to send request: %v", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		j.logger.Error(fmt.Sprintf("Authorization failed: %s", string(body)))
		return errors.New("authorization failed")
	}

	var jwtResponse JwtResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwtResponse); err != nil {
		j.logger.Error(fmt.Sprintf("Failed to parse response: %v", err))
		return err
	}

	j.tokenCache.Lock()
	defer j.tokenCache.Unlock()

	j.tokenCache.AccessToken = jwtResponse.AccessToken
	j.tokenCache.RefreshToken = jwtResponse.RefreshToken
	j.tokenCache.Expiry = time.Now().Add(time.Duration(jwtResponse.ExpiresIn) * time.Second)
	j.logger.Info("Successfully authenticated")
	return nil
}

// RefreshTokens - обновляет токены, используя refresh_token.
func (j *jwtService) RefreshTokens() error {
	j.tokenCache.RLock()
	refreshToken := j.tokenCache.RefreshToken
	j.tokenCache.RUnlock()

	data := fmt.Sprintf("client_id=s21-open-api&grant_type=refresh_token&refresh_token=%s", refreshToken)
	req, err := http.NewRequest("POST", j.authEndpoint, bytes.NewBuffer([]byte(data)))
	if err != nil {
		j.logger.Error(fmt.Sprintf("Failed to create request: %v", err))
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		j.logger.Error(fmt.Sprintf("Failed to send request: %v", err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		j.logger.Error(fmt.Sprintf("Failed to refresh tokens: %s", string(body)))
		return errors.New("failed to refresh tokens")
	}

	var jwtResponse JwtResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwtResponse); err != nil {
		j.logger.Error(fmt.Sprintf("Failed to parse response: %v", err))
		return err
	}

	j.tokenCache.Lock()
	defer j.tokenCache.Unlock()

	j.tokenCache.AccessToken = jwtResponse.AccessToken
	j.tokenCache.RefreshToken = jwtResponse.RefreshToken
	j.tokenCache.Expiry = time.Now().Add(time.Duration(jwtResponse.ExpiresIn) * time.Second)
	j.logger.Info("Tokens successfully refreshed")
	return nil
}

// GetAccessToken - возвращает текущий access_token или обновляет его, если он истёк.
func (j *jwtService) GetAccessToken() (string, error) {
	j.tokenCache.RLock()
	expiry := j.tokenCache.Expiry
	j.tokenCache.RUnlock()

	if time.Now().After(expiry) {
		j.logger.Info("Access token expired, refreshing tokens")
		if err := j.RefreshTokens(); err != nil {
			return "", err
		}
	}

	j.tokenCache.RLock()
	defer j.tokenCache.RUnlock()
	return j.tokenCache.AccessToken, nil
}

// CheckUser - проверяет ник пользователя через API платформы.
func (j *jwtService) CheckUser(login string) (*UserResponse, error) {
	// Выполняем запрос к API с обработкой 401 Unauthorized
	return j.checkUserWithRetry(login, false)
}

func (j *jwtService) checkUserWithRetry(login string, retried bool) (*UserResponse, error) {
	// Получаем Access Token
	token, err := j.GetAccessToken()
	if err != nil {
		j.logger.Error(fmt.Sprintf("Failed to get access token: %v", err))
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Формируем URL для API запроса
	url := fmt.Sprintf("%s/participants/%s", j.apiBaseURL, login)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		j.logger.Error(fmt.Sprintf("Failed to create request: %v", err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Выполняем запрос
	resp, err := j.httpClient.Do(req)
	if err != nil {
		j.logger.Error(fmt.Sprintf("Failed to send request: %v", err))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Обрабатываем ответ
	if resp.StatusCode == http.StatusUnauthorized {
		if retried {
			j.logger.Error("Re-authentication failed")
			return nil, errors.New("unauthorized after retry")
		}

		j.logger.Info("Access token expired, re-authenticating")
		if err := j.Authenticate(); err != nil {
			j.logger.Error(fmt.Sprintf("Failed to re-authenticate: %v", err))
			return nil, fmt.Errorf("failed to re-authenticate: %w", err)
		}

		// Повторяем запрос после обновления токена
		return j.checkUserWithRetry(login, true)
	}

	if resp.StatusCode == http.StatusNotFound {
		j.logger.Info(fmt.Sprintf("User not found: %s", login))
		return nil, errors.New("user not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		j.logger.Error(fmt.Sprintf("Failed to check user: %s", string(body)))
		return nil, fmt.Errorf("failed to check user: %s", string(body))
	}

	// Декодируем JSON ответ
	var userResponse UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		j.logger.Error(fmt.Sprintf("Failed to parse response: %v", err))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Дополнительные проверки
	if userResponse.ParallelName != "Core program" {
		return nil, fmt.Errorf("not core program")
	}

	if userResponse.Status != "ACTIVE" {
		return nil, fmt.Errorf("profile not active")
	}

	j.logger.Info(fmt.Sprintf("User found: %s", userResponse.Login))
	return &userResponse, nil
}
