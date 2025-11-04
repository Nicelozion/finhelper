package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ctxKey string

const ctxRequestID ctxKey = "requestID"

var allowedOrigin = env("CORS_ORIGIN", "http://localhost:5173")

// cors добавляет CORS заголовки
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// withRequestID добавляет request ID в контекст и заголовки
func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxRequestID, id)))
	})
}

// withTimeout добавляет таймаут на запрос
func withTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// withLogging логирует HTTP запросы
// Не логирует Bearer токены и чувствительные данные
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrapper для захвата status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Получаем request ID из контекста
		requestID := r.Header.Get("X-Request-Id")

		// Маскируем чувствительные заголовки
		authHeader := r.Header.Get("Authorization")
		maskedAuth := maskSensitive(authHeader)

		log.Printf("[%s] %s %s | Auth: %s | UA: %s",
			requestID,
			r.Method,
			r.URL.Path,
			maskedAuth,
			r.UserAgent(),
		)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s | Status: %d | Duration: %v",
			requestID,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// withRecovery восстанавливается после паники
func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := r.Header.Get("X-Request-Id")
				stack := debug.Stack()

				log.Printf("[%s] PANIC RECOVERED: %v\n%s", requestID, err, string(stack))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":"Internal Server Error","message":"An unexpected error occurred","request_id":"%s"}`, requestID)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// maskSensitive маскирует чувствительные данные (Bearer токены)
func maskSensitive(value string) string {
	if value == "" {
		return "none"
	}

	if strings.HasPrefix(value, "Bearer ") {
		token := value[7:]
		if len(token) > 10 {
			return "Bearer " + token[:4] + "..." + token[len(token)-4:]
		}
		return "Bearer ***"
	}

	return "***"
}

// responseWriter оборачивает http.ResponseWriter для захвата status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}