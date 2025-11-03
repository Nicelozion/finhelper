package main

import (
	"bytes"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "sync"
    "time"
)

type BankClient struct {
	cfg          Config
	httpc        *http.Client
	mu           sync.Mutex
	tokenCache   map[string]struct{ val string; exp time.Time }
	consentCache map[string]string 
}

func NewBankClient(cfg Config) *BankClient {
	return &BankClient{
		cfg: cfg,
		httpc: &http.Client{Timeout: 10 * time.Second},
		tokenCache: map[string]struct{ val string; exp time.Time }{},
		consentCache: map[string]string{},
	}
}

func (c *BankClient) bankByCode(code string) (Bank, error) {
	for _, b := range c.cfg.Banks { if b.Code == code { return b, nil } }
	return Bank{}, fmt.Errorf("unknown bank: %s", code)
}

func (c *BankClient) bankToken(b Bank) (string, error) {
	c.mu.Lock()
	if t, ok := c.tokenCache[b.Code]; ok && time.Now().Before(t.exp) { c.mu.Unlock(); return t.val, nil }
	c.mu.Unlock()

	body := map[string]string{"client_id": c.cfg.TeamID, "client_secret": c.cfg.ClientSecret}
	j, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, b.BaseURL+"/auth/bank-token", bytes.NewReader(j))
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpc.Do(req); if err != nil { return "", err }
	defer res.Body.Close()
	if res.StatusCode != 200 {
		bt, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("bank-token %s: %s %s", b.Code, res.Status, string(bt))
	}
	var tr struct{ AccessToken string `json:"access_token"`; ExpiresIn int64 `json:"expires_in"` }
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil { return "", err }
	c.mu.Lock()
	c.tokenCache[b.Code] = struct{ val string; exp time.Time }{tr.AccessToken, time.Now().Add(time.Duration(tr.ExpiresIn-60)*time.Second)}
	c.mu.Unlock()
	return tr.AccessToken, nil
}

func (c *BankClient) ensureConsent(b Bank, user string) (string, error) {
	key := b.Code + "|" + user
	c.mu.Lock(); if v, ok := c.consentCache[key]; ok { c.mu.Unlock(); return v, nil }; c.mu.Unlock()

	bt, err := c.bankToken(b); if err != nil { return "", err }
	payload := map[string]any{
		"client_id": user,
		"permissions": []string{"ReadAccountsDetail","ReadBalances","ReadTransactionsDetail"},
		"reason": "FinHelper aggregation",
		"requesting_bank_name": c.cfg.TeamID,
		"auto_approved": true,
	}
	j, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, b.BaseURL+"/account-consents/request", bytes.NewReader(j))
	req.Header.Set("Authorization", "Bearer "+bt)
	req.Header.Set("X-Requesting-Bank", c.cfg.TeamID)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpc.Do(req); if err != nil { return "", err }
	defer res.Body.Close()
	if res.StatusCode != 200 {
		bt, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("consent %s: %s %s", b.Code, res.Status, string(bt))
	}
	var cr struct {
    Status    string `json:"status"`
    ConsentID string `json:"consent_id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		return "", err
	}
	if cr.ConsentID == "" {
		return "", errors.New("empty consent_id")
	}
	c.mu.Lock()
	c.consentCache[key] = cr.ConsentID
	c.mu.Unlock()
	return cr.ConsentID, nil
}

func (c *BankClient) getWithConsent(b Bank, cid, path string) (*http.Response, error) {
	bt, err := c.bankToken(b); if err != nil { return nil, err }
	req, _ := http.NewRequest(http.MethodGet, b.BaseURL+path, nil)
	req.Header.Set("Authorization", "Bearer "+bt)
	req.Header.Set("X-Requesting-Bank", c.cfg.TeamID)
	req.Header.Set("X-Consent-Id", cid)
	return c.httpc.Do(req)
}

func (c *BankClient) FetchAccountsAllBanks(user string) ([]Account, error) {
	out := []Account{}
	for _, b := range c.cfg.Banks {
		cid, err := c.ensureConsent(b, user); if err != nil { return nil, err }
		res, err := c.getWithConsent(b, cid, "/accounts"); if err != nil { return nil, err }
		func() {
			defer res.Body.Close()
			if res.StatusCode != 200 { err = fmt.Errorf("accounts %s: %s", b.Code, res.Status); return }
			raw, _ := io.ReadAll(res.Body)

			var arr []map[string]any
			if e := json.Unmarshal(raw, &arr); e == nil {
				for _, it := range arr { out = append(out, mapAccount(b.Code, it)) }
				return
			}
			var obj map[string]any
			if e := json.Unmarshal(raw, &obj); e == nil {
				if v, ok := obj["accounts"].([]any); ok {
					for _, it := range v { if m, ok := it.(map[string]any); ok { out = append(out, mapAccount(b.Code, m)) } }
				}
			}
		}()
		if err != nil { return nil, err }
	}
	return out, nil
}

func mapAccount(bank string, m map[string]any) Account {
	id := str(m["id"]); if id == "" { id = str(m["accountId"]) }
	return Account{
		ID: id,
		ExtID: str(m["number"]),
		Bank: bank,
		Type: str(m["type"]),
		Currency: str(m["currency"]),
		Balance: num(m["balance"]),
		Owner: str(m["owner"]),
	}
}

func (c *BankClient) FetchTransactions(user, bankFilter string, from, to *time.Time) ([]Transaction, error) {
	accs, err := c.FetchAccountsAllBanks(user); if err != nil { return nil, err }
	out := []Transaction{}
	for _, a := range accs {
		if bankFilter != "" && bankFilter != "all" && a.Bank != bankFilter { continue }
		b, _ := c.bankByCode(a.Bank)
		cid, err := c.ensureConsent(b, user); if err != nil { return nil, err }
		res, err := c.getWithConsent(b, cid, "/accounts/"+a.ID+"/transactions"); if err != nil { return nil, err }
		func() {
			defer res.Body.Close()
			if res.StatusCode != 200 { err = fmt.Errorf("transactions %s: %s", a.Bank, res.Status); return }
			raw, _ := io.ReadAll(res.Body)
			var arr []map[string]any
			if e := json.Unmarshal(raw, &arr); e == nil {
				for _, it := range arr { out = append(out, mapTx(a.Bank, it)) }
				return
			}
			var obj map[string]any
			if e := json.Unmarshal(raw, &obj); e == nil {
				if v, ok := obj["transactions"].([]any); ok {
					for _, it := range v { if m, ok := it.(map[string]any); ok { out = append(out, mapTx(a.Bank, m)) } }
				}
			}
		}()
		if err != nil { return nil, err }
	}
	// фильтр дат локально
	if from != nil || to != nil {
		f := out[:0]
		for _, t := range out {
			ok := true
			if from != nil && t.Date.Before(*from) { ok = false }
			if to   != nil && t.Date.After(*to)   { ok = false }
			if ok { f = append(f, t) }
		}
		out = f
	}
	return out, nil
}

func mapTx(bank string, m map[string]any) Transaction {
	id := str(m["id"]); if id == "" { id = str(m["transactionId"]) }
	var dt time.Time
	if s := str(m["date"]); s != "" { dt, _ = time.Parse(time.RFC3339, s)
	} else if s := str(m["bookingDateTime"]); s != "" { dt, _ = time.Parse(time.RFC3339, s) }
	amt := num(m["amount"]); if amt == 0 { amt = num(m["transactionAmount"]) }
	return Transaction{
		ID: id, Date: dt, Amount: amt,
		Currency: str(m["currency"]), Merchant: str(m["merchant"]),
		Category: str(m["category"]), Description: str(m["description"]),
		Bank: bank,
	}
}

func str(v any) string {
	if v == nil { return "" }
	switch t := v.(type) {
	case string: return t
	default:
		b, _ := json.Marshal(t)
		var s string; _ = json.Unmarshal(b, &s); return s
	}
}
func num(v any) float64 {
	if v == nil { return 0 }
	switch t := v.(type) {
	case float64: return t
	case int: return float64(t)
	case int64: return float64(t)
	case json.Number: f, _ := t.Float64(); return f
	case string:
		var n json.Number; _ = json.Unmarshal([]byte(`"`+t+`"`), &n); f, _ := n.Float64(); return f
	default: return 0
	}
}
