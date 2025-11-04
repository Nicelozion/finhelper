package main

import (
	"log"
	"net/http"
)

func main() {
	cfg := Config{
		TeamID:       env("TEAM_ID", "team053"),
		ClientSecret: mustEnv("CLIENT_SECRET"),
		Banks:        parseBanks(),
	}

	log.Printf("Starting FinHelper backend with Team ID: %s", cfg.TeamID)
	log.Printf("Configured banks: %d", len(cfg.Banks))
	for _, bank := range cfg.Banks {
		log.Printf("  - %s: %s", bank.Code, bank.BaseURL)
	}

	srv := NewServer(cfg)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /healthz", srv.health)

	// Consent management (NEW)
	mux.HandleFunc("POST /api/consents", srv.createConsent)
	mux.HandleFunc("GET /api/consents/{id}", srv.getConsentStatus)
	mux.HandleFunc("DELETE /api/consents/{id}", srv.revokeConsent)

	// Legacy bank connection endpoint (обратная совместимость)
	mux.HandleFunc("POST /api/banks/{bank}/connect", srv.connectBank)

	// Account endpoints
	mux.HandleFunc("GET /api/accounts", srv.accounts)
	mux.HandleFunc("GET /api/accounts/{id}/balances", srv.accountBalances)
	mux.HandleFunc("GET /api/accounts/{id}/transactions", srv.accountTransactions)

	// Transaction endpoints
	mux.HandleFunc("GET /api/transactions", srv.transactions)

	// Apply middleware in correct order
	handler := withRecovery(
		withLogging(
			withTimeout(
				withRequestID(
					cors(mux),
				),
			),
		),
	)

	port := env("PORT", "8080")
	log.Printf("Backend listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}