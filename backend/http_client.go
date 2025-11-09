package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient обертка над http.Client с retry логикой и проверками
type HTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPClient создает новый HTTP клиент
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// RequestOptions опции для HTTP запроса
type RequestOptions struct {
	Method      string
	Path        string
	Body        interface{}
	Headers     map[string]string
	QueryParams url.Values
}

// DoRequest выполняет HTTP запрос с retry логикой и проверками
func (c *HTTPClient) DoRequest(ctx context.Context, opts RequestOptions) (*http.Response, error) {
	// Формируем полный URL
	fullURL := c.baseURL + opts.Path
	
	// Добавляем query параметры если есть
	if len(opts.QueryParams) > 0 {
		fullURL += "?" + opts.QueryParams.Encode()
	}

	// Подготавливаем тело запроса
	var bodyReader io.Reader
	var bodyBytes []byte
	
	if opts.Body != nil {
		var err error
		bodyBytes, err = json.Marshal(opts.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
		
		// Debug: логируем тело запроса
		log.Printf("[DEBUG HTTP] %s %s | Body: %s", opts.Method, fullURL, string(bodyBytes))
	} else {
		log.Printf("[DEBUG HTTP] %s %s | No body", opts.Method, fullURL)
	}

	// Retry логика (до 3 попыток)
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			// Экспоненциальная задержка между попытками
			time.Sleep(time.Duration(attempt) * time.Second)
			
			// ВАЖНО: пересоздаем bodyReader для повторной попытки
			if bodyBytes != nil {
				bodyReader = bytes.NewReader(bodyBytes)
			}
		}

		// Создаем запрос
		req, err := http.NewRequestWithContext(ctx, opts.Method, fullURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		// Устанавливаем заголовки
		req.Header.Set("Accept", "application/json")
		if opts.Body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		
		for key, value := range opts.Headers {
			req.Header.Set(key, value)
		}

		// Выполняем запрос
		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request failed: %w", err)
			continue // retry
		}

		// КРИТИЧЕСКАЯ ПРОВЕРКА: Content-Type должен быть application/json
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			// Читаем первые 256 байт для диагностики
			preview := make([]byte, 256)
			n, _ := io.ReadFull(resp.Body, preview)
			resp.Body.Close()
			
			return nil, fmt.Errorf("invalid Content-Type: %s (expected application/json). Response preview: %s", 
				contentType, string(preview[:n]))
		}

		// Retry на 5xx ошибки
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
			continue // retry
		}

		// Успех или 4xx ошибка (не retry)
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after 3 attempts: %w", lastErr)
}

// ParseJSONResponse парсит JSON ответ из response body
func ParseJSONResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if err := json.Unmarshal(bodyBytes, target); err != nil {
		return fmt.Errorf("parse JSON (body: %s): %w", string(bodyBytes), err)
	}

	return nil
}

// ReadErrorResponse читает тело ошибки для диагностики
func ReadErrorResponse(resp *http.Response) string {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("failed to read error body: %v", err)
	}
	
	return string(body)
}