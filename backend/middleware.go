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

// ContextKey тип для ключей контекста
type ContextKey string

const (
	// CtxRequestID ключ для Request ID в контексте
	CtxRequestID ContextKey = "requestID"
)

// ============================================================================
// MIDDLEWARE КОМПОЗИЦИЯ
// Правильный порядок: Recovery → RequestID → Logging → Timeout → CORS
// ============================================================================

// ApplyMiddleware применяет все middleware в правильном порядке
func ApplyMiddleware(handler http.Handler, corsOrigin string) http.Handler {
	// Применяем в обратном порядке (самый внешний последний)
	handler = withCORS(handler, corsOrigin)
	handler = withTimeout(handler, 90*time.Second)
	handler = withLogging(handler)
	handler = withRequestID(handler)
	handler = withRecovery(handler)
	
	return handler
}

// ============================================================================
// RECOVERY - восстановление после паники (самый внешний слой)
// ============================================================================

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// ✅ Берем request ID из КОНТЕКСТА, а не из заголовка
				requestID := getRequestID(r.Context())
				
				stack := debug.Stack()
				log.Printf("[%s] 🚨 PANIC RECOVERED: %v\n%s", requestID, err, string(stack))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":"Internal Server Error","message":"An unexpected error occurred","request_id":"%s"}`, requestID)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// REQUEST ID - добавление уникального ID запроса
// ============================================================================

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Берем из заголовка или генерируем новый
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Добавляем в заголовок ответа
		w.Header().Set("X-Request-Id", requestID)

		// ✅ Сохраняем в КОНТЕКСТ
		ctx := context.WithValue(r.Context(), CtxRequestID, requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ============================================================================
// LOGGING - логирование запросов
// ============================================================================

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// ✅ Берем request ID из КОНТЕКСТА
		requestID := getRequestID(r.Context())

		// Wrapper для захвата status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Маскируем Bearer токены для безопасности
		authHeader := r.Header.Get("Authorization")
		maskedAuth := maskBearer(authHeader)

		log.Printf("[%s] → %s %s | Auth: %s", 
			requestID, r.Method, r.URL.Path, maskedAuth)

		// Выполняем запрос
		next.ServeHTTP(wrapped, r)

		// Логируем результат
		duration := time.Since(start)
		log.Printf("[%s] ← %s %s | Status: %d | Duration: %v",
			requestID, r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// ============================================================================
// TIMEOUT - ограничение времени выполнения запроса
// ============================================================================

func withTimeout(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ============================================================================
// CORS - настройка Cross-Origin Resource Sharing
// ============================================================================

func withCORS(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ✅ Правильные CORS заголовки
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		
		// ✅ Разрешаем все необходимые заголовки
		w.Header().Set("Access-Control-Allow-Headers", 
			"Content-Type, X-Request-Id, X-Consent-Id, Authorization")
		
		// ✅ Разрешаем клиенту читать заголовки ответа
		w.Header().Set("Access-Control-Expose-Headers", 
			"X-Request-Id, X-Consent-Id")

		// Обрабатываем preflight запросы
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// HELPERS
// ============================================================================

// getRequestID безопасно извлекает request ID из контекста
func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(CtxRequestID).(string); ok {
		return reqID
	}
	return "unknown"
}

// maskBearer маскирует Bearer токены для безопасного логирования
func maskBearer(value string) string {
	if value == "" {
		return "none"
	}

	if strings.HasPrefix(value, "Bearer ") {
		token := value[7:]
		if len(token) > 10 {
			return fmt.Sprintf("Bearer %s...%s", token[:4], token[len(token)-4:])
		}
		return "Bearer ***"
	}

	return "***"
}

// ============================================================================
// RESPONSE WRITER WRAPPER - для захвата status code
// ============================================================================

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