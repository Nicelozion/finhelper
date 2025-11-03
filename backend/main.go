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
	srv := NewServer(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.health)
	mux.HandleFunc("POST /api/banks/{bank}/connect", srv.connectBank)
	mux.HandleFunc("GET /api/accounts", srv.accounts)
	mux.HandleFunc("GET /api/transactions", srv.transactions)

	handler := withTimeout(withRequestID(cors(mux)))
	log.Println("backend listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
