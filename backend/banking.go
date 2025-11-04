package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BankClient управляет подключениями ко всем банкам
type BankClient struct {
	cfg          Config
	mu           sync.RWMutex
	clients      map[string]BankAPI
	consentCache map[string]string // key: bank|user -> consentID
}

// NewBankClient создает новый агрегатор банковских клиентов
func NewBankClient(cfg Config) *BankClient {
	bc := &BankClient{
		cfg:          cfg,
		clients:      make(map[string]BankAPI),
		consentCache: make(map[string]string),
	}

	// Инициализируем клиентов для каждого банка
	for _, bank := range cfg.Banks {
		bc.clients[bank.Code] = NewVBankClient(
			bank.BaseURL,
			cfg.TeamID,
			cfg.ClientSecret,
			cfg.TeamID,
		)
	}

	return bc
}

// getClient возвращает клиент для указанного банка
func (c *BankClient) getClient(bankCode string) (BankAPI, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client, ok := c.clients[bankCode]
	if !ok {
		return nil, fmt.Errorf("unknown bank: %s", bankCode)
	}
	return client, nil
}

// bankByCode находит конфигурацию банка по коду
func (c *BankClient) bankByCode(code string) (Bank, error) {
	for _, b := range c.cfg.Banks {
		if b.Code == code {
			return b, nil
		}
	}
	return Bank{}, fmt.Errorf("unknown bank: %s", code)
}

// EnsureConsent создает согласие если его еще нет, или возвращает существующее
func (c *BankClient) EnsureConsent(ctx context.Context, bankCode, userID string) (string, error) {
	cacheKey := bankCode + "|" + userID

	// Проверяем кэш
	c.mu.RLock()
	if consentID, ok := c.consentCache[cacheKey]; ok {
		c.mu.RUnlock()
		return consentID, nil
	}
	c.mu.RUnlock()

	// Получаем клиент банка
	client, err := c.getClient(bankCode)
	if err != nil {
		return "", err
	}

	// Создаем согласие
	permissions := []string{
		"ReadAccountsDetail",
		"ReadBalances",
		"ReadTransactionsDetail",
	}

	consent, err := client.CreateConsent(ctx, userID, permissions, "FinHelper aggregation")
	if err != nil {
		return "", fmt.Errorf("create consent: %w", err)
	}

	if consent.ConsentID == "" {
		return "", fmt.Errorf("empty consent_id in response")
	}

	// Сохраняем в кэш
	c.mu.Lock()
	c.consentCache[cacheKey] = consent.ConsentID
	c.mu.Unlock()

	return consent.ConsentID, nil
}

// GetConsentStatus получает статус согласия
func (c *BankClient) GetConsentStatus(ctx context.Context, bankCode, consentID string) (*ConsentResponse, error) {
	client, err := c.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	return client.GetConsentStatus(ctx, consentID)
}

// RevokeConsent отзывает согласие
func (c *BankClient) RevokeConsent(ctx context.Context, bankCode, consentID string) error {
	client, err := c.getClient(bankCode)
	if err != nil {
		return err
	}

	// Удаляем из кэша
	c.mu.Lock()
	for key, id := range c.consentCache {
		if id == consentID {
			delete(c.consentCache, key)
			break
		}
	}
	c.mu.Unlock()

	return client.RevokeConsent(ctx, consentID)
}

// FetchAccountsAllBanks получает счета из всех подключенных банков
func (c *BankClient) FetchAccountsAllBanks(ctx context.Context, userID string) ([]Account, error) {
	var allAccounts []Account

	for _, bank := range c.cfg.Banks {
		// Создаем/получаем согласие
		consentID, err := c.EnsureConsent(ctx, bank.Code, userID)
		if err != nil {
			return nil, fmt.Errorf("ensure consent for %s: %w", bank.Code, err)
		}

		// Получаем клиент
		client, err := c.getClient(bank.Code)
		if err != nil {
			return nil, err
		}

		// Получаем счета
		accounts, err := client.GetAccounts(ctx, consentID, userID)
		if err != nil {
			return nil, fmt.Errorf("get accounts from %s: %w", bank.Code, err)
		}

		// Конвертируем в legacy формат
		for _, acc := range accounts {
			allAccounts = append(allAccounts, acc.ToLegacyAccount(bank.Code))
		}
	}

	return allAccounts, nil
}

// FetchAccountsFromBank получает счета из конкретного банка
func (c *BankClient) FetchAccountsFromBank(ctx context.Context, bankCode, userID string) ([]Account, error) {
	// Создаем/получаем согласие
	consentID, err := c.EnsureConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure consent: %w", err)
	}

	// Получаем клиент
	client, err := c.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	// Получаем счета
	accounts, err := client.GetAccounts(ctx, consentID, userID)
	if err != nil {
		return nil, fmt.Errorf("get accounts: %w", err)
	}

	// Конвертируем в legacy формат
	var result []Account
	for _, acc := range accounts {
		result = append(result, acc.ToLegacyAccount(bankCode))
	}

	return result, nil
}

// FetchBalances получает балансы для конкретного счета
func (c *BankClient) FetchBalances(ctx context.Context, bankCode, userID, accountID string) ([]BalanceDetail, error) {
	// Создаем/получаем согласие
	consentID, err := c.EnsureConsent(ctx, bankCode, userID)
	if err != nil {
		return nil, fmt.Errorf("ensure consent: %w", err)
	}

	// Получаем клиент
	client, err := c.getClient(bankCode)
	if err != nil {
		return nil, err
	}

	// Получаем балансы
	balances, err := client.GetBalances(ctx, consentID, accountID)
	if err != nil {
		return nil, fmt.Errorf("get balances: %w", err)
	}

	return balances, nil
}

// FetchTransactions получает транзакции со всех счетов или из конкретного банка
func (c *BankClient) FetchTransactions(ctx context.Context, userID, bankFilter string, from, to *time.Time) ([]Transaction, error) {
	var allTransactions []Transaction

	// Определяем, из каких банков получать данные
	banks := c.cfg.Banks
	if bankFilter != "" && bankFilter != "all" {
		found := false
		for _, b := range c.cfg.Banks {
			if b.Code == bankFilter {
				banks = []Bank{b}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown bank filter: %s", bankFilter)
		}
	}

	// Получаем транзакции из каждого банка
	for _, bank := range banks {
		// Создаем/получаем согласие
		consentID, err := c.EnsureConsent(ctx, bank.Code, userID)
		if err != nil {
			return nil, fmt.Errorf("ensure consent for %s: %w", bank.Code, err)
		}

		// Получаем клиент
		client, err := c.getClient(bank.Code)
		if err != nil {
			return nil, err
		}

		// Сначала получаем счета
		accounts, err := client.GetAccounts(ctx, consentID, userID)
		if err != nil {
			return nil, fmt.Errorf("get accounts from %s: %w", bank.Code, err)
		}

		// Для каждого счета получаем транзакции
		for _, account := range accounts {
			var fromTime, toTime time.Time
			if from != nil {
				fromTime = *from
			}
			if to != nil {
				toTime = *to
			}

			transactions, err := client.GetTransactions(ctx, consentID, account.AccountID, fromTime, toTime)
			if err != nil {
				// Логируем ошибку, но продолжаем для других счетов
				fmt.Printf("Warning: failed to get transactions for account %s: %v\n", account.AccountID, err)
				continue
			}

			// Конвертируем в legacy формат
			for _, tx := range transactions {
				allTransactions = append(allTransactions, tx.ToLegacyTransaction(bank.Code))
			}
		}
	}

	// Фильтрация по датам (дополнительная проверка на клиенте)
	if from != nil || to != nil {
		filtered := allTransactions[:0]
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

	return allTransactions, nil
}