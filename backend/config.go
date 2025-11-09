package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"github.com/joho/godotenv"
)

// Bank представляет конфигурацию банка
type Bank struct {
	Code    string
	BaseURL string
}

// Config содержит конфигурацию приложения
type Config struct {
	TeamID       string
	ClientSecret string
	Banks        []Bank
	CORSOrigin   string
	Port         string
}

// LoadConfig загружает конфигурацию из .env и переменных окружения
func LoadConfig() (Config, error) {
	// Загружаем .env файл (игнорируем ошибку если файла нет)
	_ = godotenv.Load()

	cfg := Config{
		TeamID:       mustEnv("TEAM_ID"),
		ClientSecret: mustEnv("CLIENT_SECRET"),
		CORSOrigin:   env("CORS_ORIGIN", "http://localhost:5173"),
		Port:         env("PORT", "8080"),
	}

	// Парсим банки
	banks, err := parseBanks()
	if err != nil {
		return Config{}, fmt.Errorf("parse banks: %w", err)
	}
	cfg.Banks = banks

	return cfg, nil
}

// parseBanks парсит конфигурацию банков из переменных окружения
func parseBanks() ([]Bank, error) {
	bankCodes := strings.Split(env("BANKS", "vbank,abank,sbank"), ",")
	
	var banks []Bank
	for _, code := range bankCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}

		// Формируем имя переменной окружения для URL банка
		envKey := "BASE_URL_" + strings.ToUpper(code)
		baseURL := os.Getenv(envKey)
		
		if baseURL == "" {
			log.Printf("Warning: %s not set, skipping bank %s", envKey, code)
			continue
		}

		// Проверяем формат URL
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			return nil, fmt.Errorf("invalid URL for bank %s: %s (must start with http:// or https://)", code, baseURL)
		}

		banks = append(banks, Bank{
			Code:    code,
			BaseURL: strings.TrimSuffix(baseURL, "/"), // убираем trailing slash
		})
	}

	if len(banks) == 0 {
		return nil, fmt.Errorf("no banks configured")
	}

	return banks, nil
}

// mustEnv возвращает значение переменной окружения или падает с ошибкой
func mustEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

// env возвращает значение переменной окружения или значение по умолчанию
func env(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}