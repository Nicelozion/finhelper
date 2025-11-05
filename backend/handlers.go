package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Server обрабатывает HTTP запросы
type Server struct {
	aggregator *BankAggregator
	config     Config
}

// NewServer создает новый HTTP сервер
func NewServer(config Config) *Server {
	return &Server{
		aggregator: NewBankAggregator(config),
		config:     config,
	}
}

// ============================================================================
// HEALTH CHECK
// ============================================================================

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"banks":     len(s.config.Banks),
	})
}

// ============================================================================
// CONSENT ENDPOINTS
// ============================================================================

// handleCreateConsent создает согласие на доступ к данным
// POST /api/consents?bank=vbank&user=user-123
func (s *Server) handleCreateConsent(w http.ResponseWriter, r *http.Request) {
	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1" // default для тестирования
	}

	// Проверяем что банк существует
	if _, err := s.aggregator.GetBankByCode(bankCode); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid bank code: "+bankCode)
		return
	}

	// Создаем consent
	consentID, err := s.aggregator.EnsureConsent(r.Context(), bankCode, userID)
	if err != nil {
		log.Printf("[%s] Failed to create consent: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to create consent: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"bank":       bankCode,
		"user":       userID,
		"consent_id": consentID,
		"message":    "Consent created successfully",
	})
}

// handleGetConsentStatus получает статус согласия
// GET /api/consents/{id}?bank=vbank
func (s *Server) handleGetConsentStatus(w http.ResponseWriter, r *http.Request) {
	consentID := r.PathValue("id")
	if consentID == "" {
		writeError(w, r, http.StatusBadRequest, "Missing consent ID in path")
		return
	}

	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	consent, err := s.aggregator.GetConsentStatus(r.Context(), bankCode, consentID)
	if err != nil {
		log.Printf("[%s] Failed to get consent status: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to get consent status: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, consent)
}

// handleRevokeConsent отзывает согласие
// DELETE /api/consents/{id}?bank=vbank
func (s *Server) handleRevokeConsent(w http.ResponseWriter, r *http.Request) {
	consentID := r.PathValue("id")
	if consentID == "" {
		writeError(w, r, http.StatusBadRequest, "Missing consent ID in path")
		return
	}

	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	if err := s.aggregator.RevokeConsent(r.Context(), bankCode, consentID); err != nil {
		log.Printf("[%s] Failed to revoke consent: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to revoke consent: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Consent revoked successfully",
	})
}

// ============================================================================
// LEGACY BANK CONNECTION ENDPOINT (обратная совместимость)
// ============================================================================

// handleConnectBank legacy endpoint для подключения банка
// POST /api/banks/{bank}/connect?user=user-123
func (s *Server) handleConnectBank(w http.ResponseWriter, r *http.Request) {
	bankCode := r.PathValue("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing bank code in path")
		return
	}

	// Валидация банка
	validBanks := map[string]bool{"vbank": true, "abank": true, "sbank": true}
	if !validBanks[bankCode] {
		writeError(w, r, http.StatusBadRequest, "Invalid bank. Allowed: vbank, abank, sbank")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	// Создаем consent
	consentID, err := s.aggregator.EnsureConsent(r.Context(), bankCode, userID)
	if err != nil {
		log.Printf("[%s] Failed to connect bank: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to connect: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"bank":       bankCode,
		"consent_id": consentID,
		"user":       userID,
		"message":    "Bank connected successfully",
	})
}

// ============================================================================
// ACCOUNT ENDPOINTS
// ============================================================================

// handleGetAccounts получает счета пользователя
// GET /api/accounts?user=user-123&bank=vbank
func (s *Server) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	bankFilter := r.URL.Query().Get("bank")

	var accounts []Account
	var err error

	if bankFilter != "" && bankFilter != "all" {
		// Счета из конкретного банка
		accounts, err = s.aggregator.GetAccountsFromBank(r.Context(), bankFilter, userID)
	} else {
		// Счета из всех банков
		accounts, err = s.aggregator.GetAccountsFromAllBanks(r.Context(), userID)
	}

	if err != nil {
		log.Printf("[%s] Failed to fetch accounts: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to fetch accounts: "+err.Error())
		return
	}

	// Возвращаем пустой массив вместо null
	if accounts == nil {
		accounts = []Account{}
	}

	writeJSON(w, http.StatusOK, accounts)
}

// handleGetAccountBalances получает балансы счета
// GET /api/accounts/{id}/balances?bank=vbank&user=user-123
func (s *Server) handleGetAccountBalances(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		writeError(w, r, http.StatusBadRequest, "Missing account ID in path")
		return
	}

	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	balances, err := s.aggregator.GetAccountBalances(r.Context(), bankCode, userID, accountID)
	if err != nil {
		log.Printf("[%s] Failed to fetch balances: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to fetch balances: "+err.Error())
		return
	}

	if balances == nil {
		balances = []BalanceDetail{}
	}

	writeJSON(w, http.StatusOK, balances)
}

// handleGetAccountTransactions получает транзакции конкретного счета
// GET /api/accounts/{id}/transactions?bank=vbank&user=user-123&from=2025-01-01&to=2025-12-31
func (s *Server) handleGetAccountTransactions(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		writeError(w, r, http.StatusBadRequest, "Missing account ID in path")
		return
	}

	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	// Парсим даты
	var fromTime, toTime time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "Invalid 'from' date format (use YYYY-MM-DD)")
			return
		}
		fromTime = t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "Invalid 'to' date format (use YYYY-MM-DD)")
			return
		}
		toTime = t
	}

	transactions, err := s.aggregator.GetAccountTransactions(r.Context(), bankCode, userID, accountID, fromTime, toTime)
	if err != nil {
		log.Printf("[%s] Failed to fetch account transactions: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to fetch transactions: "+err.Error())
		return
	}

	// Форматируем ответ
	response := formatTransactionsResponse(transactions)
	writeJSON(w, http.StatusOK, response)
}

// ============================================================================
// TRANSACTION ENDPOINTS
// ============================================================================

// handleGetTransactions получает транзакции со всех счетов или из конкретного банка
// GET /api/transactions?user=user-123&bank=vbank&from=2025-01-01T00:00:00Z&to=2025-12-31T23:59:59Z
func (s *Server) handleGetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	bankFilter := r.URL.Query().Get("bank")

	// Парсим даты в формате RFC3339
	var fromPtr, toPtr *time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "Invalid 'from' date format (use RFC3339)")
			return
		}
		fromPtr = &t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "Invalid 'to' date format (use RFC3339)")
			return
		}
		toPtr = &t
	}

	// Валидация банка
	if bankFilter != "" && bankFilter != "all" {
		if _, err := s.aggregator.GetBankByCode(bankFilter); err != nil {
			writeError(w, r, http.StatusBadRequest, "Invalid bank code: "+bankFilter)
			return
		}
	}

	transactions, err := s.aggregator.GetTransactions(r.Context(), userID, bankFilter, fromPtr, toPtr)
	if err != nil {
		log.Printf("[%s] Failed to fetch transactions: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to fetch transactions: "+err.Error())
		return
	}

	// Форматируем ответ
	response := formatTransactionsResponse(transactions)
	writeJSON(w, http.StatusOK, response)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// writeJSON отправляет JSON ответ
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// writeError отправляет ошибку в JSON формате
func writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	requestID := getRequestID(r.Context())
	
	response := ErrorResponse{
		Error:     http.StatusText(status),
		Message:   message,
		RequestID: requestID,
	}
	
	writeJSON(w, status, response)
}

// formatTransactionsResponse форматирует транзакции для ответа
func formatTransactionsResponse(transactions []Transaction) []map[string]interface{} {
	response := make([]map[string]interface{}, len(transactions))
	
	for i, tx := range transactions {
		response[i] = map[string]interface{}{
			"id":          tx.ID,
			"date":        tx.Date.Format(time.RFC3339),
			"amount":      tx.Amount,
			"currency":    tx.Currency,
			"merchant":    tx.Merchant,
			"category":    tx.Category,
			"description": tx.Description,
			"bank":        tx.Bank,
		}
	}
	
	return response
}