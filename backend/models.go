package main

import (
	"strconv"
	"time"
)

// ============================================================================
// REQUEST/RESPONSE MODELS
// ============================================================================

// ConsentRequest представляет запрос на создание согласия
type ConsentRequest struct {
	RequestingBank string   `json:"requesting_bank"`
	ClientID       string   `json:"client_id"`
	Permissions    []string `json:"permissions"`
	Reason         string   `json:"reason"`
	AutoApproved   bool     `json:"auto_approved"`
}

// ConsentResponse представляет ответ с информацией о согласии
type ConsentResponse struct {
	Status         string    `json:"status"`
	ConsentID      string    `json:"consent_id"`
	ClientID       string    `json:"client_id,omitempty"`
	Permissions    []string  `json:"permissions,omitempty"`
	ExpirationDate time.Time `json:"expiration_date,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	Reason         string    `json:"reason,omitempty"`
}

// TokenResponse представляет ответ с токеном
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	ClientID    string `json:"client_id,omitempty"`
}

// ============================================================================
// ACCOUNT MODELS
// ============================================================================

// AccountDetail представляет детальную информацию о счете
type AccountDetail struct {
	AccountID   string `json:"account_id"`
	Currency    string `json:"currency"`
	AccountType string `json:"account_type"`
	Nickname    string `json:"nickname,omitempty"`
	Servicer    struct {
		SchemeName     string `json:"scheme_name,omitempty"`
		Identification string `json:"identification,omitempty"`
	} `json:"servicer,omitempty"`
	Account struct {
		SchemeName     string `json:"scheme_name,omitempty"`
		Identification string `json:"identification,omitempty"`
		Name           string `json:"name,omitempty"`
	} `json:"account,omitempty"`
}

// AccountsWrapper обертка для разных форматов ответа со счетами
type AccountsWrapper struct {
	Accounts []AccountDetail `json:"accounts,omitempty"` // множественное число
	Data     struct {
		Accounts []AccountDetail `json:"accounts,omitempty"` // множественное число
	} `json:"data,omitempty"`
}

// ============================================================================
// BALANCE MODELS
// ============================================================================

// BalanceDetail представляет информацию о балансе счета
type BalanceDetail struct {
	AccountID            string `json:"account_id,omitempty"`
	CreditDebitIndicator string `json:"credit_debit_indicator"`
	Type                 string `json:"type"`
	DateTime             string `json:"date_time"`
	Amount               struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"amount"`
	CreditLine []struct {
		Included bool   `json:"included"`
		Type     string `json:"type"`
		Amount   struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"amount"`
	} `json:"credit_line,omitempty"`
}

// BalancesWrapper обертка для разных форматов ответа с балансами
type BalancesWrapper struct {
	Balances []BalanceDetail `json:"balances,omitempty"` // множественное число
	Data     struct {
		Balances []BalanceDetail `json:"balances,omitempty"` // множественное число
	} `json:"data,omitempty"`
}

// ============================================================================
// TRANSACTION MODELS
// ============================================================================

// TransactionDetail представляет детальную информацию о транзакции
type TransactionDetail struct {
	AccountID            string    `json:"account_id,omitempty"`
	TransactionID        string    `json:"transaction_id"`
	TransactionReference string    `json:"transaction_reference,omitempty"`
	Amount               AmountObj `json:"amount"`
	CreditDebitIndicator string    `json:"credit_debit_indicator"`
	Status               string    `json:"status"`
	BookingDateTime      time.Time `json:"booking_date_time"`
	ValueDateTime        time.Time `json:"value_date_time,omitempty"`
	TransactionInformation string  `json:"transaction_information,omitempty"`
	BankTransactionCode  struct {
		Code        string `json:"code,omitempty"`
		SubCode     string `json:"sub_code,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"bank_transaction_code,omitempty"`
	ProprietaryBankTransactionCode struct {
		Code        string `json:"code,omitempty"`
		Issuer      string `json:"issuer,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"proprietary_bank_transaction_code,omitempty"`
	Balance struct {
		Amount               AmountObj `json:"amount"`
		CreditDebitIndicator string    `json:"credit_debit_indicator"`
		Type                 string    `json:"type"`
	} `json:"balance,omitempty"`
	MerchantDetails struct {
		MerchantName         string `json:"merchant_name,omitempty"`
		MerchantCategoryCode string `json:"merchant_category_code,omitempty"`
	} `json:"merchant_details,omitempty"`
	CreditorAccount struct {
		SchemeName     string `json:"scheme_name,omitempty"`
		Identification string `json:"identification,omitempty"`
		Name           string `json:"name,omitempty"`
	} `json:"creditor_account,omitempty"`
	DebtorAccount struct {
		SchemeName     string `json:"scheme_name,omitempty"`
		Identification string `json:"identification,omitempty"`
		Name           string `json:"name,omitempty"`
	} `json:"debtor_account,omitempty"`
}

// AmountObj представляет сумму с валютой
type AmountObj struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// TransactionsWrapper обертка для разных форматов ответа с транзакциями
type TransactionsWrapper struct {
	Transactions []TransactionDetail `json:"transactions,omitempty"` // множественное число
	Data         struct {
		Transactions []TransactionDetail `json:"transactions,omitempty"` // множественное число
	} `json:"data,omitempty"`
}

// ============================================================================
// LEGACY MODELS (для обратной совместимости с фронтом)
// ============================================================================

// Account упрощенная модель счета для фронтенда
type Account struct {
	ID       string  `json:"id"`
	ExtID    string  `json:"ext_id,omitempty"`
	Bank     string  `json:"bank"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Owner    string  `json:"owner,omitempty"`
}

// Transaction упрощенная модель транзакции для фронтенда
type Transaction struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Merchant    string    `json:"merchant,omitempty"`
	Category    string    `json:"category,omitempty"`
	Description string    `json:"description,omitempty"`
	Bank        string    `json:"bank"`
}

// ErrorResponse представляет ошибку API
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// ============================================================================
// CONVERSION HELPERS
// ============================================================================

// ToLegacyAccount конвертирует AccountDetail в упрощенную модель
func (ad *AccountDetail) ToLegacyAccount(bank string) Account {
	return Account{
		ID:       ad.AccountID,
		ExtID:    ad.Account.Identification,
		Bank:     bank,
		Type:     ad.AccountType,
		Currency: ad.Currency,
		Balance:  0, // баланс получаем отдельным запросом
		Owner:    ad.Account.Name,
	}
}

// ToLegacyTransaction конвертирует TransactionDetail в упрощенную модель
func (td *TransactionDetail) ToLegacyTransaction(bank string) Transaction {
	amount := parseAmount(td.Amount.Amount)
	
	// Если это дебет (списание), делаем сумму отрицательной
	if td.CreditDebitIndicator == "Debit" {
		amount = -amount
	}

	return Transaction{
		ID:          td.TransactionID,
		Date:        td.BookingDateTime,
		Amount:      amount,
		Currency:    td.Amount.Currency,
		Merchant:    td.MerchantDetails.MerchantName,
		Category:    td.ProprietaryBankTransactionCode.Description,
		Description: td.TransactionInformation,
		Bank:        bank,
	}
}

// parseAmount безопасно парсит строковую сумму в float64
func parseAmount(s string) float64 {
	if s == "" {
		return 0
	}
	
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	
	return f
}