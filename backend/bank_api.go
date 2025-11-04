package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// BankAPI определяет интерфейс для работы с банковским API
// Реализует паттерн Repository для абстракции от конкретного банка
type BankAPI interface {
	// CreateConsent создает согласие на доступ к данным клиента
	// Возвращает consent_id и authorize_url для подтверждения клиентом
	CreateConsent(ctx context.Context, clientID string, permissions []string, reason string) (*ConsentResponse, error)
	
	// GetConsentStatus получает статус согласия
	GetConsentStatus(ctx context.Context, consentID string) (*ConsentResponse, error)
	
	// RevokeConsent отзывает согласие
	RevokeConsent(ctx context.Context, consentID string) error
	
	// GetAccounts получает список счетов клиента
	GetAccounts(ctx context.Context, consentID, clientID string) ([]AccountDetail, error)
	
	// GetAccountDetail получает детали конкретного счета
	GetAccountDetail(ctx context.Context, consentID, accountID string) (*AccountDetail, error)
	
	// GetBalances получает балансы счета
	GetBalances(ctx context.Context, consentID, accountID string) ([]BalanceDetail, error)
	
	// GetTransactions получает транзакции счета за период
	GetTransactions(ctx context.Context, consentID, accountID string, from, to time.Time) ([]TransactionDetail, error)
}

// VBankClient реализует BankAPI для Virtual Bank
type VBankClient struct {
	baseURL       string
	clientID      string
	clientSecret  string
	requestingBank string
	httpc         *http.Client
	
	// Кэш токенов с защитой от гонок
	mu          sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

// NewVBankClient создает новый клиент для работы с Virtual Bank API
func NewVBankClient(baseURL, clientID, clientSecret, requestingBank string) *VBankClient {
	return &VBankClient{
		baseURL:        baseURL,
		clientID:       clientID,
		clientSecret:   clientSecret,
		requestingBank: requestingBank,
		httpc: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ensureBankToken получает или обновляет bank token
// Автоматически обновляет токен за 60 секунд до истечения
func (c *VBankClient) ensureBankToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	// Проверяем, есть ли валидный токен (с запасом 60 секунд)
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-60*time.Second)) {
		token := c.accessToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	// Получаем новый токен
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check после получения lock
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-60*time.Second)) {
		return c.accessToken, nil
	}

	// Формируем запрос на получение bank token
	reqBody := map[string]string{
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, 
		c.baseURL+"/auth/bank-token", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Выполняем запрос с retry логикой
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = c.httpc.Do(req)
		if err != nil {
			if attempt == 2 {
				return "", fmt.Errorf("bank-token request failed after 3 attempts: %w", err)
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		
		// Retry на 5xx ошибки
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			resp.Body.Close()
			if attempt == 2 {
				return "", fmt.Errorf("bank-token returned 5xx after 3 attempts: %d", resp.StatusCode)
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		
		break
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bank-token failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
		ClientID    string `json:"client_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response")
	}

	// Сохраняем токен в кэш
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return c.accessToken, nil
}

// CreateConsent создает согласие на доступ к данным клиента
func (c *VBankClient) CreateConsent(ctx context.Context, clientID string, permissions []string, reason string) (*ConsentResponse, error) {
	token, err := c.ensureBankToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	// Формируем тело запроса согласно спецификации
	reqBody := map[string]interface{}{
		"requesting_bank": c.requestingBank,
		"client_id":       clientID,
		"permissions":     permissions,
		"reason":          reason,
		"auto_approved":   true, // Для упрощения в хакатоне
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal consent request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/account-consents/request", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create consent request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requesting-Bank", c.requestingBank)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("consent request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create consent failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var consentResp ConsentResponse
	if err := json.NewDecoder(resp.Body).Decode(&consentResp); err != nil {
		return nil, fmt.Errorf("decode consent response: %w", err)
	}

	return &consentResp, nil
}

// GetConsentStatus получает статус согласия
func (c *VBankClient) GetConsentStatus(ctx context.Context, consentID string) (*ConsentResponse, error) {
	token, err := c.ensureBankToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/account-consents/"+consentID, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requesting-Bank", c.requestingBank)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("get consent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get consent failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var consentResp ConsentResponse
	if err := json.NewDecoder(resp.Body).Decode(&consentResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &consentResp, nil
}

// RevokeConsent отзывает согласие
func (c *VBankClient) RevokeConsent(ctx context.Context, consentID string) error {
	token, err := c.ensureBankToken(ctx)
	if err != nil {
		return fmt.Errorf("ensure token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		c.baseURL+"/account-consents/"+consentID, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requesting-Bank", c.requestingBank)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return fmt.Errorf("revoke consent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke consent failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	return nil
}

// GetAccounts получает список счетов клиента
func (c *VBankClient) GetAccounts(ctx context.Context, consentID, clientID string) ([]AccountDetail, error) {
	token, err := c.ensureBankToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	// Формируем URL с query параметром
	url := c.baseURL + "/accounts"
	if clientID != "" {
		url += "?client_id=" + clientID
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requesting-Bank", c.requestingBank)
	req.Header.Set("X-Consent-Id", consentID)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get accounts failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Читаем сырой ответ для гибкости парсинга
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Пробуем разные форматы ответа
	var accounts []AccountDetail
	
	// Вариант 1: массив напрямую
	if err := json.Unmarshal(bodyBytes, &accounts); err == nil {
		return accounts, nil
	}

	// Вариант 2: объект с полем "accounts"
	var wrapper struct {
		Accounts []AccountDetail `json:"accounts"`
		Data     struct {
			Accounts []AccountDetail `json:"account"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil {
		if len(wrapper.Accounts) > 0 {
			return wrapper.Accounts, nil
		}
		if len(wrapper.Data.Accounts) > 0 {
			return wrapper.Data.Accounts, nil
		}
	}

	return nil, fmt.Errorf("failed to parse accounts response")
}

// GetAccountDetail получает детали конкретного счета
func (c *VBankClient) GetAccountDetail(ctx context.Context, consentID, accountID string) (*AccountDetail, error) {
	token, err := c.ensureBankToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/accounts/"+accountID, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requesting-Bank", c.requestingBank)
	req.Header.Set("X-Consent-Id", consentID)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get account failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var account AccountDetail
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &account, nil
}

// GetBalances получает балансы счета
func (c *VBankClient) GetBalances(ctx context.Context, consentID, accountID string) ([]BalanceDetail, error) {
	token, err := c.ensureBankToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/accounts/"+accountID+"/balances", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requesting-Bank", c.requestingBank)
	req.Header.Set("X-Consent-Id", consentID)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("get balances: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get balances failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var balances []BalanceDetail
	
	// Вариант 1: массив напрямую
	if err := json.Unmarshal(bodyBytes, &balances); err == nil {
		return balances, nil
	}

	// Вариант 2: объект с полем "balances"
	var wrapper struct {
		Balances []BalanceDetail `json:"balances"`
		Data     struct {
			Balance []BalanceDetail `json:"balance"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil {
		if len(wrapper.Balances) > 0 {
			return wrapper.Balances, nil
		}
		if len(wrapper.Data.Balance) > 0 {
			return wrapper.Data.Balance, nil
		}
	}

	return nil, fmt.Errorf("failed to parse balances response")
}

// GetTransactions получает транзакции счета за период
func (c *VBankClient) GetTransactions(ctx context.Context, consentID, accountID string, from, to time.Time) ([]TransactionDetail, error) {
	token, err := c.ensureBankToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	// Формируем URL с query параметрами
	url := c.baseURL + "/accounts/" + accountID + "/transactions"
	if !from.IsZero() || !to.IsZero() {
		url += "?"
		if !from.IsZero() {
			url += "from_date=" + from.Format("2006-01-02")
		}
		if !to.IsZero() {
			if !from.IsZero() {
				url += "&"
			}
			url += "to_date=" + to.Format("2006-01-02")
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Requesting-Bank", c.requestingBank)
	req.Header.Set("X-Consent-Id", consentID)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get transactions failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var transactions []TransactionDetail
	
	// Вариант 1: массив напрямую
	if err := json.Unmarshal(bodyBytes, &transactions); err == nil {
		return transactions, nil
	}

	// Вариант 2: объект с полем "transactions"
	var wrapper struct {
		Transactions []TransactionDetail `json:"transactions"`
		Data         struct {
			Transaction []TransactionDetail `json:"transaction"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil {
		if len(wrapper.Transactions) > 0 {
			return wrapper.Transactions, nil
		}
		if len(wrapper.Data.Transaction) > 0 {
			return wrapper.Data.Transaction, nil
		}
	}

	return nil, fmt.Errorf("failed to parse transactions response")
}

// doWithRetry выполняет HTTP запрос с retry логикой на 5xx ошибки
func (c *VBankClient) doWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < 3; attempt++ {
		// Клонируем запрос для повторных попыток
		reqClone := req.Clone(req.Context())
		
		resp, err = c.httpc.Do(reqClone)
		if err != nil {
			if attempt == 2 {
				return nil, fmt.Errorf("request failed after 3 attempts: %w", err)
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Retry на 5xx ошибки
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			resp.Body.Close()
			if attempt == 2 {
				return nil, fmt.Errorf("server returned 5xx after 3 attempts: %d", resp.StatusCode)
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Успешный ответ или 4xx ошибка (не retry)
		return resp, nil
	}

	return resp, err
}