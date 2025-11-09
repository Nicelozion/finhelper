# Скрипт для запуска Backend сервера
# Использование: .\start.ps1

Write-Host "=== FinHelper Backend Server ===" -ForegroundColor Cyan
Write-Host ""

# Проверка наличия Go
$goVersion = & go version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go не установлен!" -ForegroundColor Red
    Write-Host "Скачайте и установите Go: https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

Write-Host "Go версия: $goVersion" -ForegroundColor Green
Write-Host ""

# Проверка наличия .env файла
if (-not (Test-Path ".env")) {
    Write-Host "WARNING: Файл .env не найден!" -ForegroundColor Yellow
    if (Test-Path ".env.example") {
        Write-Host "Копирую .env.example -> .env..." -ForegroundColor Cyan
        Copy-Item ".env.example" ".env"
        Write-Host "ВАЖНО: Отредактируйте файл .env и укажите правильные значения!" -ForegroundColor Red
        Write-Host ""
        Start-Sleep -Seconds 2
    } else {
        Write-Host "ERROR: Файл .env.example также не найден!" -ForegroundColor Red
        exit 1
    }
}

# Установка зависимостей
Write-Host "Проверка зависимостей..." -ForegroundColor Cyan
& go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Не удалось загрузить зависимости!" -ForegroundColor Red
    exit 1
}

Write-Host "Зависимости установлены" -ForegroundColor Green
Write-Host ""

# Запуск сервера
Write-Host "Запуск сервера..." -ForegroundColor Cyan
Write-Host "Для остановки нажмите Ctrl+C" -ForegroundColor Yellow
Write-Host ""
Write-Host "===========================================" -ForegroundColor Gray
Write-Host ""

& go run .
