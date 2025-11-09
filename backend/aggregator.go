package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// BankAggregator агрегирует данные из нескольких банков
type BankAggregator struct {
	config  Config
	clients map[string]*BankAPIClient

	// Кэш consent ID для каждого банка и пользователя
	mu                     sync.RWMutex
	consentCache           map[string]string // key: "bank|userID" - consentID (account consent)
	paymentConsentCache    map[string]string // key: "bank|userID" - payment consent ID
	paConsentCache         map[string]string // key: "bank|userID" - PA consent ID
}

// NewBankAggregator создает новый агрегатор банков
func NewBankAggregator(config Config) *BankAggregator {
	agg := &BankAggregator{
		config:              config,
		clients:             make(map[string]*BankAPIClient),
		consentCache:        make(map[string]string),
		paymentConsentCache: make(map[string]string),
		paConsentCache:      make(map[string]string),
	}

	// Создаем клиентов для каждого банка
	for _, bank := range config.Banks {
		agg.clients[bank.Code] = NewBankAPIClient(
			bank.BaseURL,
			config.TeamID,
			config.ClientSecret,
			config.TeamID,
		)
		log.Printf("Initialized client for bank: %s (%s)", bank.Code, bank.BaseURL)
	}

	return agg
}

// CONSENT MANAGEMENT

// EnsureConsent создает consent если его нет, или возвращает существующий
func (a *BankAggregator) EnsureConsent(ctx context.Context, bankCode, userID string) (string, error) {
	cacheKey := bankCode + "|" + userID

	// Проверяем кэш
	a.mu.RLock()
	if consentID, exists := a.consentCache[cacheKey]; exists {
		a.mu.RUnlock()
		return consentID, nil
	}
	a.mu.RUnlock()

	// Получаем клиент банка
	client, err := a.getClient(bankCode)
	if err != nil {
		return "", err
	}

	// Создаем consent
	permissions := []string{
		"ReadAccountsDetail",
		"ReadBalances",
		"ReadTransactionsDetail",
	}

	consent, err := client.CreateConsent(ctx, userID, permissions, "FinHelper aggregation service")
	if err != nil {
		return "", fmt.Errorf("create consent for %s: %w", bankCode, err)
	}

	if consent.ConsentID == "" {
		return "", fmt.Errorf("empty consent_id for bank %s", bankCode)
	}

	// Сохраняем в кэш
	a.mu.Lock()
	a.consentCache[cacheKey] = consent.ConsentID
	a.mu.Unlock()

	log.Printf("Created consent for bank=%s user=%s: %s", bankCode, userID, consent.ConsentID)
	return consent.ConsentID, nil
}

// GetConsentStatus получает статус согласия
func (a *BankAggregator) GetConsentStatus(ctx context.Context, bankCode, consentID string) (*ConsentResponse, error) {
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	return client.GetConsentStatus(ctx, consentID)
}

// RevokeConsent отзывает согласие
func (a *BankAggregator) RevokeConsent(ctx context.Context, bankCode, consentID string) error {
	client, err := a.getClient(bankCode)
	if err != nil {
		return err
	}

	// Удаляем из кэша
	a.mu.Lock()
	for key, id := range a.consentCache {
		if id == consentID {
			delete(a.consentCache, key)
			break
		}
	}
	a.mu.Unlock()

	return client.RevokeConsent(ctx, consentID)
}

// ACCOUNTS

// GetAccountsFromAllBanks получает счета из всех банков
func (a *BankAggregator) GetAccountsFromAllBanks(ctx context.Context, userID string) ([]Account, error) {
	var allAccounts []Account

	for _, bank := range a.config.Banks {
		accounts, err := a.GetAccountsFromBank(ctx, bank.Code, userID)
		if err != nil {
			log.Printf("Warning: failed to get accounts from %s: %v", bank.Code, err)
			continue // продолжаем с другими банками
		}
		allAccounts = append(allAccounts, accounts...)
	}

	log.Printf("Aggregated %d accounts from %d banks for user %s", len(allAccounts), len(a.config.Banks), userID)
	return allAccounts, nil
}

// GetAccountsFromBank получает счета из конкретного банка
func (a *BankAggregator) GetAccountsFromBank(ctx context.Context, bankCode, userID string) ([]Account, error) {
	// Получаем или создаем consent
	consentID, err := a.EnsureConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure consent: %w", err)
	}

	// Получаем клиент
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	// Получаем счета
	accountDetails, err := client.GetAccounts(ctx, consentID, userID)
	if err != nil {
		return nil, fmt.Errorf("get accounts from %s: %w", bankCode, err)
	}

	// Конвертируем в legacy формат
	var accounts []Account
	for _, detail := range accountDetails {
		accounts = append(accounts, detail.ToLegacyAccount(bankCode))
	}

	log.Printf("Fetched %d accounts from bank %s for user %s", len(accounts), bankCode, userID)
	return accounts, nil
}

// GetAccountBalances получает балансы для конкретного счета
func (a *BankAggregator) GetAccountBalances(ctx context.Context, bankCode, userID, accountID string) ([]BalanceDetail, error) {
	consentID, err := a.EnsureConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure consent: %w", err)
	}

	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	return client.GetBalances(ctx, consentID, accountID, userID)
}

// TRANSACTIONS

// GetTransactions получает транзакции из одного или всех банков
func (a *BankAggregator) GetTransactions(ctx context.Context, userID, bankFilter string, from, to *time.Time) ([]Transaction, error) {
	// Определяем список банков для запроса
	banks := a.config.Banks
	if bankFilter != "" && bankFilter != "all" {
		found := false
		for _, b := range a.config.Banks {
			if b.Code == bankFilter {
				banks = []Bank{b}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown bank: %s", bankFilter)
		}
	}

	var allTransactions []Transaction

	// Получаем транзакции из каждого банка
	for _, bank := range banks {
		txs, err := a.getTransactionsFromBank(ctx, bank.Code, userID, from, to)
		if err != nil {
			log.Printf("Warning: failed to get transactions from %s: %v", bank.Code, err)
			continue
		}
		allTransactions = append(allTransactions, txs...)
	}

	// Дополнительная фильтрация по датам (на клиенте)
	if from != nil || to != nil {
		filtered := make([]Transaction, 0)
		for _, tx := range allTransactions {
			if from != nil && tx.Date.Before(*from) {
				continue
			}
			if to != nil && tx.Date.After(*to) {
				continue
			}
			filtered = append(filtered, tx)
		}
		allTransactions = filtered
	}

	log.Printf("Aggregated %d transactions from %d banks for user %s", len(allTransactions), len(banks), userID)
	return allTransactions, nil
}

// getTransactionsFromBank получает транзакции из конкретного банка
func (a *BankAggregator) getTransactionsFromBank(ctx context.Context, bankCode, userID string, from, to *time.Time) ([]Transaction, error) {
	// Получаем consent
	consentID, err := a.EnsureConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure consent: %w", err)
	}

	// Получаем клиент
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	// Сначала получаем список счетов
	accounts, err := client.GetAccounts(ctx, consentID, userID)
	if err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}

	var allTransactions []Transaction

	// Для каждого счета получаем транзакции
	for _, account := range accounts {
		var fromTime, toTime time.Time
		if from != nil {
			fromTime = *from
		}
		if to != nil {
			toTime = *to
		}

		txDetails, err := client.GetTransactions(ctx, consentID, account.AccountID, userID, fromTime, toTime)
		if err != nil {
			log.Printf("Warning: failed to get transactions for account %s: %v", account.AccountID, err)
			continue
		}

		// Конвертируем в legacy формат
		for _, detail := range txDetails {
			allTransactions = append(allTransactions, detail.ToLegacyTransaction(bankCode))
		}
	}

	return allTransactions, nil
}

// GetAccountTransactions получает транзакции конкретного счета
func (a *BankAggregator) GetAccountTransactions(ctx context.Context, bankCode, userID, accountID string, from, to time.Time) ([]Transaction, error) {
	consentID, err := a.EnsureConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure consent: %w", err)
	}

	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	txDetails, err := client.GetTransactions(ctx, consentID, accountID, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	// Конвертируем в legacy формат
	var transactions []Transaction
	for _, detail := range txDetails {
		transactions = append(transactions, detail.ToLegacyTransaction(bankCode))
	}

	return transactions, nil
}

// HELPERS

// getClient возвращает клиент для указанного банка
func (a *BankAggregator) getClient(bankCode string) (*BankAPIClient, error) {
	client, exists := a.clients[bankCode]
	if !exists {
		return nil, fmt.Errorf("unknown bank: %s", bankCode)
	}
	return client, nil
}

// GetBankByCode находит конфигурацию банка по коду
func (a *BankAggregator) GetBankByCode(code string) (Bank, error) {
	for _, bank := range a.config.Banks {
		if bank.Code == code {
			return bank, nil
		}
	}
	return Bank{}, fmt.Errorf("unknown bank: %s", code)
}

// PAYMENT CONSENT MANAGEMENT

// EnsurePaymentConsent создает payment consent если его нет, или возвращает существующий
func (a *BankAggregator) EnsurePaymentConsent(ctx context.Context, bankCode, userID string, paymentInfo PaymentInfo) (string, error) {
	cacheKey := bankCode + "|" + userID

	// Проверяем кэш
	a.mu.RLock()
	if consentID, exists := a.paymentConsentCache[cacheKey]; exists {
		a.mu.RUnlock()
		return consentID, nil
	}
	a.mu.RUnlock()

	// Получаем клиент банка
	client, err := a.getClient(bankCode)
	if err != nil {
		return "", err
	}

	// Создаем payment consent
	req := PaymentConsentRequest{
		RequestingBank: a.config.TeamID,
		ClientID:       userID,
		PaymentDetails: paymentInfo,
		Reason:         "FinHelper payment service",
		AutoApproved:   true,
	}

	consent, err := client.CreatePaymentConsent(ctx, req)
	if err != nil {
		return "", fmt.Errorf("create payment consent for %s: %w", bankCode, err)
	}

	if consent.ConsentID == "" {
		return "", fmt.Errorf("empty payment consent_id for bank %s", bankCode)
	}

	// Сохраняем в кэш
	a.mu.Lock()
	a.paymentConsentCache[cacheKey] = consent.ConsentID
	a.mu.Unlock()

	log.Printf("Created payment consent for bank=%s user=%s: %s", bankCode, userID, consent.ConsentID)
	return consent.ConsentID, nil
}

// GetPaymentConsentStatus получает статус payment consent
func (a *BankAggregator) GetPaymentConsentStatus(ctx context.Context, bankCode, consentID string) (*PaymentConsentResponse, error) {
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	return client.GetPaymentConsentStatus(ctx, consentID)
}

// PAYMENTS

// CreatePayment создает платеж в указанном банке
func (a *BankAggregator) CreatePayment(ctx context.Context, bankCode, userID string, req PaymentRequest) (*PaymentResponse, error) {
	// Создаем payment consent
	paymentInfo := PaymentInfo{
		DebtorAccount:   req.DebtorAccount,
		CreditorAccount: req.CreditorAccount,
		Amount:          req.Amount,
		Reference:       req.Reference,
	}

	consentID, err := a.EnsurePaymentConsent(ctx, bankCode, userID, paymentInfo)
	if err != nil {
		return nil, fmt.Errorf("ensure payment consent: %w", err)
	}

	// Получаем клиент
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	// Создаем платеж
	payment, err := client.CreatePayment(ctx, consentID, userID, req)
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	log.Printf("Created payment %s for bank=%s user=%s", payment.PaymentID, bankCode, userID)
	return payment, nil
}

// GetPaymentStatus получает статус платежа
func (a *BankAggregator) GetPaymentStatus(ctx context.Context, bankCode, paymentID, userID string) (*PaymentResponse, error) {
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	return client.GetPaymentStatus(ctx, paymentID, userID)
}

// PRODUCT AGREEMENT CONSENT MANAGEMENT

// EnsureProductAgreementConsent создает PA consent если его нет, или возвращает существующий
func (a *BankAggregator) EnsureProductAgreementConsent(ctx context.Context, bankCode, userID string) (string, error) {
	cacheKey := bankCode + "|" + userID

	// Проверяем кэш
	a.mu.RLock()
	if consentID, exists := a.paConsentCache[cacheKey]; exists {
		a.mu.RUnlock()
		return consentID, nil
	}
	a.mu.RUnlock()

	// Получаем клиент банка
	client, err := a.getClient(bankCode)
	if err != nil {
		return "", err
	}

	// Создаем PA consent
	permissions := []string{
		"ReadProducts",
		"ReadAgreements",
		"CreateAgreement",
		"CloseAgreement",
	}

	req := ProductAgreementConsentRequest{
		RequestingBank: a.config.TeamID,
		ClientID:       userID,
		Permissions:    permissions,
		Reason:         "FinHelper product agreement service",
		AutoApproved:   true,
	}

	consent, err := client.CreateProductAgreementConsent(ctx, req)
	if err != nil {
		return "", fmt.Errorf("create PA consent for %s: %w", bankCode, err)
	}

	if consent.ConsentID == "" {
		return "", fmt.Errorf("empty PA consent_id for bank %s", bankCode)
	}

	// Сохраняем в кэш
	a.mu.Lock()
	a.paConsentCache[cacheKey] = consent.ConsentID
	a.mu.Unlock()

	log.Printf("Created PA consent for bank=%s user=%s: %s", bankCode, userID, consent.ConsentID)
	return consent.ConsentID, nil
}

// GetProductAgreementConsentStatus получает статус PA consent
func (a *BankAggregator) GetProductAgreementConsentStatus(ctx context.Context, bankCode, consentID string) (*ProductAgreementConsentResponse, error) {
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	return client.GetProductAgreementConsentStatus(ctx, consentID)
}

// PRODUCTS

// GetProducts получает список продуктов из банка
func (a *BankAggregator) GetProducts(ctx context.Context, bankCode, userID, productType string) ([]Product, error) {
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	products, err := client.GetProducts(ctx, userID, productType)
	if err != nil {
		return nil, fmt.Errorf("get products from %s: %w", bankCode, err)
	}

	log.Printf("Fetched %d products from bank %s", len(products), bankCode)
	return products, nil
}

// AGREEMENTS

// OpenAgreement открывает договор (вклад/кредит/карта)
func (a *BankAggregator) OpenAgreement(ctx context.Context, bankCode, userID string, req AgreementRequest) (*AgreementResponse, error) {
	// Получаем или создаем PA consent
	paConsentID, err := a.EnsureProductAgreementConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure PA consent: %w", err)
	}

	// Получаем клиент
	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	// Открываем договор
	agreement, err := client.OpenAgreement(ctx, paConsentID, userID, req)
	if err != nil {
		return nil, fmt.Errorf("open agreement: %w", err)
	}

	log.Printf("Opened agreement %s for bank=%s user=%s", agreement.AgreementID, bankCode, userID)
	return agreement, nil
}

// GetAgreementDetails получает детали договора
func (a *BankAggregator) GetAgreementDetails(ctx context.Context, bankCode, agreementID, userID string) (*AgreementResponse, error) {
	paConsentID, err := a.EnsureProductAgreementConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure PA consent: %w", err)
	}

	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	return client.GetAgreementDetails(ctx, paConsentID, agreementID, userID)
}

// CloseAgreement закрывает договор
func (a *BankAggregator) CloseAgreement(ctx context.Context, bankCode, agreementID, userID string) (*AgreementResponse, error) {
	paConsentID, err := a.EnsureProductAgreementConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure PA consent: %w", err)
	}

	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	agreement, err := client.CloseAgreement(ctx, paConsentID, agreementID, userID)
	if err != nil {
		return nil, fmt.Errorf("close agreement: %w", err)
	}

	log.Printf("Closed agreement %s for bank=%s user=%s", agreementID, bankCode, userID)
	return agreement, nil
}

// GetAgreements получает список договоров клиента
func (a *BankAggregator) GetAgreements(ctx context.Context, bankCode, userID string) ([]AgreementResponse, error) {
	paConsentID, err := a.EnsureProductAgreementConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure PA consent: %w", err)
	}

	client, err := a.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	agreements, err := client.GetAgreements(ctx, paConsentID, userID)
	if err != nil {
		return nil, fmt.Errorf("get agreements from %s: %w", bankCode, err)
	}

	log.Printf("Fetched %d agreements from bank %s for user %s", len(agreements), bankCode, userID)
	return agreements, nil
}