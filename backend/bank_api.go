package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// BankAPIClient клиент для работы с Banking API одного банка
type BankAPIClient struct {
	httpClient     *HTTPClient
	clientID       string
	clientSecret   string
	requestingBank string

	// Кэш токенов с защитой от race condition
	mu          sync.RWMutex
	accessToken string
	tokenExpiry time.Time
}

// NewBankAPIClient создает новый клиент для Banking API
func NewBankAPIClient(baseURL, clientID, clientSecret, requestingBank string) *BankAPIClient {
	return &BankAPIClient{
		httpClient:     NewHTTPClient(baseURL),
		clientID:       clientID,
		clientSecret:   clientSecret,
		requestingBank: requestingBank,
	}
}

// AUTHENTICATION

// EnsureToken получает или возвращает кэшированный токен
func (c *BankAPIClient) EnsureToken(ctx context.Context) (string, error) {
	// Проверяем кэш с защитой от гонок
	c.mu.RLock()
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

	// Запрашиваем токен
	token, err := c.requestToken(ctx)
	if err != nil {
		return "", err
	}

	return token, nil
}

// requestToken запрашивает новый bank token
func (c *BankAPIClient) requestToken(ctx context.Context) (string, error) {
	// ИСПРАВЛЕНИЕ: API ожидает параметры в QUERY STRING!
	queryParams := url.Values{}
	queryParams.Set("client_id", c.clientID)
	queryParams.Set("client_secret", c.clientSecret)

	// Debug логирование
	log.Printf("[DEBUG] Requesting token with client_id=%s (secret length: %d)", 
		c.clientID, len(c.clientSecret))

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodPost,
		Path:        "/auth/bank-token",
		QueryParams: queryParams, // Параметры в query!
		Body:        nil,         // Body пустой!
	})
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errBody := ReadErrorResponse(resp)
		log.Printf("[ERROR] Token request failed: status=%d, body=%s", resp.StatusCode, errBody)
		return "", fmt.Errorf("token request failed: %s", errBody)
	}

	var tokenResp TokenResponse
	if err := ParseJSONResponse(resp, &tokenResp); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response")
	}

	// Сохраняем в кэш
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return c.accessToken, nil
}

// CONSENT MANAGEMENT

// CreateConsent создает согласие на доступ к данным клиента
func (c *BankAPIClient) CreateConsent(ctx context.Context, clientID string, permissions []string, reason string) (*ConsentResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	requestBody := ConsentRequest{
		RequestingBank: c.requestingBank,
		ClientID:       clientID,
		Permissions:    permissions,
		Reason:         reason,
		AutoApproved:   true,
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:  http.MethodPost,
		Path:    "/account-consents/request",
		Body:    requestBody,
		Headers: headers,
	})
	if err != nil {
		return nil, fmt.Errorf("create consent: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create consent failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var consent ConsentResponse
	if err := ParseJSONResponse(resp, &consent); err != nil {
		return nil, fmt.Errorf("parse consent response: %w", err)
	}

	return &consent, nil
}

// GetConsentStatus получает статус согласия
func (c *BankAPIClient) GetConsentStatus(ctx context.Context, consentID string) (*ConsentResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:  http.MethodGet,
		Path:    "/account-consents/" + consentID,
		Headers: headers,
	})
	if err != nil {
		return nil, fmt.Errorf("get consent: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get consent failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var consent ConsentResponse
	if err := ParseJSONResponse(resp, &consent); err != nil {
		return nil, fmt.Errorf("parse consent response: %w", err)
	}

	return &consent, nil
}

// RevokeConsent отзывает согласие
func (c *BankAPIClient) RevokeConsent(ctx context.Context, consentID string) error {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:  http.MethodDelete,
		Path:    "/account-consents/" + consentID,
		Headers: headers,
	})
	if err != nil {
		return fmt.Errorf("revoke consent: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("revoke consent failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	resp.Body.Close()
	return nil
}

// ACCOUNTS

// GetAccounts получает список счетов клиента
func (c *BankAPIClient) GetAccounts(ctx context.Context, consentID, clientID string) ([]AccountDetail, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
		"X-Consent-Id":      consentID,
	}

	//ИСПОЛЬЗУЕМ url.Values для безопасного построения query
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodGet,
		Path:        "/accounts",
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get accounts failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	// Парсим ответ с поддержкой разных форматов
	return c.parseAccountsResponse(resp)
}

// parseAccountsResponse парсит разные форматы ответа со счетами
func (c *BankAPIClient) parseAccountsResponse(resp *http.Response) ([]AccountDetail, error) {
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Вариант 1 массив напрямую
	var directArray []AccountDetail
	if err := json.Unmarshal(bodyBytes, &directArray); err == nil && len(directArray) > 0 {
		return directArray, nil
	}

	// Вариант 2 обертка с полем "accounts" или "account"
	var wrapper AccountsWrapper
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil {
		log.Printf("[DEBUG PARSE] wrapper.Accounts=%d, wrapper.Account=%d, wrapper.Data.Accounts=%d, wrapper.Data.Account=%d",
			len(wrapper.Accounts), len(wrapper.Account), len(wrapper.Data.Accounts), len(wrapper.Data.Account))

		// Проверяем множественное число "accounts"
		if len(wrapper.Accounts) > 0 {
			log.Printf("[DEBUG PARSE] Returning wrapper.Accounts")
			return wrapper.Accounts, nil
		}
		// Проверяем единственное число "account"
		if len(wrapper.Account) > 0 {
			log.Printf("[DEBUG PARSE] Returning wrapper.Account")
			return wrapper.Account, nil
		}
		// Проверяем data.accounts
		if len(wrapper.Data.Accounts) > 0 {
			log.Printf("[DEBUG PARSE] Returning wrapper.Data.Accounts")
			return wrapper.Data.Accounts, nil
		}
		// Проверяем data.account
		if len(wrapper.Data.Account) > 0 {
			log.Printf("[DEBUG PARSE] Returning wrapper.Data.Account")
			return wrapper.Data.Account, nil
		}
	} else {
		log.Printf("[DEBUG PARSE] Unmarshal error: %v", err)
	}

	// Если ничего не распарсилось - возвращаем пустой массив
	log.Printf("[DEBUG] Failed to parse accounts response, returning empty array. Body: %s", string(bodyBytes))
	return []AccountDetail{}, nil
}

// GetAccountDetail получает детали конкретного счета
func (c *BankAPIClient) GetAccountDetail(ctx context.Context, consentID, accountID, clientID string) (*AccountDetail, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
		"X-Consent-Id":      consentID,
	}

	// Добавляем client_id в query если указан
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodGet,
		Path:        "/accounts/" + accountID,
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get account failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var account AccountDetail
	if err := ParseJSONResponse(resp, &account); err != nil {
		return nil, fmt.Errorf("parse account response: %w", err)
	}

	return &account, nil
}

// BALANCES

// GetBalances получает балансы счета
func (c *BankAPIClient) GetBalances(ctx context.Context, consentID, accountID, clientID string) ([]BalanceDetail, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
		"X-Consent-Id":      consentID,
	}

	// Добавляем client_id в query если указан
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodGet,
		Path:        "/accounts/" + accountID + "/balances",
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("get balances: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get balances failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	return c.parseBalancesResponse(resp)
}

// parseBalancesResponse парсит разные форматы ответа с балансами
func (c *BankAPIClient) parseBalancesResponse(resp *http.Response) ([]BalanceDetail, error) {
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Вариант 1 массив напрямую
	var directArray []BalanceDetail
	if err := json.Unmarshal(bodyBytes, &directArray); err == nil && len(directArray) > 0 {
		return directArray, nil
	}

	// Вариант 2 обертка с полем "balances" или "balance"
	var wrapper BalancesWrapper
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil {
		if len(wrapper.Balances) > 0 {
			return wrapper.Balances, nil
		}
		if len(wrapper.Balance) > 0 {
			return wrapper.Balance, nil
		}
		if len(wrapper.Data.Balances) > 0 {
			return wrapper.Data.Balances, nil
		}
		if len(wrapper.Data.Balance) > 0 {
			return wrapper.Data.Balance, nil
		}
	}

	return []BalanceDetail{}, nil
}

// TRANSACTIONS

// GetTransactions получает транзакции счета за период
func (c *BankAPIClient) GetTransactions(ctx context.Context, consentID, accountID, clientID string, from, to time.Time) ([]TransactionDetail, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
		"X-Consent-Id":      consentID,
	}

	// ИСПОЛЬЗУЕМ url.Values для безопасного построения query
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}
	if !from.IsZero() {
		queryParams.Set("from_date", from.Format("2006-01-02"))
	}
	if !to.IsZero() {
		queryParams.Set("to_date", to.Format("2006-01-02"))
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodGet,
		Path:        "/accounts/" + accountID + "/transactions",
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get transactions failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	return c.parseTransactionsResponse(resp)
}

// parseTransactionsResponse парсит разные форматы ответа с транзакциями
func (c *BankAPIClient) parseTransactionsResponse(resp *http.Response) ([]TransactionDetail, error) {
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Вариант 1 массив напрямую
	var directArray []TransactionDetail
	if err := json.Unmarshal(bodyBytes, &directArray); err == nil && len(directArray) > 0 {
		return directArray, nil
	}

	// Вариант 2 обертка с полем "transactions" 
	var wrapper TransactionsWrapper
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil {
		if len(wrapper.Transactions) > 0 {
			return wrapper.Transactions, nil
		}
		if len(wrapper.Data.Transactions) > 0 {
			return wrapper.Data.Transactions, nil
		}
	}

	return []TransactionDetail{}, nil
}

// PAYMENT CONSENT MANAGEMENT

// CreatePaymentConsent создает согласие на выполнение платежа
func (c *BankAPIClient) CreatePaymentConsent(ctx context.Context, req PaymentConsentRequest) (*PaymentConsentResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:  http.MethodPost,
		Path:    "/payment-consents/request",
		Body:    req,
		Headers: headers,
	})
	if err != nil {
		return nil, fmt.Errorf("create payment consent: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create payment consent failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var consent PaymentConsentResponse
	if err := ParseJSONResponse(resp, &consent); err != nil {
		return nil, fmt.Errorf("parse payment consent response: %w", err)
	}

	return &consent, nil
}

// GetPaymentConsentStatus получает статус согласия на платеж
func (c *BankAPIClient) GetPaymentConsentStatus(ctx context.Context, consentID string) (*PaymentConsentResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:  http.MethodGet,
		Path:    "/payment-consents/" + consentID,
		Headers: headers,
	})
	if err != nil {
		return nil, fmt.Errorf("get payment consent: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get payment consent failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var consent PaymentConsentResponse
	if err := ParseJSONResponse(resp, &consent); err != nil {
		return nil, fmt.Errorf("parse payment consent response: %w", err)
	}

	return &consent, nil
}

// PAYMENTS

// CreatePayment создает платеж
func (c *BankAPIClient) CreatePayment(ctx context.Context, paymentConsentID, clientID string, req PaymentRequest) (*PaymentResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":          "Bearer " + token,
		"X-Requesting-Bank":      c.requestingBank,
		"X-Payment-Consent-Id":   paymentConsentID,
	}

	// Добавляем client_id в query если указан
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodPost,
		Path:        "/payments",
		Body:        req,
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create payment failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var payment PaymentResponse
	if err := ParseJSONResponse(resp, &payment); err != nil {
		return nil, fmt.Errorf("parse payment response: %w", err)
	}

	return &payment, nil
}

// GetPaymentStatus получает статус платежа
func (c *BankAPIClient) GetPaymentStatus(ctx context.Context, paymentID, clientID string) (*PaymentResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
	}

	// Добавляем client_id в query если указан
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodGet,
		Path:        "/payments/" + paymentID,
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get payment failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var payment PaymentResponse
	if err := ParseJSONResponse(resp, &payment); err != nil {
		return nil, fmt.Errorf("parse payment response: %w", err)
	}

	return &payment, nil
}

// PRODUCT AGREEMENT CONSENT MANAGEMENT

// CreateProductAgreementConsent создает согласие для работы с продуктами/договорами
func (c *BankAPIClient) CreateProductAgreementConsent(ctx context.Context, req ProductAgreementConsentRequest) (*ProductAgreementConsentResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:  http.MethodPost,
		Path:    "/product-agreement-consents/request",
		Body:    req,
		Headers: headers,
	})
	if err != nil {
		return nil, fmt.Errorf("create PA consent: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create PA consent failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var consent ProductAgreementConsentResponse
	if err := ParseJSONResponse(resp, &consent); err != nil {
		return nil, fmt.Errorf("parse PA consent response: %w", err)
	}

	return &consent, nil
}

// GetProductAgreementConsentStatus получает статус PA согласия
func (c *BankAPIClient) GetProductAgreementConsentStatus(ctx context.Context, consentID string) (*ProductAgreementConsentResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":     "Bearer " + token,
		"X-Requesting-Bank": c.requestingBank,
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:  http.MethodGet,
		Path:    "/product-agreement-consents/" + consentID,
		Headers: headers,
	})
	if err != nil {
		return nil, fmt.Errorf("get PA consent: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get PA consent failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var consent ProductAgreementConsentResponse
	if err := ParseJSONResponse(resp, &consent); err != nil {
		return nil, fmt.Errorf("parse PA consent response: %w", err)
	}

	return &consent, nil
}

// PRODUCTS

// GetProducts получает список доступных продуктов банка
func (c *BankAPIClient) GetProducts(ctx context.Context, clientID string, productType string) ([]Product, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	// Добавляем query параметры
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}
	if productType != "" {
		queryParams.Set("product_type", productType)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodGet,
		Path:        "/products",
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get products failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	return c.parseProductsResponse(resp)
}

// parseProductsResponse парсит разные форматы ответа с продуктами
func (c *BankAPIClient) parseProductsResponse(resp *http.Response) ([]Product, error) {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Вариант 1: массив напрямую
	var directArray []Product
	if err := json.Unmarshal(bodyBytes, &directArray); err == nil && len(directArray) > 0 {
		return directArray, nil
	}

	// Вариант 2: обертка с полем "products"
	var wrapper ProductsWrapper
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil {
		if len(wrapper.Products) > 0 {
			return wrapper.Products, nil
		}
		if len(wrapper.Data.Products) > 0 {
			return wrapper.Data.Products, nil
		}
	}

	return []Product{}, nil
}

// AGREEMENTS (DEPOSIT/LOAN/CARD)

// OpenAgreement открывает договор (вклад/кредит/карта)
func (c *BankAPIClient) OpenAgreement(ctx context.Context, paConsentID, clientID string, req AgreementRequest) (*AgreementResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":                   "Bearer " + token,
		"X-Product-Agreement-Consent-Id":  paConsentID,
	}

	// Добавляем client_id в query если указан
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodPost,
		Path:        "/agreements",
		Body:        req,
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("open agreement: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("open agreement failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var agreement AgreementResponse
	if err := ParseJSONResponse(resp, &agreement); err != nil {
		return nil, fmt.Errorf("parse agreement response: %w", err)
	}

	return &agreement, nil
}

// GetAgreementDetails получает детали договора
func (c *BankAPIClient) GetAgreementDetails(ctx context.Context, paConsentID, agreementID, clientID string) (*AgreementResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":                   "Bearer " + token,
		"X-Product-Agreement-Consent-Id":  paConsentID,
	}

	// Добавляем client_id в query если указан
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodGet,
		Path:        "/agreements/" + agreementID,
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("get agreement: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get agreement failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	var agreement AgreementResponse
	if err := ParseJSONResponse(resp, &agreement); err != nil {
		return nil, fmt.Errorf("parse agreement response: %w", err)
	}

	return &agreement, nil
}

// CloseAgreement закрывает договор
func (c *BankAPIClient) CloseAgreement(ctx context.Context, paConsentID, agreementID, clientID string) (*AgreementResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":                   "Bearer " + token,
		"X-Product-Agreement-Consent-Id":  paConsentID,
	}

	// Добавляем client_id в query если указан
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodDelete,
		Path:        "/agreements/" + agreementID,
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("close agreement: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("close agreement failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	// Для DELETE может вернуться 204 No Content
	if resp.StatusCode == http.StatusNoContent {
		resp.Body.Close()
		return &AgreementResponse{
			AgreementID: agreementID,
			Status:      "closed",
		}, nil
	}

	var agreement AgreementResponse
	if err := ParseJSONResponse(resp, &agreement); err != nil {
		return nil, fmt.Errorf("parse agreement response: %w", err)
	}

	return &agreement, nil
}

// GetAgreements получает список договоров клиента
func (c *BankAPIClient) GetAgreements(ctx context.Context, paConsentID, clientID string) ([]AgreementResponse, error) {
	token, err := c.EnsureToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("ensure token: %w", err)
	}

	headers := map[string]string{
		"Authorization":                   "Bearer " + token,
		"X-Product-Agreement-Consent-Id":  paConsentID,
	}

	// Добавляем client_id в query если указан
	queryParams := url.Values{}
	if clientID != "" {
		queryParams.Set("client_id", clientID)
	}

	resp, err := c.httpClient.DoRequest(ctx, RequestOptions{
		Method:      http.MethodGet,
		Path:        "/agreements",
		Headers:     headers,
		QueryParams: queryParams,
	})
	if err != nil {
		return nil, fmt.Errorf("get agreements: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get agreements failed (%d): %s", resp.StatusCode, ReadErrorResponse(resp))
	}

	return c.parseAgreementsResponse(resp)
}

// parseAgreementsResponse парсит разные форматы ответа с договорами
func (c *BankAPIClient) parseAgreementsResponse(resp *http.Response) ([]AgreementResponse, error) {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Вариант 1 массив напрямую
	var directArray []AgreementResponse
	if err := json.Unmarshal(bodyBytes, &directArray); err == nil && len(directArray) > 0 {
		return directArray, nil
	}

	// Вариант 2 обертка с полем "agreements"
	var wrapper AgreementsWrapper
	if err := json.Unmarshal(bodyBytes, &wrapper); err == nil {
		if len(wrapper.Agreements) > 0 {
			return wrapper.Agreements, nil
		}
		if len(wrapper.Data.Agreements) > 0 {
			return wrapper.Data.Agreements, nil
		}
	}

	return []AgreementResponse{}, nil
}