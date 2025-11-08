package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("🚀 Starting FinHelper Banking Aggregator")

	// Загружаем конфигурацию (включая .env файл)
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("✓ Configuration loaded")
	log.Printf("  Team ID: %s", config.TeamID)
	log.Printf("  Configured banks: %d", len(config.Banks))
	for _, bank := range config.Banks {
		log.Printf("    - %s: %s", bank.Code, bank.BaseURL)
	}
	log.Printf("  CORS Origin: %s", config.CORSOrigin)
	log.Printf("  Port: %s", config.Port)

	// Создаем HTTP сервер
	server := NewServer(config)
	
	// Создаем роутер
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /health", server.handleHealth)

	// Consent management endpoints
	mux.HandleFunc("POST /api/consents", server.handleCreateConsent)
	mux.HandleFunc("GET /api/consents/{id}", server.handleGetConsentStatus)
	mux.HandleFunc("DELETE /api/consents/{id}", server.handleRevokeConsent)

	// Legacy bank connection endpoint (обратная совместимость)
	mux.HandleFunc("POST /api/banks/{bank}/connect", server.handleConnectBank)

	// Account endpoints
	mux.HandleFunc("GET /api/accounts", server.handleGetAccounts)
	mux.HandleFunc("GET /api/accounts/{id}/balances", server.handleGetAccountBalances)
	mux.HandleFunc("GET /api/accounts/{id}/transactions", server.handleGetAccountTransactions)

	// Transaction endpoints
	mux.HandleFunc("GET /api/transactions", server.handleGetTransactions)

	// Payment consent endpoints
	mux.HandleFunc("POST /api/payment-consents", server.handleCreatePaymentConsent)
	mux.HandleFunc("GET /api/payment-consents/{id}", server.handleGetPaymentConsentStatus)

	// Payment endpoints
	mux.HandleFunc("POST /api/payments", server.handleCreatePayment)
	mux.HandleFunc("GET /api/payments/{id}", server.handleGetPaymentStatus)

	// Product agreement consent endpoints
	mux.HandleFunc("POST /api/pa-consents", server.handleCreatePAConsent)
	mux.HandleFunc("GET /api/pa-consents/{id}", server.handleGetPAConsentStatus)

	// Product endpoints
	mux.HandleFunc("GET /api/products", server.handleGetProducts)

	// Agreement endpoints
	mux.HandleFunc("POST /api/agreements", server.handleOpenAgreement)
	mux.HandleFunc("GET /api/agreements", server.handleGetAgreements)
	mux.HandleFunc("GET /api/agreements/{id}", server.handleGetAgreementDetails)
	mux.HandleFunc("DELETE /api/agreements/{id}", server.handleCloseAgreement)

	// Применяем middleware в правильном порядке
	handler := ApplyMiddleware(mux, config.CORSOrigin)

	// Запускаем сервер
	addr := ":" + config.Port
	log.Printf("✓ Server listening on %s", addr)
	log.Println("✓ Ready to accept requests")
	log.Println()
	log.Println("Available endpoints:")
	log.Println("  GET  /healthz")
	log.Println()
	log.Println("Account Consents:")
	log.Println("  POST /api/consents?bank=<bank>&user=<user>")
	log.Println("  GET  /api/consents/{id}?bank=<bank>")
	log.Println("  DELETE /api/consents/{id}?bank=<bank>")
	log.Println()
	log.Println("Accounts & Transactions:")
	log.Println("  POST /api/banks/{bank}/connect?user=<user>")
	log.Println("  GET  /api/accounts?user=<user>&bank=<bank>")
	log.Println("  GET  /api/accounts/{id}/balances?bank=<bank>&user=<user>")
	log.Println("  GET  /api/accounts/{id}/transactions?bank=<bank>&user=<user>")
	log.Println("  GET  /api/transactions?user=<user>&bank=<bank>&from=<date>&to=<date>")
	log.Println()
	log.Println("Payment Consents:")
	log.Println("  POST /api/payment-consents?bank=<bank>&user=<user>")
	log.Println("  GET  /api/payment-consents/{id}?bank=<bank>")
	log.Println()
	log.Println("Payments:")
	log.Println("  POST /api/payments?bank=<bank>&user=<user>")
	log.Println("  GET  /api/payments/{id}?bank=<bank>&user=<user>")
	log.Println()
	log.Println("Product Agreement Consents:")
	log.Println("  POST /api/pa-consents?bank=<bank>&user=<user>")
	log.Println("  GET  /api/pa-consents/{id}?bank=<bank>")
	log.Println()
	log.Println("Products & Agreements:")
	log.Println("  GET  /api/products?bank=<bank>&user=<user>&type=<DEPOSIT|LOAN|CARD>")
	log.Println("  POST /api/agreements?bank=<bank>&user=<user>")
	log.Println("  GET  /api/agreements?bank=<bank>&user=<user>")
	log.Println("  GET  /api/agreements/{id}?bank=<bank>&user=<user>")
	log.Println("  DELETE /api/agreements/{id}?bank=<bank>&user=<user>")
	log.Println()

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}