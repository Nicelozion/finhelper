# Тестирование работы с разными типами пользователей
# team053-1 до team053-10

$BASE_URL = "http://localhost:8080"
$BANK = "vbank"

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Тестирование пользователей team053-X" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Функция для красивого вывода JSON
function Show-Json {
    param($Response, $Title)
    Write-Host "`n--- $Title ---" -ForegroundColor Yellow
    $Response | ConvertTo-Json -Depth 5 | Write-Host
}

# Типы пользователей согласно документации
$userTypes = @{
    "1" = "employee (обычный)"
    "2" = "vip (депозит 500k-2MP)"
    "3" = "entrepreneur (кредит 500k-2MP)"
    "4" = "business/ООО"
    "5" = "student"
    "6" = "pensioner (депозит)"
    "7" = "employee (дебетовая карта)"
    "8" = "employee (кредитная карта)"
    "9" = "entrepreneur (много продуктов)"
    "10" = "vip premium (ВСЕ продукты)"
}

# Тестируем каждого пользователя
for ($i = 1; $i -le 3; $i++) {
    $userId = "team053-$i"
    $userType = $userTypes["$i"]

    Write-Host "`n================================================" -ForegroundColor Green
    Write-Host "Пользователь: $userId - $userType" -ForegroundColor Green
    Write-Host "================================================" -ForegroundColor Green

    try {
        # 1. Создаем consent
        Write-Host "`n[1] Создание consent для $userId..." -ForegroundColor Cyan
        $consentResponse = Invoke-RestMethod -Method POST -Uri "$BASE_URL/api/consents?bank=$BANK&user=$userId"
        Show-Json $consentResponse "Ответ создания consent"

        $consentId = $consentResponse.consent_id
        Write-Host "`nConsent ID: $consentId" -ForegroundColor Green

        # 2. Проверяем статус consent
        Write-Host "`n[2] Проверка статуса consent..." -ForegroundColor Cyan
        Start-Sleep -Seconds 1
        $consentStatus = Invoke-RestMethod -Uri "$BASE_URL/api/consents/${consentId}?bank=$BANK"
        Show-Json $consentStatus "Статус consent"

        # 3. Получаем счета
        Write-Host "`n[3] Получение счетов для $userId..." -ForegroundColor Cyan
        Start-Sleep -Seconds 1
        $accounts = Invoke-RestMethod -Uri "$BASE_URL/api/accounts?user=$userId&bank=$BANK"
        Show-Json $accounts "Счета пользователя"

        if ($accounts.accounts -and $accounts.accounts.Count -gt 0) {
            $accountId = $accounts.accounts[0].id
            Write-Host "`nПервый счет ID: $accountId" -ForegroundColor Green

            # 4. Получаем баланс первого счета
            Write-Host "`n[4] Получение баланса счета $accountId..." -ForegroundColor Cyan
            Start-Sleep -Seconds 1
            $balance = Invoke-RestMethod -Uri "$BASE_URL/api/accounts/$accountId/balances?bank=$BANK&user=$userId"
            Show-Json $balance "Баланс счета"

            # 5. Получаем транзакции
            Write-Host "`n[5] Получение транзакций счета $accountId..." -ForegroundColor Cyan
            Start-Sleep -Seconds 1
            $transactions = Invoke-RestMethod -Uri "$BASE_URL/api/accounts/$accountId/transactions?bank=$BANK&user=$userId"
            Show-Json $transactions "Транзакции"
        } else {
            Write-Host "`nНет счетов для этого пользователя" -ForegroundColor Yellow
        }

        # 6. Получаем продукты (вклады/кредиты/карты)
        Write-Host "`n[6] Получение доступных продуктов..." -ForegroundColor Cyan
        Start-Sleep -Seconds 1

        foreach ($productType in @("DEPOSIT", "LOAN", "CARD")) {
            try {
                $products = Invoke-RestMethod -Uri "$BASE_URL/api/products?bank=$BANK&user=$userId&type=$productType"
                if ($products.products -and $products.products.Count -gt 0) {
                    Write-Host "`n  Продукты типа $productType`: $($products.products.Count) шт." -ForegroundColor Magenta
                    $products.products | ForEach-Object {
                        Write-Host "    - $($_.name) (ID: $($_.product_id))" -ForegroundColor Gray
                    }
                }
            } catch {
                Write-Host "    Ошибка получения $productType`: $_" -ForegroundColor Red
            }
        }

        Write-Host "`n✅ Тестирование $userId завершено успешно" -ForegroundColor Green

    } catch {
        Write-Host "`n❌ Ошибка при тестировании $userId`: $_" -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red
    }

    Write-Host "`n" -NoNewline
    Start-Sleep -Seconds 2
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Тестирование завершено!" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan
