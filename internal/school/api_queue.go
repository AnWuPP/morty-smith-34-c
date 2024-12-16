package school

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"morty-smith-34-c/pkg/logger"
	"net/http"
	"sync"
	"time"
)

// Request - структура для запросов в очередь.
type Request struct {
	Method   string
	Endpoint string
	Headers  map[string]string
	Body     io.Reader
	Response chan interface{}
}

// APIQueue - структура очереди запросов.
type APIQueue struct {
	httpClient   *http.Client
	requestQueue chan Request
	maxRequests  int
	interval     time.Duration
	lastExecuted time.Time
	mu           sync.Mutex
	logger       *logger.Logger
}

// NewAPIQueue создает новый экземпляр APIQueue.
func NewAPIQueue(maxRequests int, interval time.Duration, httpClient *http.Client, logger *logger.Logger) *APIQueue {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &APIQueue{
		httpClient:   httpClient,
		requestQueue: make(chan Request, 500),
		maxRequests:  maxRequests,
		interval:     interval,
		logger:       logger,
	}
}

// Start запускает обработку очереди запросов.
func (q *APIQueue) Start(ctx context.Context) {
	go func() {
		q.logger.Info(ctx, "Starting API queue")
		limiter := time.NewTicker(q.interval / time.Duration(q.maxRequests))
		defer limiter.Stop()

		for {
			select {
			case req := <-q.requestQueue:
				<-limiter.C
				q.mu.Lock()
				q.lastExecuted = time.Now()
				q.mu.Unlock()
				q.logger.Debug(ctx, fmt.Sprintf("Processing request to %s", req.Endpoint))
				go q.executeRequest(ctx, req)
			case <-ctx.Done():
				q.logger.Info(ctx, "Stopping API queue")
				return
			}
		}
	}()
}

// executeRequest выполняет запрос из очереди.
func (q *APIQueue) executeRequest(ctx context.Context, req Request) {
	q.logger.Debug(ctx, fmt.Sprintf("Executing request: %s %s", req.Method, req.Endpoint))

	// Создаем HTTP-запрос
	httpReq, err := http.NewRequest(req.Method, req.Endpoint, req.Body)
	if err != nil {
		q.logger.Error(ctx, fmt.Sprintf("Failed to create request: %v", err))
		req.Response <- fmt.Errorf("failed to create request: %w", err)
		return
	}

	// Добавляем заголовки
	for key, value := range req.Headers {
		httpReq.Header.Add(key, value)
	}

	// Выполняем запрос
	resp, err := q.httpClient.Do(httpReq)
	if err != nil {
		q.logger.Error(ctx, fmt.Sprintf("Failed to execute request: %v", err))
		req.Response <- fmt.Errorf("failed to execute request: %w", err)
		return
	}

	// Читаем тело ответа в память
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close() // Закрываем оригинальное тело, так как оно уже прочитано
	if err != nil {
		q.logger.Error(ctx, fmt.Sprintf("Failed to read response body: %v", err))
		req.Response <- fmt.Errorf("failed to read response body: %w", err)
		return
	}

	// Создаем новое тело из прочитанного контента
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	q.logger.Debug(ctx, fmt.Sprintf("Request to %s completed with status %d", req.Endpoint, resp.StatusCode))

	// Возвращаем ответ через канал
	req.Response <- resp
}

// AddRequest добавляет запрос в очередь.
func (q *APIQueue) AddRequest(ctx context.Context, method, endpoint string, headers map[string]string, body io.Reader) chan interface{} {
	q.logger.Debug(ctx, fmt.Sprintf("Adding request to queue: %s %s", method, endpoint))
	responseChan := make(chan interface{})
	q.requestQueue <- Request{
		Method:   method,
		Endpoint: endpoint,
		Headers:  headers,
		Body:     body,
		Response: responseChan,
	}
	return responseChan
}
