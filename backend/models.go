package main

import "time"

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
