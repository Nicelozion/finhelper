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

	// Применяем middleware в правильном порядке
	handler := ApplyMiddleware(mux, config.CORSOrigin)

	// Запускаем сервер
	addr := ":" + config.Port
	log.Printf("✓ Server listening on %s", addr)
	log.Println("✓ Ready to accept requests")
	log.Println()
	log.Println("Available endpoints:")
	log.Println("  GET  /healthz")
	log.Println("  POST /api/consents?bank=<bank>&user=<user>")
	log.Println("  GET  /api/consents/{id}?bank=<bank>")
	log.Println("  DELETE /api/consents/{id}?bank=<bank>")
	log.Println("  POST /api/banks/{bank}/connect?user=<user>")
	log.Println("  GET  /api/accounts?user=<user>&bank=<bank>")
	log.Println("  GET  /api/accounts/{id}/balances?bank=<bank>&user=<user>")
	log.Println("  GET  /api/accounts/{id}/transactions?bank=<bank>&user=<user>")
	log.Println("  GET  /api/transactions?user=<user>&bank=<bank>&from=<date>&to=<date>")
	log.Println()

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}