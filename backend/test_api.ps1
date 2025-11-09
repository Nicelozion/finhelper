# PowerShell скрипт для быстрого тестирования API
# Использование: .\test_api.ps1

$BaseUrl = "http://localhost:8080"
$Bank = "vbank"
$User = "testuser"

Write-Host "=== Тестирование Backend API ===" -ForegroundColor Cyan
Write-Host ""

# Функция для вывода результата
function Print-Result {
    param (
        [bool]$Success,
        [string]$Message
    )
    if ($Success) {
        Write-Host "✓ $Message" -ForegroundColor Green
    } else {
        Write-Host "✗ $Message" -ForegroundColor Red
    }
}

# 1. Health Check
Write-Host "1. Health Check" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/health" -Method Get -ErrorAction Stop
    Print-Result -Success $true -Message "Health check passed"
    Write-Host "   Response: $($response | ConvertTo-Json -Compress)" -ForegroundColor Gray
} catch {
    Print-Result -Success $false -Message "Health check failed"
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 2. Создание консента
Write-Host "2. Создание консента" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/consents?bank=$Bank&user=$User" -Method Post -ErrorAction Stop
    Print-Result -Success $true -Message "Consent created"
    $ConsentId = $response.consentId
    Write-Host "   Consent ID: $ConsentId" -ForegroundColor Gray
    Write-Host "   Status: $($response.status)" -ForegroundColor Gray
} catch {
    Print-Result -Success $false -Message "Consent creation failed"
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 3. Получение статуса консента
if ($ConsentId) {
    Write-Host "3. Получение статуса консента" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/consents/$ConsentId?bank=$Bank" -Method Get -ErrorAction Stop
        Print-Result -Success $true -Message "Consent status retrieved"
        Write-Host "   Status: $($response.status)" -ForegroundColor Gray
        Write-Host "   Response: $($response | ConvertTo-Json -Compress)" -ForegroundColor Gray
    } catch {
        Print-Result -Success $false -Message "Consent status failed"
        Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""
}

# 4. Получение счетов
Write-Host "4. Получение счетов" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/accounts?user=$User&bank=$Bank" -Method Get -ErrorAction Stop
    Print-Result -Success $true -Message "Accounts retrieved"
    $AccountId = $null
    if ($response.accounts -and $response.accounts.Count -gt 0) {
        $AccountId = $response.accounts[0].accountId
        Write-Host "   First Account ID: $AccountId" -ForegroundColor Gray
        Write-Host "   Total accounts: $($response.accounts.Count)" -ForegroundColor Gray
    }
    Write-Host "   Response: $($response | ConvertTo-Json -Compress -Depth 3)" -ForegroundColor Gray
} catch {
    Print-Result -Success $false -Message "Accounts retrieval failed"
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 5. Получение баланса (если есть accountId)
if ($AccountId) {
    Write-Host "5. Получение баланса счета" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/accounts/$AccountId/balances?bank=$Bank&user=$User" -Method Get -ErrorAction Stop
        Print-Result -Success $true -Message "Balance retrieved"
        if ($response.balances -and $response.balances.Count -gt 0) {
            $balance = $response.balances[0]
            Write-Host "   Amount: $($balance.amount.amount) $($balance.amount.currency)" -ForegroundColor Gray
        }
        Write-Host "   Response: $($response | ConvertTo-Json -Compress -Depth 3)" -ForegroundColor Gray
    } catch {
        Print-Result -Success $false -Message "Balance retrieval failed"
        Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""
}

# 6. Получение транзакций (если есть accountId)
if ($AccountId) {
    Write-Host "6. Получение транзакций счета" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/accounts/$AccountId/transactions?bank=$Bank&user=$User" -Method Get -ErrorAction Stop
        Print-Result -Success $true -Message "Transactions retrieved"
        if ($response.transactions) {
            Write-Host "   Total transactions: $($response.transactions.Count)" -ForegroundColor Gray
        }
        Write-Host "   Response: $($response | ConvertTo-Json -Compress -Depth 3)" -ForegroundColor Gray
    } catch {
        Print-Result -Success $false -Message "Transactions retrieval failed"
        Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host ""
}

# 7. Получение продуктов
Write-Host "7. Получение банковских продуктов (вклады)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/products?bank=$Bank&user=$User&type=DEPOSIT" -Method Get -ErrorAction Stop
    Print-Result -Success $true -Message "Products retrieved"
    if ($response.products -and $response.products.Count -gt 0) {
        $product = $response.products[0]
        Write-Host "   First Product ID: $($product.productId)" -ForegroundColor Gray
        Write-Host "   Product Name: $($product.productName)" -ForegroundColor Gray
        Write-Host "   Total products: $($response.products.Count)" -ForegroundColor Gray
    }
    Write-Host "   Response: $($response | ConvertTo-Json -Compress -Depth 3)" -ForegroundColor Gray
} catch {
    Print-Result -Success $false -Message "Products retrieval failed"
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 8. Получение транзакций за период (все банки)
Write-Host "8. Получение всех транзакций за период" -ForegroundColor Yellow
$fromDate = (Get-Date).AddDays(-30).ToString("yyyy-MM-dd")
$toDate = (Get-Date).ToString("yyyy-MM-dd")
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/transactions?user=$User&bank=$Bank&from=$fromDate&to=$toDate" -Method Get -ErrorAction Stop
    Print-Result -Success $true -Message "Transactions for period retrieved"
    if ($response.transactions) {
        Write-Host "   Total transactions: $($response.transactions.Count)" -ForegroundColor Gray
        Write-Host "   Period: $fromDate to $toDate" -ForegroundColor Gray
    }
} catch {
    Print-Result -Success $false -Message "Transactions for period failed"
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "=== Тестирование завершено ===" -ForegroundColor Green
Write-Host ""
Write-Host "Tip: Используйте -Verbose для более подробного вывода" -ForegroundColor Cyan
