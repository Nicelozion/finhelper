package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// FlexibleTime поддерживает парсинг времени в разных форматах
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON парсит время в нескольких форматах
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if s == "null" || s == "" {
		ft.Time = time.Time{}
		return nil
	}

	// Попробуем несколько форматов
	formats := []string{
		time.RFC3339Nano,                    // 2006-01-02T15:04:05.999999999Z07:00
		time.RFC3339,                        // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05.999999999",     // Без timezone
		"2006-01-02T15:04:05.999999",        // Без timezone (6 цифр)
		"2006-01-02T15:04:05",               // Без timezone и милисекунд
		"2006-01-02",                        // Только дата
	}

	var err error
	for _, format := range formats {
		ft.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return err
}

// MarshalJSON сериализует время в RFC3339Nano формат
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	if ft.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(ft.Format(time.RFC3339Nano))
}

// REQUEST/RESPONSE MODELS

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
	Status         string       `json:"status"`
	ConsentID      string       `json:"consent_id"`
	ClientID       string       `json:"client_id,omitempty"`
	Permissions    []string     `json:"permissions,omitempty"`
	ExpirationDate FlexibleTime `json:"expiration_date,omitempty"`
	CreatedAt      FlexibleTime `json:"created_at,omitempty"`
	UpdatedAt      FlexibleTime `json:"updated_at,omitempty"`
	Reason         string       `json:"reason,omitempty"`
}

// TokenResponse представляет ответ с токеном
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	ClientID    string `json:"client_id,omitempty"`
}

// ACCOUNT MODELS

// AccountDetail представляет детальную информацию о счете
type AccountDetail struct {
	AccountID   string `json:"accountId"`           // camelCase для банковского API
	Status      string `json:"status,omitempty"`
	Currency    string `json:"currency"`
	AccountType string `json:"accountType"`         // camelCase для банковского API
	Nickname    string `json:"nickname,omitempty"`
	Servicer    struct {
		SchemeName     string `json:"schemeName,omitempty"`     // camelCase
		Identification string `json:"identification,omitempty"`
	} `json:"servicer,omitempty"`
	Account []struct {
		SchemeName     string `json:"schemeName,omitempty"`     // camelCase
		Identification string `json:"identification,omitempty"`
		Name           string `json:"name,omitempty"`
	} `json:"account,omitempty"`
}

// AccountsWrapper обертка для разных форматов ответа со счетами
type AccountsWrapper struct {
	Accounts []AccountDetail `json:"accounts,omitempty"` // множественное число
	Account  []AccountDetail `json:"account,omitempty"`  // единственное число (для совместимости)
	Data     struct {
		Accounts []AccountDetail `json:"accounts,omitempty"` // множественное число
		Account  []AccountDetail `json:"account,omitempty"`  // единственное число
	} `json:"data,omitempty"`
}

// BALANCE MODELS

// BalanceDetail представляет информацию о балансе счета
type BalanceDetail struct {
	AccountID            string `json:"accountId,omitempty"`            // camelCase
	CreditDebitIndicator string `json:"creditDebitIndicator"`           // camelCase
	Type                 string `json:"type"`
	DateTime             string `json:"dateTime"`                       // camelCase
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
	} `json:"creditLine,omitempty"`                                      // camelCase
}

// BalancesWrapper обертка для разных форматов ответа с балансами
type BalancesWrapper struct {
	Balances []BalanceDetail `json:"balances,omitempty"` // множественное число
	Balance  []BalanceDetail `json:"balance,omitempty"`  // единственное число (для совместимости)
	Data     struct {
		Balances []BalanceDetail `json:"balances,omitempty"` // множественное число
		Balance  []BalanceDetail `json:"balance,omitempty"`  // единственное число
	} `json:"data,omitempty"`
}

// TRANSACTION MODELS

// TransactionDetail представляет детальную информацию о транзакции
type TransactionDetail struct {
	AccountID            string       `json:"account_id,omitempty"`
	TransactionID        string       `json:"transaction_id"`
	TransactionReference string       `json:"transaction_reference,omitempty"`
	Amount               AmountObj    `json:"amount"`
	CreditDebitIndicator string       `json:"credit_debit_indicator"`
	Status               string       `json:"status"`
	BookingDateTime      FlexibleTime `json:"booking_date_time"`
	ValueDateTime        FlexibleTime `json:"value_date_time,omitempty"`
	TransactionInformation string     `json:"transaction_information,omitempty"`
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

// LEGACY MODELS (для обратной совместимости с фронтом)

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

// CONVERSION HELPERS

// ToLegacyAccount конвертирует AccountDetail в упрощенную модель
func (ad *AccountDetail) ToLegacyAccount(bank string) Account {
	extID := ""
	owner := ""
	if len(ad.Account) > 0 {
		extID = ad.Account[0].Identification
		owner = ad.Account[0].Name
	}

	return Account{
		ID:       ad.AccountID,
		ExtID:    extID,
		Bank:     bank,
		Type:     ad.AccountType,
		Currency: ad.Currency,
		Balance:  0, // баланс получаем отдельным запросом
		Owner:    owner,
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
		Date:        td.BookingDateTime.Time,
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

// PAYMENT CONSENT MODELS

// PaymentConsentRequest запрос на создание согласия для платежа
type PaymentConsentRequest struct {
	RequestingBank string      `json:"requesting_bank"`
	ClientID       string      `json:"client_id"`
	PaymentDetails PaymentInfo `json:"payment_details"`
	Reason         string      `json:"reason,omitempty"`
	AutoApproved   bool        `json:"auto_approved"`
}

// PaymentInfo информация о платеже для consent
type PaymentInfo struct {
	DebtorAccount  AccountInfo `json:"debtor_account"`
	CreditorAccount AccountInfo `json:"creditor_account"`
	Amount         AmountObj   `json:"amount"`
	Reference      string      `json:"reference,omitempty"`
}

// AccountInfo информация о счёте
type AccountInfo struct {
	SchemeName     string `json:"scheme_name"`
	Identification string `json:"identification"`
	Name           string `json:"name,omitempty"`
}

// PaymentConsentResponse ответ с информацией о payment consent
type PaymentConsentResponse struct {
	ConsentID      string       `json:"consent_id"`
	Status         string       `json:"status"`
	ClientID       string       `json:"client_id,omitempty"`
	PaymentDetails PaymentInfo  `json:"payment_details,omitempty"`
	CreatedAt      FlexibleTime `json:"created_at,omitempty"`
	UpdatedAt      FlexibleTime `json:"updated_at,omitempty"`
	ExpirationDate FlexibleTime `json:"expiration_date,omitempty"`
}

// PAYMENT MODELS

// PaymentRequest запрос на создание платежа
type PaymentRequest struct {
	DebtorAccount   AccountInfo `json:"debtor_account"`
	CreditorAccount AccountInfo `json:"creditor_account"`
	Amount          AmountObj   `json:"amount"`
	Reference       string      `json:"reference,omitempty"`
	RemittanceInfo  string      `json:"remittance_information,omitempty"`
}

// PaymentResponse ответ с информацией о платеже
type PaymentResponse struct {
	PaymentID       string       `json:"payment_id"`
	Status          string       `json:"status"`
	DebtorAccount   AccountInfo  `json:"debtor_account,omitempty"`
	CreditorAccount AccountInfo  `json:"creditor_account,omitempty"`
	Amount          AmountObj    `json:"amount,omitempty"`
	Reference       string       `json:"reference,omitempty"`
	CreatedAt       FlexibleTime `json:"created_at,omitempty"`
	UpdatedAt       FlexibleTime `json:"updated_at,omitempty"`
}

// PRODUCT AGREEMENT CONSENT MODELS

// ProductAgreementConsentRequest запрос на создание PA consent
type ProductAgreementConsentRequest struct {
	RequestingBank string   `json:"requesting_bank"`
	ClientID       string   `json:"client_id"`
	Permissions    []string `json:"permissions"`
	Reason         string   `json:"reason,omitempty"`
	AutoApproved   bool     `json:"auto_approved"`
}

// ProductAgreementConsentResponse ответ с информацией о PA consent
type ProductAgreementConsentResponse struct {
	ConsentID      string       `json:"consent_id"`
	Status         string       `json:"status"`
	ClientID       string       `json:"client_id,omitempty"`
	Permissions    []string     `json:"permissions,omitempty"`
	CreatedAt      FlexibleTime `json:"created_at,omitempty"`
	UpdatedAt      FlexibleTime `json:"updated_at,omitempty"`
	ExpirationDate FlexibleTime `json:"expiration_date,omitempty"`
}

// PRODUCT MODELS

// Product представляет банковский продукт
type Product struct {
	ProductID   string `json:"product_id"`
	ProductType string `json:"product_type"` // DEPOSIT, LOAN, CARD
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Currency    string `json:"currency,omitempty"`
	InterestRate struct {
		Rate string `json:"rate,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"interest_rate,omitempty"`
	MinAmount string `json:"min_amount,omitempty"`
	MaxAmount string `json:"max_amount,omitempty"`
	Term      struct {
		Min  int    `json:"min,omitempty"`
		Max  int    `json:"max,omitempty"`
		Unit string `json:"unit,omitempty"` // DAYS, MONTHS, YEARS
	} `json:"term,omitempty"`
}

// ProductsWrapper обертка для списка продуктов
type ProductsWrapper struct {
	Products []Product `json:"products,omitempty"`
	Data     struct {
		Products []Product `json:"products,omitempty"`
	} `json:"data,omitempty"`
}

// AGREEMENT MODELS

// AgreementRequest запрос на открытие договора
type AgreementRequest struct {
	ProductID string    `json:"product_id"`
	ClientID  string    `json:"client_id"`
	Amount    AmountObj `json:"amount,omitempty"`
	Term      int       `json:"term,omitempty"`      // в единицах согласно продукту
	TermUnit  string    `json:"term_unit,omitempty"` // DAYS, MONTHS, YEARS
	AccountID string    `json:"account_id,omitempty"` // счёт для вклада/списания
}

// AgreementResponse ответ с информацией о договоре
type AgreementResponse struct {
	AgreementID   string       `json:"agreement_id"`
	ProductID     string       `json:"product_id,omitempty"`
	ProductType   string       `json:"product_type,omitempty"`
	Status        string       `json:"status"`
	ClientID      string       `json:"client_id,omitempty"`
	Amount        AmountObj    `json:"amount,omitempty"`
	InterestRate  string       `json:"interest_rate,omitempty"`
	Term          int          `json:"term,omitempty"`
	TermUnit      string       `json:"term_unit,omitempty"`
	StartDate     FlexibleTime `json:"start_date,omitempty"`
	EndDate       FlexibleTime `json:"end_date,omitempty"`
	AccountID     string       `json:"account_id,omitempty"`
	CreatedAt     FlexibleTime `json:"created_at,omitempty"`
	UpdatedAt     FlexibleTime `json:"updated_at,omitempty"`
}

// AgreementsWrapper обертка для списка договоров
type AgreementsWrapper struct {
	Agreements []AgreementResponse `json:"agreements,omitempty"`
	Data       struct {
		Agreements []AgreementResponse `json:"agreements,omitempty"`
	} `json:"data,omitempty"`
}