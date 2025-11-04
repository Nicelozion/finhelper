package main

import (
	"fmt"
	"time"
)

// LEGACY MODELS (для обратной совместимости)

type Account struct {
	ID       string  `json:"id"`
	ExtID    string  `json:"ext_id"`
	Bank     string  `json:"bank"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Owner    string  `json:"owner"`
}

type Transaction struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Merchant    string    `json:"merchant"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Bank        string    `json:"bank"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ============================================================================
// NEW MODELS (согласно OpenAPI спецификации)
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
	ClientID       string    `json:"client_id"`
	Permissions    []string  `json:"permissions"`
	ExpirationDate time.Time `json:"expiration_date,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	Reason         string    `json:"reason,omitempty"`
}

// AccountDetail представляет детальную информацию о счете из банковского API
type AccountDetail struct {
	AccountID   string `json:"account_id"`
	Currency    string `json:"currency"`
	AccountType string `json:"account_type"`
	Nickname    string `json:"nickname,omitempty"`
	Servicer    struct {
		SchemeName     string `json:"scheme_name"`
		Identification string `json:"identification"`
	} `json:"servicer,omitempty"`
	Account struct {
		SchemeName     string `json:"scheme_name"`
		Identification string `json:"identification"`
		Name           string `json:"name"`
	} `json:"account,omitempty"`
}

// BalanceDetail представляет информацию о балансе счета
type BalanceDetail struct {
	AccountID            string `json:"account_id,omitempty"`
	CreditDebitIndicator string `json:"credit_debit_indicator"` // Credit или Debit
	Type                 string `json:"type"`                   // InterimAvailable, InterimBooked, Expected и т.д.
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

// TransactionDetail представляет детальную информацию о транзакции
type TransactionDetail struct {
	AccountID            string `json:"account_id,omitempty"`
	TransactionID        string `json:"transaction_id"`
	TransactionReference string `json:"transaction_reference,omitempty"`
	Amount               struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"amount"`
	CreditDebitIndicator string    `json:"credit_debit_indicator"` // Credit или Debit
	Status               string    `json:"status"`                 // Booked, Pending
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
		Amount struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"amount"`
		CreditDebitIndicator string `json:"credit_debit_indicator"`
		Type                 string `json:"type"`
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

// ============================================================================
// HELPER CONVERSION FUNCTIONS
// ============================================================================

// ToLegacyAccount конвертирует AccountDetail в старую структуру Account
func (ad *AccountDetail) ToLegacyAccount(bank string) Account {
	balance := 0.0
	return Account{
		ID:       ad.AccountID,
		ExtID:    ad.Account.Identification,
		Bank:     bank,
		Type:     ad.AccountType,
		Currency: ad.Currency,
		Balance:  balance,
		Owner:    ad.Account.Name,
	}
}

// ToLegacyTransaction конвертирует TransactionDetail в старую структуру Transaction
func (td *TransactionDetail) ToLegacyTransaction(bank string) Transaction {
	amount := parseFloat(td.Amount.Amount)
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

// parseFloat безопасно парсит строку в float64
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}