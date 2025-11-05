package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// Ваши credentials
	teamID := "team053"
	clientSecret := "EdmB1m5yQ0PjSqEuccSMhyxwq1fo5ITW"
	
	// ✅ URL с QUERY параметрами!
	bankURL := fmt.Sprintf("https://vbank.open.bankingapi.ru/auth/bank-token?client_id=%s&client_secret=%s",
		teamID, clientSecret)
	
	fmt.Printf("Request URL: %s\n", bankURL)
	fmt.Println("Request Method: POST")
	fmt.Println("Request Body: (empty)")
	fmt.Println()
	
	// Создаем запрос БЕЗ body
	req, err := http.NewRequest("POST", bankURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	
	// Заголовки
	req.Header.Set("Accept", "application/json")
	
	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}
	
	fmt.Printf("Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Headers:\n")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
	fmt.Printf("Response Body: %s\n", string(body))
	
	if resp.StatusCode == 200 {
		var tokenResp map[string]interface{}
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			return
		}
		
		fmt.Println("\n✅ SUCCESS!")
		fmt.Printf("Access Token: %v\n", tokenResp["access_token"])
		fmt.Printf("Token Type: %v\n", tokenResp["token_type"])
		fmt.Printf("Expires In: %v seconds\n", tokenResp["expires_in"])
	} else {
		fmt.Println("\n❌ FAILED!")
	}
}