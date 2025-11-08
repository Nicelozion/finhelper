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
// PAYMENT CONSENT ENDPOINTS
// ============================================================================

// handleCreatePaymentConsent создает согласие на платеж
// POST /api/payment-consents?bank=vbank&user=user-123
func (s *Server) handleCreatePaymentConsent(w http.ResponseWriter, r *http.Request) {
	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	// Парсим тело запроса
	var paymentInfo PaymentInfo
	if err := json.NewDecoder(r.Body).Decode(&paymentInfo); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	consentID, err := s.aggregator.EnsurePaymentConsent(r.Context(), bankCode, userID, paymentInfo)
	if err != nil {
		log.Printf("[%s] Failed to create payment consent: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to create payment consent: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"bank":       bankCode,
		"user":       userID,
		"consent_id": consentID,
		"message":    "Payment consent created successfully",
	})
}

// handleGetPaymentConsentStatus получает статус payment consent
// GET /api/payment-consents/{id}?bank=vbank
func (s *Server) handleGetPaymentConsentStatus(w http.ResponseWriter, r *http.Request) {
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

	consent, err := s.aggregator.GetPaymentConsentStatus(r.Context(), bankCode, consentID)
	if err != nil {
		log.Printf("[%s] Failed to get payment consent status: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to get payment consent status: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, consent)
}

// ============================================================================
// PAYMENT ENDPOINTS
// ============================================================================

// handleCreatePayment создает платеж
// POST /api/payments?bank=vbank&user=user-123
func (s *Server) handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	// Парсим тело запроса
	var paymentReq PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&paymentReq); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	payment, err := s.aggregator.CreatePayment(r.Context(), bankCode, userID, paymentReq)
	if err != nil {
		log.Printf("[%s] Failed to create payment: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to create payment: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, payment)
}

// handleGetPaymentStatus получает статус платежа
// GET /api/payments/{id}?bank=vbank&user=user-123
func (s *Server) handleGetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("id")
	if paymentID == "" {
		writeError(w, r, http.StatusBadRequest, "Missing payment ID in path")
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

	payment, err := s.aggregator.GetPaymentStatus(r.Context(), bankCode, paymentID, userID)
	if err != nil {
		log.Printf("[%s] Failed to get payment status: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to get payment status: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, payment)
}

// ============================================================================
// PRODUCT AGREEMENT CONSENT ENDPOINTS
// ============================================================================

// handleCreatePAConsent создает PA consent
// POST /api/pa-consents?bank=vbank&user=user-123
func (s *Server) handleCreatePAConsent(w http.ResponseWriter, r *http.Request) {
	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	consentID, err := s.aggregator.EnsureProductAgreementConsent(r.Context(), bankCode, userID)
	if err != nil {
		log.Printf("[%s] Failed to create PA consent: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to create PA consent: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"bank":       bankCode,
		"user":       userID,
		"consent_id": consentID,
		"message":    "PA consent created successfully",
	})
}

// handleGetPAConsentStatus получает статус PA consent
// GET /api/pa-consents/{id}?bank=vbank
func (s *Server) handleGetPAConsentStatus(w http.ResponseWriter, r *http.Request) {
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

	consent, err := s.aggregator.GetProductAgreementConsentStatus(r.Context(), bankCode, consentID)
	if err != nil {
		log.Printf("[%s] Failed to get PA consent status: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to get PA consent status: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, consent)
}

// ============================================================================
// PRODUCT ENDPOINTS
// ============================================================================

// handleGetProducts получает список продуктов
// GET /api/products?bank=vbank&user=user-123&type=DEPOSIT
func (s *Server) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	productType := r.URL.Query().Get("type")

	products, err := s.aggregator.GetProducts(r.Context(), bankCode, userID, productType)
	if err != nil {
		log.Printf("[%s] Failed to get products: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to get products: "+err.Error())
		return
	}

	if products == nil {
		products = []Product{}
	}

	writeJSON(w, http.StatusOK, products)
}

// ============================================================================
// AGREEMENT ENDPOINTS
// ============================================================================

// handleOpenAgreement открывает договор
// POST /api/agreements?bank=vbank&user=user-123
func (s *Server) handleOpenAgreement(w http.ResponseWriter, r *http.Request) {
	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	// Парсим тело запроса
	var agreementReq AgreementRequest
	if err := json.NewDecoder(r.Body).Decode(&agreementReq); err != nil {
		writeError(w, r, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	agreement, err := s.aggregator.OpenAgreement(r.Context(), bankCode, userID, agreementReq)
	if err != nil {
		log.Printf("[%s] Failed to open agreement: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to open agreement: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, agreement)
}

// handleGetAgreements получает список договоров
// GET /api/agreements?bank=vbank&user=user-123
func (s *Server) handleGetAgreements(w http.ResponseWriter, r *http.Request) {
	bankCode := r.URL.Query().Get("bank")
	if bankCode == "" {
		writeError(w, r, http.StatusBadRequest, "Missing 'bank' query parameter")
		return
	}

	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "demo-user-1"
	}

	agreements, err := s.aggregator.GetAgreements(r.Context(), bankCode, userID)
	if err != nil {
		log.Printf("[%s] Failed to get agreements: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to get agreements: "+err.Error())
		return
	}

	if agreements == nil {
		agreements = []AgreementResponse{}
	}

	writeJSON(w, http.StatusOK, agreements)
}

// handleGetAgreementDetails получает детали договора
// GET /api/agreements/{id}?bank=vbank&user=user-123
func (s *Server) handleGetAgreementDetails(w http.ResponseWriter, r *http.Request) {
	agreementID := r.PathValue("id")
	if agreementID == "" {
		writeError(w, r, http.StatusBadRequest, "Missing agreement ID in path")
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

	agreement, err := s.aggregator.GetAgreementDetails(r.Context(), bankCode, agreementID, userID)
	if err != nil {
		log.Printf("[%s] Failed to get agreement details: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to get agreement details: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, agreement)
}

// handleCloseAgreement закрывает договор
// DELETE /api/agreements/{id}?bank=vbank&user=user-123
func (s *Server) handleCloseAgreement(w http.ResponseWriter, r *http.Request) {
	agreementID := r.PathValue("id")
	if agreementID == "" {
		writeError(w, r, http.StatusBadRequest, "Missing agreement ID in path")
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

	agreement, err := s.aggregator.CloseAgreement(r.Context(), bankCode, agreementID, userID)
	if err != nil {
		log.Printf("[%s] Failed to close agreement: %v", getRequestID(r.Context()), err)
		writeError(w, r, http.StatusInternalServerError, "Failed to close agreement: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, agreement)
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