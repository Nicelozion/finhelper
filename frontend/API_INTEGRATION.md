# FinHelper Backend API Integration Guide

## Backend URL
```
http://localhost:8080
```

## Архитектура потоков

### 1. Account Flow (Счета)
```
1. Создать Account Consent
2. Получить список счетов
3. Получить балансы
4. Получить транзакции
```

### 2. Payment Flow (Платежи)
```
1. Создать Payment Consent
2. Создать платеж
3. Проверить статус платежа
```

### 3. Product Agreement Flow (Продукты/Договоры)
```
1. Создать PA Consent
2. Получить список продуктов
3. Открыть договор (вклад/кредит/карта)
4. Получить список договоров
5. Закрыть договор
```

---

## API Endpoints

### Health Check

#### GET `/healthz`
Проверка работоспособности сервера.

```javascript
const response = await fetch('http://localhost:8080/healthz');
const data = await response.json();
// { status: "ok", timestamp: "2025-11-08T20:00:00Z", banks: 3 }
```

---

## 1. Account Consents (Согласия на счета)

### POST `/api/consents`
Создать согласие на доступ к счетам клиента.

**Query Parameters:**
- `bank` - код банка (vbank, abank, sbank)
- `user` - ID пользователя (опционально, по умолчанию "demo-user-1")

**Request:**
```javascript
const response = await fetch('http://localhost:8080/api/consents?bank=vbank&user=user-123', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  }
});

const data = await response.json();
// {
//   ok: true,
//   bank: "vbank",
//   user: "user-123",
//   consent_id: "consent-abc123",
//   message: "Consent created successfully"
// }
```

### GET `/api/consents/{id}`
Получить статус согласия.

**Query Parameters:**
- `bank` - код банка

**Request:**
```javascript
const consentId = 'consent-abc123';
const response = await fetch(`http://localhost:8080/api/consents/${consentId}?bank=vbank`);
const data = await response.json();
// {
//   status: "Authorised",
//   consent_id: "consent-abc123",
//   permissions: ["ReadAccountsDetail", "ReadBalances", "ReadTransactionsDetail"],
//   expiration_date: "2025-12-08T20:00:00Z"
// }
```

### DELETE `/api/consents/{id}`
Отозвать согласие.

**Request:**
```javascript
const response = await fetch(`http://localhost:8080/api/consents/${consentId}?bank=vbank`, {
  method: 'DELETE'
});
const data = await response.json();
// { ok: true, message: "Consent revoked successfully" }
```

---

## 2. Accounts & Balances (Счета и балансы)

### POST `/api/banks/{bank}/connect`
Legacy endpoint - подключить банк (создает consent автоматически).

**Request:**
```javascript
const response = await fetch('http://localhost:8080/api/banks/vbank/connect?user=user-123', {
  method: 'POST'
});
const data = await response.json();
// { ok: true, bank: "vbank", consent_id: "...", user: "user-123" }
```

### GET `/api/accounts`
Получить список счетов.

**Query Parameters:**
- `user` - ID пользователя
- `bank` - код банка или "all" для всех банков

**Request:**
```javascript
// Счета из одного банка
const response = await fetch('http://localhost:8080/api/accounts?user=user-123&bank=vbank');
const accounts = await response.json();

// Счета из всех банков
const allResponse = await fetch('http://localhost:8080/api/accounts?user=user-123&bank=all');
const allAccounts = await allResponse.json();

// [
//   {
//     id: "acc-123",
//     bank: "vbank",
//     type: "Current",
//     currency: "RUB",
//     balance: 50000.00,
//     owner: "Иван Иванов"
//   }
// ]
```

### GET `/api/accounts/{id}/balances`
Получить балансы счета.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя

**Request:**
```javascript
const accountId = 'acc-123';
const response = await fetch(`http://localhost:8080/api/accounts/${accountId}/balances?bank=vbank&user=user-123`);
const balances = await response.json();

// [
//   {
//     credit_debit_indicator: "Credit",
//     type: "InterimAvailable",
//     date_time: "2025-11-08T20:00:00Z",
//     amount: {
//       amount: "50000.00",
//       currency: "RUB"
//     }
//   }
// ]
```

---

## 3. Transactions (Транзакции)

### GET `/api/accounts/{id}/transactions`
Получить транзакции конкретного счета.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя
- `from` - дата начала (YYYY-MM-DD)
- `to` - дата окончания (YYYY-MM-DD)

**Request:**
```javascript
const accountId = 'acc-123';
const response = await fetch(
  `http://localhost:8080/api/accounts/${accountId}/transactions?bank=vbank&user=user-123&from=2025-01-01&to=2025-12-31`
);
const transactions = await response.json();

// [
//   {
//     id: "tx-123",
//     date: "2025-11-01T10:30:00Z",
//     amount: -1500.00,
//     currency: "RUB",
//     merchant: "Магазин Пятёрочка",
//     category: "Покупки",
//     description: "Оплата покупок",
//     bank: "vbank"
//   }
// ]
```

### GET `/api/transactions`
Получить все транзакции пользователя (со всех счетов).

**Query Parameters:**
- `user` - ID пользователя
- `bank` - код банка или "all"
- `from` - дата начала (RFC3339)
- `to` - дата окончания (RFC3339)

**Request:**
```javascript
const response = await fetch(
  'http://localhost:8080/api/transactions?user=user-123&bank=all&from=2025-01-01T00:00:00Z&to=2025-12-31T23:59:59Z'
);
const transactions = await response.json();
```

---

## 4. Payment Consents (Согласия на платежи)

### POST `/api/payment-consents`
Создать согласие на выполнение платежа.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя

**Body:**
```json
{
  "debtor_account": {
    "scheme_name": "RU.CBR.PAN",
    "identification": "40817810099910004312",
    "name": "Иван Иванов"
  },
  "creditor_account": {
    "scheme_name": "RU.CBR.PAN",
    "identification": "40817810099910005423",
    "name": "Магазин ООО"
  },
  "amount": {
    "amount": "1500.00",
    "currency": "RUB"
  },
  "reference": "Payment for order #12345"
}
```

**Request:**
```javascript
const paymentInfo = {
  debtor_account: {
    scheme_name: "RU.CBR.PAN",
    identification: "40817810099910004312",
    name: "Иван Иванов"
  },
  creditor_account: {
    scheme_name: "RU.CBR.PAN",
    identification: "40817810099910005423",
    name: "Магазин ООО"
  },
  amount: {
    amount: "1500.00",
    currency: "RUB"
  },
  reference: "Payment for order #12345"
};

const response = await fetch('http://localhost:8080/api/payment-consents?bank=vbank&user=user-123', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(paymentInfo)
});

const data = await response.json();
// {
//   ok: true,
//   bank: "vbank",
//   user: "user-123",
//   consent_id: "payment-consent-123",
//   message: "Payment consent created successfully"
// }
```

### GET `/api/payment-consents/{id}`
Получить статус payment consent.

**Request:**
```javascript
const consentId = 'payment-consent-123';
const response = await fetch(`http://localhost:8080/api/payment-consents/${consentId}?bank=vbank`);
const data = await response.json();
```

---

## 5. Payments (Платежи)

### POST `/api/payments`
Создать платеж.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя

**Body:**
```json
{
  "debtor_account": {
    "scheme_name": "RU.CBR.PAN",
    "identification": "40817810099910004312",
    "name": "Иван Иванов"
  },
  "creditor_account": {
    "scheme_name": "RU.CBR.PAN",
    "identification": "40817810099910005423",
    "name": "Магазин ООО"
  },
  "amount": {
    "amount": "1500.00",
    "currency": "RUB"
  },
  "reference": "Payment for order #12345",
  "remittance_information": "Оплата заказа"
}
```

**Request:**
```javascript
const paymentRequest = {
  debtor_account: {
    scheme_name: "RU.CBR.PAN",
    identification: "40817810099910004312",
    name: "Иван Иванов"
  },
  creditor_account: {
    scheme_name: "RU.CBR.PAN",
    identification: "40817810099910005423",
    name: "Магазин ООО"
  },
  amount: {
    amount: "1500.00",
    currency: "RUB"
  },
  reference: "Payment for order #12345",
  remittance_information: "Оплата заказа"
};

const response = await fetch('http://localhost:8080/api/payments?bank=vbank&user=user-123', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(paymentRequest)
});

const payment = await response.json();
// {
//   payment_id: "payment-123",
//   status: "Pending",
//   debtor_account: {...},
//   creditor_account: {...},
//   amount: {...},
//   created_at: "2025-11-08T20:00:00Z"
// }
```

### GET `/api/payments/{id}`
Получить статус платежа.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя

**Request:**
```javascript
const paymentId = 'payment-123';
const response = await fetch(`http://localhost:8080/api/payments/${paymentId}?bank=vbank&user=user-123`);
const payment = await response.json();
// {
//   payment_id: "payment-123",
//   status: "AcceptedSettlementCompleted",
//   amount: {...}
// }
```

---

## 6. Product Agreement Consents (Согласия на продукты)

### POST `/api/pa-consents`
Создать согласие для работы с продуктами/договорами.

**Request:**
```javascript
const response = await fetch('http://localhost:8080/api/pa-consents?bank=vbank&user=user-123', {
  method: 'POST'
});

const data = await response.json();
// {
//   ok: true,
//   bank: "vbank",
//   user: "user-123",
//   consent_id: "pa-consent-123",
//   message: "PA consent created successfully"
// }
```

### GET `/api/pa-consents/{id}`
Получить статус PA consent.

**Request:**
```javascript
const consentId = 'pa-consent-123';
const response = await fetch(`http://localhost:8080/api/pa-consents/${consentId}?bank=vbank`);
const data = await response.json();
```

---

## 7. Products (Продукты банка)

### GET `/api/products`
Получить список доступных продуктов банка.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя
- `type` - тип продукта (DEPOSIT, LOAN, CARD) - опционально

**Request:**
```javascript
// Все продукты
const response = await fetch('http://localhost:8080/api/products?bank=vbank&user=user-123');
const products = await response.json();

// Только вклады
const depositsResponse = await fetch('http://localhost:8080/api/products?bank=vbank&user=user-123&type=DEPOSIT');
const deposits = await depositsResponse.json();

// [
//   {
//     product_id: "deposit-1",
//     product_type: "DEPOSIT",
//     name: "Накопительный вклад",
//     description: "Вклад с возможностью пополнения",
//     currency: "RUB",
//     interest_rate: {
//       rate: "5.5",
//       type: "Annual"
//     },
//     min_amount: "10000.00",
//     max_amount: "5000000.00",
//     term: {
//       min: 3,
//       max: 36,
//       unit: "MONTHS"
//     }
//   }
// ]
```

---

## 8. Agreements (Договоры)

### POST `/api/agreements`
Открыть договор (вклад/кредит/карта).

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя

**Body:**
```json
{
  "product_id": "deposit-1",
  "client_id": "user-123",
  "amount": {
    "amount": "100000.00",
    "currency": "RUB"
  },
  "term": 12,
  "term_unit": "MONTHS",
  "account_id": "acc-123"
}
```

**Request:**
```javascript
const agreementRequest = {
  product_id: "deposit-1",
  client_id: "user-123",
  amount: {
    amount: "100000.00",
    currency: "RUB"
  },
  term: 12,
  term_unit: "MONTHS",
  account_id: "acc-123"
};

const response = await fetch('http://localhost:8080/api/agreements?bank=vbank&user=user-123', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(agreementRequest)
});

const agreement = await response.json();
// {
//   agreement_id: "agreement-123",
//   product_id: "deposit-1",
//   product_type: "DEPOSIT",
//   status: "Active",
//   amount: {...},
//   interest_rate: "5.5",
//   term: 12,
//   term_unit: "MONTHS",
//   start_date: "2025-11-08T00:00:00Z",
//   end_date: "2026-11-08T00:00:00Z",
//   created_at: "2025-11-08T20:00:00Z"
// }
```

### GET `/api/agreements`
Получить список договоров клиента.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя

**Request:**
```javascript
const response = await fetch('http://localhost:8080/api/agreements?bank=vbank&user=user-123');
const agreements = await response.json();

// [
//   {
//     agreement_id: "agreement-123",
//     product_type: "DEPOSIT",
//     status: "Active",
//     amount: {...},
//     interest_rate: "5.5"
//   }
// ]
```

### GET `/api/agreements/{id}`
Получить детали договора.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя

**Request:**
```javascript
const agreementId = 'agreement-123';
const response = await fetch(`http://localhost:8080/api/agreements/${agreementId}?bank=vbank&user=user-123`);
const agreement = await response.json();
```

### DELETE `/api/agreements/{id}`
Закрыть договор.

**Query Parameters:**
- `bank` - код банка
- `user` - ID пользователя

**Request:**
```javascript
const agreementId = 'agreement-123';
const response = await fetch(`http://localhost:8080/api/agreements/${agreementId}?bank=vbank&user=user-123`, {
  method: 'DELETE'
});
const result = await response.json();
// {
//   agreement_id: "agreement-123",
//   status: "closed"
// }
```

---

## Примеры типовых сценариев

### Сценарий 1: Получить все счета и транзакции пользователя

```javascript
const userId = 'user-123';
const bank = 'vbank';

// 1. Подключить банк (создать consent)
const connectResponse = await fetch(`http://localhost:8080/api/banks/${bank}/connect?user=${userId}`, {
  method: 'POST'
});
const { consent_id } = await connectResponse.json();

// 2. Получить счета
const accountsResponse = await fetch(`http://localhost:8080/api/accounts?user=${userId}&bank=${bank}`);
const accounts = await accountsResponse.json();

// 3. Для каждого счета получить транзакции
for (const account of accounts) {
  const txResponse = await fetch(
    `http://localhost:8080/api/accounts/${account.id}/transactions?bank=${bank}&user=${userId}&from=2025-01-01&to=2025-12-31`
  );
  const transactions = await txResponse.json();
  console.log(`Счёт ${account.id}: ${transactions.length} транзакций`);
}
```

### Сценарий 2: Выполнить платёж

```javascript
const userId = 'user-123';
const bank = 'vbank';

// 1. Создать платёж (consent создается автоматически)
const paymentRequest = {
  debtor_account: {
    scheme_name: "RU.CBR.PAN",
    identification: "40817810099910004312",
    name: "Иван Иванов"
  },
  creditor_account: {
    scheme_name: "RU.CBR.PAN",
    identification: "40817810099910005423",
    name: "Магазин ООО"
  },
  amount: {
    amount: "1500.00",
    currency: "RUB"
  },
  reference: "Order #12345"
};

const paymentResponse = await fetch(`http://localhost:8080/api/payments?bank=${bank}&user=${userId}`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(paymentRequest)
});
const payment = await paymentResponse.json();

// 2. Проверить статус платежа
const statusResponse = await fetch(`http://localhost:8080/api/payments/${payment.payment_id}?bank=${bank}&user=${userId}`);
const status = await statusResponse.json();
console.log('Статус платежа:', status.status);
```

### Сценарий 3: Открыть вклад

```javascript
const userId = 'user-123';
const bank = 'vbank';

// 1. Получить список продуктов (вклады)
const productsResponse = await fetch(`http://localhost:8080/api/products?bank=${bank}&user=${userId}&type=DEPOSIT`);
const products = await productsResponse.json();

// 2. Выбрать продукт и открыть договор
const selectedProduct = products[0];
const agreementRequest = {
  product_id: selectedProduct.product_id,
  client_id: userId,
  amount: {
    amount: "100000.00",
    currency: "RUB"
  },
  term: 12,
  term_unit: "MONTHS"
};

const agreementResponse = await fetch(`http://localhost:8080/api/agreements?bank=${bank}&user=${userId}`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(agreementRequest)
});
const agreement = await agreementResponse.json();
console.log('Договор открыт:', agreement.agreement_id);
```

---

## Обработка ошибок

Все endpoints возвращают ошибки в едином формате:

```json
{
  "error": "Bad Request",
  "message": "Missing 'bank' query parameter",
  "request_id": "abc-123-def"
}
```

**HTTP Status Codes:**
- `200 OK` - успешный запрос
- `201 Created` - ресурс создан
- `400 Bad Request` - неверные параметры
- `401 Unauthorized` - отсутствует токен/consent
- `403 Forbidden` - недостаточно прав
- `404 Not Found` - ресурс не найден
- `422 Unprocessable Entity` - ошибка валидации
- `500 Internal Server Error` - ошибка сервера

**Пример обработки:**
```javascript
const response = await fetch('http://localhost:8080/api/accounts?user=user-123&bank=vbank');

if (!response.ok) {
  const error = await response.json();
  console.error(`Error ${response.status}:`, error.message);
  throw new Error(error.message);
}

const accounts = await response.json();
```

---

## Настройка CORS

Backend настроен на `http://localhost:5173` (Vite dev server).

Если фронтенд запущен на другом порту, обновите `.env`:
```
CORS_ORIGIN=http://localhost:3000
```

---

## TypeScript Types

```typescript
interface Account {
  id: string;
  bank: string;
  type: string;
  currency: string;
  balance: number;
  owner?: string;
}

interface Transaction {
  id: string;
  date: string;
  amount: number;
  currency: string;
  merchant?: string;
  category?: string;
  description?: string;
  bank: string;
}

interface Payment {
  payment_id: string;
  status: string;
  debtor_account: AccountInfo;
  creditor_account: AccountInfo;
  amount: Amount;
  reference?: string;
  created_at: string;
}

interface Product {
  product_id: string;
  product_type: 'DEPOSIT' | 'LOAN' | 'CARD';
  name: string;
  description?: string;
  currency?: string;
  interest_rate?: {
    rate: string;
    type: string;
  };
  min_amount?: string;
  max_amount?: string;
  term?: {
    min: number;
    max: number;
    unit: 'DAYS' | 'MONTHS' | 'YEARS';
  };
}

interface Agreement {
  agreement_id: string;
  product_id?: string;
  product_type?: string;
  status: string;
  amount?: Amount;
  interest_rate?: string;
  term?: number;
  term_unit?: string;
  start_date?: string;
  end_date?: string;
}

interface Amount {
  amount: string;
  currency: string;
}

interface AccountInfo {
  scheme_name: string;
  identification: string;
  name?: string;
}
```

---

## Запуск Backend

```bash
cd finhelper/backend
./finhelper-backend
```

Backend запустится на `http://localhost:8080` и выведет список всех доступных endpoints.

---

## Важные замечания

1. **Все суммы передаются как строки** (например, `"1500.00"`, не `1500`)
2. **Даты в ISO-8601** формате
3. **client_id = user** для межбанковских операций
4. **Автоматическое создание consent'ов** - backend автоматически создает и кэширует consent'ы для упрощения работы
5. **Все запросы логируются** с Request ID для отладки
