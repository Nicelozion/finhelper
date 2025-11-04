package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Server struct {
	cfg Config
	bc  *BankClient
}

func NewServer(cfg Config) *Server {
	return &Server{
		cfg: cfg,
		bc:  NewBankClient(cfg),
	}
}

// health проверка здоровья сервиса
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// createConsent создает согласие на доступ к данным клиента
// POST /api/consents?bank=vbank&user=user-123
func (s *Server) createConsent(w http.ResponseWriter, r *http.Request) {
	bank := r.URL.Query().Get("bank")
	if bank == "" {
		writeErr(w, r, 400, "Missing 'bank' query parameter")
		return
	}

	user := r.URL.Query().Get("user")
	if user == "" {
		user = "demo-user-1"
	}

	// Проверяем, что банк существует
	if _, err := s.bc.bankByCode(bank); err != nil {
		writeErr(w, r, 400, "Invalid bank code: "+bank)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Создаем согласие
	consentID, err := s.bc.EnsureConsent(ctx, bank, user)
	if err != nil {
		writeErr(w, r, 500, "Failed to create consent: "+err.Error())
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"ok":         true,
		"bank":       bank,
		"user":       user,
		"consent_id": consentID,
		"message":    "Consent created successfully",
	})
}

// getConsentStatus получает статус согласия
// GET /api/consents/{id}?bank=vbank
func (s *Server) getConsentStatus(w http.ResponseWriter, r *http.Request) {
	bank := r.URL.Query().Get("bank")
	if bank == "" {
		writeErr(w, r, 400, "Missing 'bank' query parameter")
		return
	}

	consentID := r.PathValue("id")
	if consentID == "" {
		writeErr(w, r, 400, "Missing consent ID in path")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	consent, err := s.bc.GetConsentStatus(ctx, bank, consentID)
	if err != nil {
		writeErr(w, r, 500, "Failed to get consent status: "+err.Error())
		return
	}

	writeJSON(w, 200, consent)
}

// revokeConsent отзывает согласие
// DELETE /api/consents/{id}?bank=vbank
func (s *Server) revokeConsent(w http.ResponseWriter, r *http.Request) {
	bank := r.URL.Query().Get("bank")
	if bank == "" {
		writeErr(w, r, 400, "Missing 'bank' query parameter")
		return
	}

	consentID := r.PathValue("id")
	if consentID == "" {
		writeErr(w, r, 400, "Missing consent ID in path")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := s.bc.RevokeConsent(ctx, bank, consentID); err != nil {
		writeErr(w, r, 500, "Failed to revoke consent: "+err.Error())
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"ok":      true,
		"message": "Consent revoked successfully",
	})
}

// connectBank подключает банк (legacy endpoint для обратной совместимости)
// POST /api/banks/{bank}/connect?user=user-123
func (s *Server) connectBank(w http.ResponseWriter, r *http.Request) {
	bank := r.PathValue("bank")
	switch bank {
	case "vbank", "abank", "sbank":
	default:
		writeErr(w, r, 400, "Invalid bank. Allowed: vbank, abank, sbank")
		return
	}

	user := r.URL.Query().Get("user")
	if user == "" {
		user = "demo-user-1"
	}

	if _, err := s.bc.bankByCode(bank); err != nil {
		writeErr(w, r, 400, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	consentID, err := s.bc.EnsureConsent(ctx, bank, user)
	if err != nil {
		writeErr(w, r, 500, "Failed to connect: "+err.Error())
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"ok":         true,
		"bank":       bank,
		"consent_id": consentID,
	})
}

// accounts получает список всех счетов пользователя
// GET /api/accounts?user=user-123&bank=vbank
func (s *Server) accounts(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		user = "demo-user-1"
	}

	bankFilter := r.URL.Query().Get("bank")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var accs []Account
	var err error

	if bankFilter != "" && bankFilter != "all" {
		// Получаем счета из конкретного банка
		accs, err = s.bc.FetchAccountsFromBank(ctx, bankFilter, user)
	} else {
		// Получаем счета из всех банков
		accs, err = s.bc.FetchAccountsAllBanks(ctx, user)
	}

	if err != nil {
		writeErr(w, r, 500, "Failed to fetch accounts: "+err.Error())
		return
	}

	if accs == nil {
		accs = []Account{}
	}

	writeJSON(w, 200, accs)
}

// accountBalances получает балансы конкретного счета
// GET /api/accounts/{id}/balances?bank=vbank&user=user-123
func (s *Server) accountBalances(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		writeErr(w, r, 400, "Missing account ID in path")
		return
	}

	bank := r.URL.Query().Get("bank")
	if bank == "" {
		writeErr(w, r, 400, "Missing 'bank' query parameter")
		return
	}

	user := r.URL.Query().Get("user")
	if user == "" {
		user = "demo-user-1"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	balances, err := s.bc.FetchBalances(ctx, bank, user, accountID)
	if err != nil {
		writeErr(w, r, 500, "Failed to fetch balances: "+err.Error())
		return
	}

	if balances == nil {
		balances = []BalanceDetail{}
	}

	writeJSON(w, 200, balances)
}

// transactions получает транзакции
// GET /api/transactions?user=user-123&bank=vbank&from=2025-01-01T00:00:00Z&to=2025-12-31T23:59:59Z
func (s *Server) transactions(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		user = "demo-user-1"
	}

	bank := r.URL.Query().Get("bank")

	var fromPtr, toPtr *time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		t, e := time.Parse(time.RFC3339, v)
		if e != nil {
			writeErr(w, r, 400, "Invalid 'from' (RFC3339)")
			return
		}
		fromPtr = &t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, e := time.Parse(time.RFC3339, v)
		if e != nil {
			writeErr(w, r, 400, "Invalid 'to' (RFC3339)")
			return
		}
		toPtr = &t
	}

	if bank != "" && bank != "all" {
		// Проверяем валидность банка
		valid := false
		for _, b := range s.cfg.Banks {
			if b.Code == bank {
				valid = true
				break
			}
		}
		if !valid {
			writeErr(w, r, 400, "Invalid bank")
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	txs, err := s.bc.FetchTransactions(ctx, user, bank, fromPtr, toPtr)
	if err != nil {
		writeErr(w, r, 500, "Failed to fetch transactions: "+err.Error())
		return
	}

	// Форматируем ответ
	resp := make([]map[string]interface{}, len(txs))
	for i, t := range txs {
		resp[i] = map[string]interface{}{
			"id":          t.ID,
			"date":        t.Date.Format(time.RFC3339),
			"amount":      t.Amount,
			"currency":    t.Currency,
			"merchant":    t.Merchant,
			"category":    t.Category,
			"description": t.Description,
			"bank":        t.Bank,
		}
	}

	writeJSON(w, 200, resp)
}

// accountTransactions получает транзакции конкретного счета
// GET /api/accounts/{id}/transactions?bank=vbank&user=user-123&from=2025-01-01&to=2025-12-31
func (s *Server) accountTransactions(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("id")
	if accountID == "" {
		writeErr(w, r, 400, "Missing account ID in path")
		return
	}

	bank := r.URL.Query().Get("bank")
	if bank == "" {
		writeErr(w, r, 400, "Missing 'bank' query parameter")
		return
	}

	user := r.URL.Query().Get("user")
	if user == "" {
		user = "demo-user-1"
	}

	// Парсим даты
	var fromTime, toTime time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeErr(w, r, 400, "Invalid 'from' date format (use YYYY-MM-DD)")
			return
		}
		fromTime = t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			writeErr(w, r, 400, "Invalid 'to' date format (use YYYY-MM-DD)")
			return
		}
		toTime = t
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Получаем согласие
	consentID, err := s.bc.EnsureConsent(ctx, bank, user)
	if err != nil {
		writeErr(w, r, 500, "Failed to ensure consent: "+err.Error())
		return
	}

	// Получаем клиент
	client, err := s.bc.getClient(bank)
	if err != nil {
		writeErr(w, r, 500, "Failed to get bank client: "+err.Error())
		return
	}

	// Получаем транзакции
	transactions, err := client.GetTransactions(ctx, consentID, accountID, fromTime, toTime)
	if err != nil {
		writeErr(w, r, 500, "Failed to fetch transactions: "+err.Error())
		return
	}

	// Конвертируем в legacy формат для ответа
	resp := make([]map[string]interface{}, len(transactions))
	for i, tx := range transactions {
		legacyTx := tx.ToLegacyTransaction(bank)
		resp[i] = map[string]interface{}{
			"id":          legacyTx.ID,
			"date":        legacyTx.Date.Format(time.RFC3339),
			"amount":      legacyTx.Amount,
			"currency":    legacyTx.Currency,
			"merchant":    legacyTx.Merchant,
			"category":    legacyTx.Category,
			"description": legacyTx.Description,
			"bank":        legacyTx.Bank,
		}
	}

	writeJSON(w, 200, resp)
}

// writeJSON пишет JSON ответ
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeErr пишет ошибку в формате JSON
func writeErr(w http.ResponseWriter, r *http.Request, status int, msg string) {
	writeJSON(w, status, ErrorResponse{
		Message: msg,
		Error:   http.StatusText(status),
	})
}