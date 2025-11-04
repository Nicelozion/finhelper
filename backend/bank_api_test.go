package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestCreateConsent тестирует создание согласия
func TestCreateConsent(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockStatus     int
		clientID       string
		permissions    []string
		wantErr        bool
		wantConsentID  string
	}{
		{
			name: "successful consent creation",
			mockResponse: `{
				"status": "AwaitingAuthorisation",
				"consent_id": "consent-abc123",
				"client_id": "user-001",
				"permissions": ["ReadAccountsDetail", "ReadBalances"]
			}`,
			mockStatus:    http.StatusOK,
			clientID:      "user-001",
			permissions:   []string{"ReadAccountsDetail", "ReadBalances"},
			wantErr:       false,
			wantConsentID: "consent-abc123",
		},
		{
			name: "consent creation with all permissions",
			mockResponse: `{
				"status": "Authorised",
				"consent_id": "consent-xyz789",
				"client_id": "user-002",
				"permissions": ["ReadAccountsDetail", "ReadBalances", "ReadTransactionsDetail"]
			}`,
			mockStatus:    http.StatusCreated,
			clientID:      "user-002",
			permissions:   []string{"ReadAccountsDetail", "ReadBalances", "ReadTransactionsDetail"},
			wantErr:       false,
			wantConsentID: "consent-xyz789",
		},
		{
			name:         "server error",
			mockResponse: `{"error": "internal server error"}`,
			mockStatus:   http.StatusInternalServerError,
			clientID:     "user-003",
			permissions:  []string{"ReadAccountsDetail"},
			wantErr:      true,
		},
		{
			name:         "unauthorized",
			mockResponse: `{"error": "invalid credentials"}`,
			mockStatus:   http.StatusUnauthorized,
			clientID:     "user-004",
			permissions:  []string{"ReadAccountsDetail"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем mock server
			tokenCallCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Первый запрос - получение токена
				if strings.HasSuffix(r.URL.Path, "/auth/bank-token") {
					tokenCallCount++
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"access_token": "test-token-123",
						"token_type":   "bearer",
						"expires_in":   3600,
					})
					return
				}

				// Второй запрос - создание согласия
				if strings.HasSuffix(r.URL.Path, "/account-consents/request") {
					// Проверяем заголовки
					if auth := r.Header.Get("Authorization"); auth != "Bearer test-token-123" {
						t.Errorf("Expected Authorization header 'Bearer test-token-123', got '%s'", auth)
					}
					if reqBank := r.Header.Get("X-Requesting-Bank"); reqBank != "team053" {
						t.Errorf("Expected X-Requesting-Bank 'team053', got '%s'", reqBank)
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					w.Write([]byte(tt.mockResponse))
					return
				}

				t.Errorf("Unexpected request to %s", r.URL.Path)
			}))
			defer server.Close()

			// Создаем клиент
			client := NewVBankClient(server.URL, "team053", "secret123", "team053")

			// Выполняем тест
			ctx := context.Background()
			result, err := client.CreateConsent(ctx, tt.clientID, tt.permissions, "test reason")

			// Проверяем результат
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateConsent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CreateConsent() returned nil result")
					return
				}
				if result.ConsentID != tt.wantConsentID {
					t.Errorf("CreateConsent() ConsentID = %v, want %v", result.ConsentID, tt.wantConsentID)
				}
				if tokenCallCount != 1 {
					t.Errorf("Expected 1 token request, got %d", tokenCallCount)
				}
			}
		})
	}
}

// TestGetAccounts тестирует получение списка счетов
func TestGetAccounts(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockStatus     int
		consentID      string
		clientID       string
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful accounts fetch - array format",
			mockResponse: `[
				{
					"account_id": "acc-001",
					"currency": "RUB",
					"account_type": "Personal",
					"nickname": "Основной счет",
					"account": {
						"identification": "40817810200010318134",
						"name": "Иван Иванов"
					}
				},
				{
					"account_id": "acc-002",
					"currency": "USD",
					"account_type": "Savings",
					"nickname": "Сберегательный",
					"account": {
						"identification": "40817840300020428245",
						"name": "Иван Иванов"
					}
				}
			]`,
			mockStatus: http.StatusOK,
			consentID:  "consent-123",
			clientID:   "user-001",
			wantErr:    false,
			wantCount:  2,
		},
		{
			name: "successful accounts fetch - wrapper format",
			mockResponse: `{
				"accounts": [
					{
						"account_id": "acc-003",
						"currency": "EUR",
						"account_type": "Business"
					}
				]
			}`,
			mockStatus: http.StatusOK,
			consentID:  "consent-456",
			clientID:   "user-002",
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:         "empty accounts list",
			mockResponse: `[]`,
			mockStatus:   http.StatusOK,
			consentID:    "consent-789",
			clientID:     "user-003",
			wantErr:      false,
			wantCount:    0,
		},
		{
			name:         "forbidden - no consent",
			mockResponse: `{"error": "consent not found"}`,
			mockStatus:   http.StatusForbidden,
			consentID:    "invalid-consent",
			clientID:     "user-004",
			wantErr:      true,
		},
		{
			name:         "unauthorized",
			mockResponse: `{"error": "invalid token"}`,
			mockStatus:   http.StatusUnauthorized,
			consentID:    "consent-999",
			clientID:     "user-005",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Токен endpoint
				if strings.HasSuffix(r.URL.Path, "/auth/bank-token") {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]interface{}{
						"access_token": "test-token-456",
						"expires_in":   3600,
					})
					return
				}

				// Accounts endpoint
				if strings.HasSuffix(r.URL.Path, "/accounts") {
					// Проверяем заголовки
					if consentID := r.Header.Get("X-Consent-Id"); consentID != tt.consentID {
						t.Errorf("Expected X-Consent-Id '%s', got '%s'", tt.consentID, consentID)
					}

					// Проверяем query параметр
					if clientID := r.URL.Query().Get("client_id"); tt.clientID != "" && clientID != tt.clientID {
						t.Errorf("Expected client_id '%s', got '%s'", tt.clientID, clientID)
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.mockStatus)
					w.Write([]byte(tt.mockResponse))
					return
				}

				t.Errorf("Unexpected request to %s", r.URL.Path)
			}))
			defer server.Close()

			// Создаем клиент
			client := NewVBankClient(server.URL, "team053", "secret456", "team053")

			// Выполняем тест
			ctx := context.Background()
			accounts, err := client.GetAccounts(ctx, tt.consentID, tt.clientID)

			// Проверяем результат
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(accounts) != tt.wantCount {
					t.Errorf("GetAccounts() returned %d accounts, want %d", len(accounts), tt.wantCount)
				}

				// Проверяем структуру первого счета, если есть
				if tt.wantCount > 0 && len(accounts) > 0 {
					acc := accounts[0]
					if acc.AccountID == "" {
						t.Error("Account ID is empty")
					}
					if acc.Currency == "" {
						t.Error("Currency is empty")
					}
				}
			}
		})
	}
}

// TestTokenCaching тестирует кэширование токенов
func TestTokenCaching(t *testing.T) {
	tokenCallCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/auth/bank-token") {
			tokenCallCount++
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "cached-token",
				"expires_in":   3600,
			})
			return
		}

		if strings.HasSuffix(r.URL.Path, "/accounts") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[]`))
			return
		}
	}))
	defer server.Close()

	client := NewVBankClient(server.URL, "team053", "secret789", "team053")
	ctx := context.Background()

	// Первый запрос - должен получить токен
	_, err := client.GetAccounts(ctx, "consent-1", "user-1")
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}

	// Второй запрос - должен использовать кэшированный токен
	_, err = client.GetAccounts(ctx, "consent-1", "user-1")
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}

	// Третий запрос - тоже должен использовать кэш
	_, err = client.GetAccounts(ctx, "consent-1", "user-1")
	if err != nil {
		t.Fatalf("Third request failed: %v", err)
	}

	// Проверяем, что токен был запрошен только один раз
	if tokenCallCount != 1 {
		t.Errorf("Expected 1 token request, got %d", tokenCallCount)
	}
}

// TestRetryLogic тестирует retry на 5xx ошибки
func TestRetryLogic(t *testing.T) {
	attemptCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/auth/bank-token") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "retry-token",
				"expires_in":   3600,
			})
			return
		}

		if strings.HasSuffix(r.URL.Path, "/accounts") {
			attemptCount++
			// Первые 2 попытки возвращаем 503
			if attemptCount < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error": "service unavailable"}`))
				return
			}
			// Третья попытка успешна
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[]`))
			return
		}
	}))
	defer server.Close()

	client := NewVBankClient(server.URL, "team053", "secret-retry", "team053")
	ctx := context.Background()

	// Должно быть 3 попытки
	_, err := client.GetAccounts(ctx, "consent-retry", "user-retry")
	if err != nil {
		t.Fatalf("Request failed after retry: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

// TestConsentLifecycle тестирует полный жизненный цикл согласия
func TestConsentLifecycle(t *testing.T) {
	createdConsents := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/auth/bank-token") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "lifecycle-token",
				"expires_in":   3600,
			})
			return
		}

		// Создание согласия
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/account-consents/request") {
			consentID := "consent-lifecycle-123"
			createdConsents[consentID] = true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":     "Authorised",
				"consent_id": consentID,
				"client_id":  "user-lifecycle",
			})
			return
		}

		// Получение статуса согласия
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/account-consents/") {
			parts := strings.Split(r.URL.Path, "/")
			consentID := parts[len(parts)-1]
			if !createdConsents[consentID] {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error": "consent not found"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":     "Authorised",
				"consent_id": consentID,
			})
			return
		}

		// Отзыв согласия
		if r.Method == "DELETE" && strings.Contains(r.URL.Path, "/account-consents/") {
			parts := strings.Split(r.URL.Path, "/")
			consentID := parts[len(parts)-1]
			delete(createdConsents, consentID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}))
	defer server.Close()

	client := NewVBankClient(server.URL, "team053", "secret-lifecycle", "team053")
	ctx := context.Background()

	// 1. Создаем согласие
	consent, err := client.CreateConsent(ctx, "user-lifecycle", []string{"ReadAccountsDetail"}, "test")
	if err != nil {
		t.Fatalf("Failed to create consent: %v", err)
	}
	if consent.ConsentID == "" {
		t.Fatal("Consent ID is empty")
	}

	// 2. Проверяем статус
	status, err := client.GetConsentStatus(ctx, consent.ConsentID)
	if err != nil {
		t.Fatalf("Failed to get consent status: %v", err)
	}
	if status.Status != "Authorised" {
		t.Errorf("Expected status 'Authorised', got '%s'", status.Status)
	}

	// 3. Отзываем согласие
	err = client.RevokeConsent(ctx, consent.ConsentID)
	if err != nil {
		t.Fatalf("Failed to revoke consent: %v", err)
	}

	// 4. Проверяем, что согласие удалено
	_, err = client.GetConsentStatus(ctx, consent.ConsentID)
	if err == nil {
		t.Error("Expected error when getting revoked consent, got nil")
	}
}

// TestGetTransactions тестирует получение транзакций
func TestGetTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/auth/bank-token") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "tx-token",
				"expires_in":   3600,
			})
			return
		}

		if strings.Contains(r.URL.Path, "/transactions") {
			// Проверяем query параметры
			fromDate := r.URL.Query().Get("from_date")
			toDate := r.URL.Query().Get("to_date")

			if fromDate != "" && fromDate != "2025-01-01" {
				t.Errorf("Unexpected from_date: %s", fromDate)
			}
			if toDate != "" && toDate != "2025-01-31" {
				t.Errorf("Unexpected to_date: %s", toDate)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[
				{
					"transaction_id": "tx-001",
					"booking_date_time": "2025-01-15T10:30:00Z",
					"amount": {"amount": "1500.00", "currency": "RUB"},
					"credit_debit_indicator": "Debit",
					"status": "Booked",
					"merchant_details": {"merchant_name": "Grocery Store"}
				}
			]`))
			return
		}
	}))
	defer server.Close()

	client := NewVBankClient(server.URL, "team053", "secret-tx", "team053")
	ctx := context.Background()

	fromTime, _ := time.Parse("2006-01-02", "2025-01-01")
	toTime, _ := time.Parse("2006-01-02", "2025-01-31")

	transactions, err := client.GetTransactions(ctx, "consent-tx", "acc-001", fromTime, toTime)
	if err != nil {
		t.Fatalf("GetTransactions failed: %v", err)
	}

	if len(transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(transactions))
	}

	if len(transactions) > 0 {
		tx := transactions[0]
		if tx.TransactionID != "tx-001" {
			t.Errorf("Expected transaction ID 'tx-001', got '%s'", tx.TransactionID)
		}
		if tx.Amount.Amount != "1500.00" {
			t.Errorf("Expected amount '1500.00', got '%s'", tx.Amount.Amount)
		}
	}
}