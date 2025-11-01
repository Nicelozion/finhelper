# FinHelper Backend API

Инструкция по реализации backend API для интеграции с React frontend.

## 📋 Общая информация

- **Frontend URL**: `http://localhost:5173` (Vite dev server)
- **Backend URL**: `http://localhost:8080` (по умолчанию)
- **Content-Type**: `application/json`
- **Заголовок трассировки**: `X-Request-Id` (обязательно)

## 🔧 Критически важные требования

### 1. CORS настройки

Frontend работает на другом порту, поэтому **обязательно** нужно настроить CORS:

```go
import (
    "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
)

func setupCORS() http.Handler {
    cors := handlers.CORS(
        handlers.AllowedOrigins([]string{"http://localhost:5173"}),
        handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
        handlers.AllowedHeaders([]string{"Content-Type", "X-Request-Id"}),
        handlers.ExposedHeaders([]string{"X-Request-Id"}),
    )
    return cors
}

// В main.go:
r := mux.NewRouter()
// ... ваши routes ...
handler := setupCORS()(r)
http.ListenAndServe(":8080", handler)
```

**Альтернатива с стандартным net/http:**
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id")
        w.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### 2. X-Request-Id заголовок

**Важно**: Frontend отправляет заголовок `X-Request-Id` и ожидает его в ответе для трассировки ошибок.

```go
func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-Id")
        
        // Если заголовок не передан, генерируем новый UUID
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        // Возвращаем тот же requestId в ответе
        w.Header().Set("X-Request-Id", requestID)
        
        // Можно добавить в контекст для логирования
        ctx := context.WithValue(r.Context(), "requestID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

При ошибках также возвращайте `X-Request-Id`:

```go
func errorResponse(w http.ResponseWriter, message string, statusCode int, requestID string) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-Id", requestID)
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(map[string]string{
        "message": message,
        "error": http.StatusText(statusCode),
    })
}
```

## 📡 API Endpoints

### 1. Подключение банка

**Endpoint**: `POST /api/banks/{bank}/connect`

**Параметры пути:**
- `bank` - один из: `vbank`, `abank`, `sbank`

**Заголовки:**
```
Content-Type: application/json
X-Request-Id: <UUID>
```

**Валидация:**
- `bank` должен быть строго `vbank`, `abank` или `sbank`
- Если не соответствует → вернуть `400 Bad Request`

**Успешный ответ:** `200 OK`
```json
{
  "ok": true,
  "bank": "vbank",
  "consent_id": "consent-12345-abcde"
}
```

**Ошибка (неверный банк):** `400 Bad Request`
```json
{
  "message": "Invalid bank. Allowed values: vbank, abank, sbank"
}
```

**Ошибка (серверная):** `500 Internal Server Error`
```json
{
  "message": "Failed to connect to bank"
}
```

**Пример реализации на Go:**

```go
type ConnectBankRequest struct {
    // Тело запроса может быть пустым, все данные в URL
}

type ConnectBankResponse struct {
    OK        bool   `json:"ok"`
    Bank      string `json:"bank"`
    ConsentID string `json:"consent_id"`
}

func connectBankHandler(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-Id")
    
    vars := mux.Vars(r)
    bank := vars["bank"]
    
    // Валидация банка
    validBanks := map[string]bool{
        "vbank": true,
        "abank": true,
        "sbank": true,
    }
    
    if !validBanks[bank] {
        errorResponse(w, 
            "Invalid bank. Allowed values: vbank, abank, sbank", 
            http.StatusBadRequest, 
            requestID)
        return
    }
    
    // Ваша логика подключения банка
    consentID := connectToBank(bank) // ваша функция
    
    response := ConnectBankResponse{
        OK:        true,
        Bank:      bank,
        ConsentID: consentID,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-Id", requestID)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
```

---

### 2. Получение списка счетов

**Endpoint**: `GET /api/accounts`

**Заголовки:**
```
X-Request-Id: <UUID>
```

**Успешный ответ:** `200 OK`
```json
[
  {
    "id": "acc-001",
    "ext_id": "40817810099910004312",
    "bank": "vbank",
    "type": "current",
    "currency": "RUB",
    "balance": 12345.67,
    "owner": "Иванов Иван Иванович"
  },
  {
    "id": "acc-002",
    "ext_id": "40817810188820005321",
    "bank": "abank",
    "type": "savings",
    "currency": "RUB",
    "balance": 50000.00,
    "owner": "Петров Петр Петрович"
  }
]
```

**Пустой список:** `200 OK`
```json
[]
```

**Ошибка:** `500 Internal Server Error`
```json
{
  "message": "Failed to fetch accounts"
}
```

**Структуры Go:**

```go
type Account struct {
    ID       string  `json:"id"`
    ExtID    string  `json:"ext_id"`    // Внешний номер счета
    Bank     string  `json:"bank"`      // vbank, abank, sbank
    Type     string  `json:"type"`      // current, savings, etc.
    Currency string  `json:"currency"`  // RUB, USD, EUR
    Balance  float64 `json:"balance"`
    Owner    string  `json:"owner"`
}

func accountsHandler(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-Id")
    
    accounts, err := getAccounts() // ваша функция получения счетов
    if err != nil {
        errorResponse(w, "Failed to fetch accounts", http.StatusInternalServerError, requestID)
        return
    }
    
    // Если счетов нет, возвращаем пустой массив
    if accounts == nil {
        accounts = []Account{}
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-Id", requestID)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(accounts)
}
```

---

### 3. Получение операций/транзакций

**Endpoint**: `GET /api/transactions`

**Query параметры:**
- `from` (опционально) - дата начала в формате ISO-8601: `2024-01-01T00:00:00Z`
- `to` (опционально) - дата окончания в формате ISO-8601: `2024-01-31T23:59:59.999Z`
- `bank` (опционально) - фильтр по банку: `vbank`, `abank`, `sbank` (если `"all"` или отсутствует - все банки)

**Пример запроса:**
```
GET /api/transactions?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59.999Z&bank=vbank
```

**Валидация дат:**
- Формат: `RFC3339` / `ISO-8601`: `YYYY-MM-DDTHH:mm:ssZ` или `YYYY-MM-DDTHH:mm:ss.sssZ`
- Frontend отправляет: `from` с временем `00:00:00Z`, `to` с временем `23:59:59.999Z`
- Парсинг в Go: `time.Parse(time.RFC3339, dateStr)`

**Валидация bank:**
- Если `bank` присутствует и не равен `"all"`, должен быть `vbank`, `abank` или `sbank`
- Если не соответствует → вернуть `400 Bad Request`

**Успешный ответ:** `200 OK`
```json
[
  {
    "id": "tx-001",
    "date": "2024-01-15T10:30:00Z",
    "amount": -1500.00,
    "currency": "RUB",
    "merchant": "Магазин Электроника",
    "category": "shopping",
    "description": "Покупка наушников",
    "bank": "vbank"
  },
  {
    "id": "tx-002",
    "date": "2024-01-20T14:20:00Z",
    "amount": 50000.00,
    "currency": "RUB",
    "merchant": "",
    "category": "salary",
    "description": "Зарплата за январь",
    "bank": "abank"
  }
]
```

**Важно:**
- `amount` может быть отрицательным (расход) или положительным (доход)
- Frontend фильтрует: отрицательные = расходы, положительные = доходы
- Если операций нет → вернуть пустой массив `[]`

**Пустой ответ:** `200 OK`
```json
[]
```

**Ошибка валидации:** `400 Bad Request`
```json
{
  "message": "Invalid date format. Expected RFC3339 (e.g., 2024-01-01T00:00:00Z)"
}
```

**Структуры Go:**

```go
type Transaction struct {
    ID          string    `json:"id"`
    Date        time.Time `json:"date"`        // ISO-8601 / RFC3339
    Amount      float64   `json:"amount"`      // отрицательное = расход, положительное = доход
    Currency    string    `json:"currency"`
    Merchant    string    `json:"merchant"`    // может быть пустой строкой
    Category    string    `json:"category"`    // может быть пустой строкой
    Description string    `json:"description"` // может быть пустой строкой
    Bank        string    `json:"bank"`        // vbank, abank, sbank
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-Id")
    
    // Парсинг query параметров
    fromStr := r.URL.Query().Get("from")
    toStr := r.URL.Query().Get("to")
    bank := r.URL.Query().Get("bank")
    
    var fromTime, toTime *time.Time
    
    // Парсинг from
    if fromStr != "" {
        t, err := time.Parse(time.RFC3339, fromStr)
        if err != nil {
            errorResponse(w, 
                "Invalid 'from' date format. Expected RFC3339 (e.g., 2024-01-01T00:00:00Z)", 
                http.StatusBadRequest, 
                requestID)
            return
        }
        fromTime = &t
    }
    
    // Парсинг to
    if toStr != "" {
        t, err := time.Parse(time.RFC3339, toStr)
        if err != nil {
            errorResponse(w, 
                "Invalid 'to' date format. Expected RFC3339 (e.g., 2024-01-31T23:59:59.999Z)", 
                http.StatusBadRequest, 
                requestID)
            return
        }
        toTime = &t
    }
    
    // Валидация bank
    if bank != "" && bank != "all" {
        validBanks := map[string]bool{"vbank": true, "abank": true, "sbank": true}
        if !validBanks[bank] {
            errorResponse(w, 
                "Invalid bank. Allowed values: vbank, abank, sbank, all", 
                http.StatusBadRequest, 
                requestID)
            return
        }
    }
    
    // Ваша логика получения транзакций
    transactions, err := getTransactions(fromTime, toTime, bank)
    if err != nil {
        errorResponse(w, "Failed to fetch transactions", http.StatusInternalServerError, requestID)
        return
    }
    
    // Форматирование ответа с правильными датами
    response := make([]map[string]interface{}, len(transactions))
    for i, tx := range transactions {
        response[i] = map[string]interface{}{
            "id":          tx.ID,
            "date":        tx.Date.Format(time.RFC3339),
            "amount":      tx.Amount,
            "currency":    tx.Currency,
            "merchant":    tx.Merchant,
            "category":    tx.Category,
            "description": tx.Description,
            "bank":        tx.Bank,
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-Id", requestID)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
```

## 🔍 Валидация и обработка ошибок

### Общие правила валидации:

1. **Банк**: Только `vbank`, `abank`, `sbank` (регистр важен в URL)
2. **Даты**: Строго RFC3339 формат
3. **Обязательные поля**: Все поля в ответах должны присутствовать (даже если пустые строки)

### Коды ответов:

- `200 OK` - успешный запрос
- `400 Bad Request` - ошибка валидации (неверный формат данных, неверный банк)
- `500 Internal Server Error` - серверная ошибка
- `404 Not Found` - endpoint не найден (не используется в текущем API)

### Формат ошибок:

Все ошибки должны возвращать JSON:

```go
type ErrorResponse struct {
    Message string `json:"message"`
    Error   string `json:"error,omitempty"` // опционально
}
```

## 🧪 Тестирование интеграции

### Мини-тест план:

1. **Health check** (опционально, но рекомендуется):
   ```
   GET /healthz
   → 200 OK {"status": "ok"}
   ```

2. **Подключение банка:**
   ```bash
   curl -X POST http://localhost:8080/api/banks/vbank/connect \
     -H "Content-Type: application/json" \
     -H "X-Request-Id: test-123"
   ```
   Ожидается: `{"ok":true,"bank":"vbank","consent_id":"..."}`

3. **Получение счетов:**
   ```bash
   curl http://localhost:8080/api/accounts \
     -H "X-Request-Id: test-123"
   ```
   Ожидается: массив счетов (может быть пустым)

4. **Получение операций:**
   ```bash
   curl "http://localhost:8080/api/transactions?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59.999Z&bank=vbank" \
     -H "X-Request-Id: test-123"
   ```
   Ожидается: массив транзакций (может быть пустым)

5. **Проверка CORS:**
   ```bash
   curl -X OPTIONS http://localhost:8080/api/accounts \
     -H "Origin: http://localhost:5173" \
     -H "Access-Control-Request-Method: GET" \
     -v
   ```
   Ожидается: заголовки `Access-Control-Allow-Origin: http://localhost:5173`

6. **Проверка X-Request-Id:**
   ```bash
   curl http://localhost:8080/api/accounts \
     -H "X-Request-Id: my-custom-id" \
     -v
   ```
   Ожидается: `X-Request-Id: my-custom-id` в ответе

## 📝 Пример полной структуры роутера

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "time"
    
    "github.com/gorilla/mux"
    "github.com/gorilla/handlers"
)

func main() {
    r := mux.NewRouter()
    
    // API routes
    api := r.PathPrefix("/api").Subrouter()
    api.HandleFunc("/banks/{bank}/connect", connectBankHandler).Methods("POST")
    api.HandleFunc("/accounts", accountsHandler).Methods("GET")
    api.HandleFunc("/transactions", transactionsHandler).Methods("GET")
    
    // Health check (опционально)
    r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }).Methods("GET")
    
    // CORS middleware
    cors := handlers.CORS(
        handlers.AllowedOrigins([]string{"http://localhost:5173"}),
        handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
        handlers.AllowedHeaders([]string{"Content-Type", "X-Request-Id"}),
        handlers.ExposedHeaders([]string{"X-Request-Id"}),
    )
    
    // Request ID middleware
    handler := requestIDMiddleware(cors(r))
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}

func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-Id")
        if requestID == "" {
            requestID = generateUUID() // ваша функция генерации UUID
        }
        w.Header().Set("X-Request-Id", requestID)
        next.ServeHTTP(w, r)
    })
}
```

## ✅ Чеклист для backend разработчика

- [ ] CORS настроен для `http://localhost:5173`
- [ ] Все endpoints возвращают заголовок `X-Request-Id`
- [ ] `/api/banks/{bank}/connect` валидирует банк (vbank/abank/sbank)
- [ ] `/api/accounts` возвращает массив (даже если пустой)
- [ ] `/api/transactions` парсит даты в RFC3339 формате
- [ ] `/api/transactions` валидирует параметр `bank`
- [ ] Все ошибки возвращают JSON с полем `message`
- [ ] В ошибках возвращается заголовок `X-Request-Id`
- [ ] Пустые массивы возвращаются как `[]`, а не `null`
- [ ] Даты в ответах форматируются в RFC3339
- [ ] Health check endpoint (опционально, но рекомендуется)

## 🚀 Запуск и тестирование

1. Запустите backend на порту `8080`
2. Убедитесь, что CORS настроен правильно
3. Запустите frontend: `cd frontend && pnpm run dev`
4. Откройте `http://localhost:5173` в браузере
5. Проверьте все три страницы: Connect, Accounts, Transactions

## 📞 Контакты для согласования

При необходимости согласовать изменения в API форматах, свяжитесь с frontend разработчиком.

---

**Важно**: Эта документация синхронизирована с frontend. При изменении форматов API обновите эту документацию и сообщите команде.

