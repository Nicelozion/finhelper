package main

import (
	"net/http"
	"time"
)

type Server struct {
	cfg Config
	bc  *BankClient
}
func NewServer(cfg Config) *Server { return &Server{cfg: cfg, bc: NewBankClient(cfg)} }

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *Server) connectBank(w http.ResponseWriter, r *http.Request) {
	bank := r.PathValue("bank")
	switch bank { case "vbank","abank","sbank": default:
		writeErr(w, r, 400, "Invalid bank. Allowed: vbank, abank, sbank"); return }
	user := r.URL.Query().Get("user"); if user == "" { user = "demo-user-1" }
	b, err := s.bc.bankByCode(bank); if err != nil { writeErr(w, r, 400, err.Error()); return }
	cid, err := s.bc.ensureConsent(b, user); if err != nil { writeErr(w, r, 500, "Failed to connect: "+err.Error()); return }
	writeJSON(w, 200, map[string]any{"ok": true, "bank": bank, "consent_id": cid})
}

func (s *Server) accounts(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user"); if user == "" { user = "demo-user-1" }
	accs, err := s.bc.FetchAccountsAllBanks(user); if err != nil { writeErr(w, r, 500, "Failed to fetch accounts: "+err.Error()); return }
	if accs == nil { accs = []Account{} }
	writeJSON(w, 200, accs)
}

func (s *Server) transactions(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user"); if user == "" { user = "demo-user-1" }
	bank := r.URL.Query().Get("bank") 
	var fromPtr, toPtr *time.Time
	if v := r.URL.Query().Get("from"); v != "" { t, e := time.Parse(time.RFC3339, v); if e != nil { writeErr(w, r, 400, "Invalid 'from' (RFC3339)"); return }; fromPtr = &t }
	if v := r.URL.Query().Get("to");   v != "" { t, e := time.Parse(time.RFC3339, v); if e != nil { writeErr(w, r, 400, "Invalid 'to' (RFC3339)"); return }; toPtr = &t }
	if bank != "" && bank != "all" && bank != "vbank" && bank != "abank" && bank != "sbank" { writeErr(w, r, 400, "Invalid bank"); return }

	txs, err := s.bc.FetchTransactions(user, bank, fromPtr, toPtr); if err != nil { writeErr(w, r, 500, "Failed to fetch transactions: "+err.Error()); return }
	resp := make([]map[string]any, len(txs))
	for i, t := range txs {
		resp[i] = map[string]any{
			"id": t.ID, "date": t.Date.Format(time.RFC3339), "amount": t.Amount, "currency": t.Currency,
			"merchant": t.Merchant, "category": t.Category, "description": t.Description, "bank": t.Bank,
		}
	}
	writeJSON(w, 200, resp)
}
