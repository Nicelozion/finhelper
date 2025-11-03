package main

import (
	"log"
	"os"
	"strings"
)

type Bank struct {
	Code    string
	BaseURL string
}

type Config struct {
	TeamID       string
	ClientSecret string
	Banks        []Bank
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" { log.Fatalf("required env %s", k) }
	return v
}
func env(k, def string) string {
	v := os.Getenv(k); if v == "" { return def }; return v
}

func parseBanks() []Bank {
	var out []Bank
	for _, code := range strings.Split(env("BANKS", ""), ",") {
		code = strings.TrimSpace(code)
		if code == "" { continue }
		envKey := "BASE_URL_" + strings.ToUpper(code)
		out = append(out, Bank{Code: code, BaseURL: mustEnv(envKey)})
	}
	if len(out) == 0 { log.Fatal("BANKS is empty (e.g. vbank,abank)") }
	return out
}
