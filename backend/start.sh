#!/bin/bash

# Скрипт для запуска Backend сервера
# Использование: ./start.sh

echo "=== FinHelper Backend Server ==="
echo ""

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;37m'
NC='\033[0m' # No Color

# Проверка наличия Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}ERROR: Go не установлен!${NC}"
    echo -e "${YELLOW}Скачайте и установите Go: https://go.dev/dl/${NC}"
    exit 1
fi

GO_VERSION=$(go version)
echo -e "${GREEN}Go версия: $GO_VERSION${NC}"
echo ""

# Проверка наличия .env файла
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}WARNING: Файл .env не найден!${NC}"
    if [ -f ".env.example" ]; then
        echo -e "${CYAN}Копирую .env.example -> .env...${NC}"
        cp ".env.example" ".env"
        echo -e "${RED}ВАЖНО: Отредактируйте файл .env и укажите правильные значения!${NC}"
        echo ""
        sleep 2
    else
        echo -e "${RED}ERROR: Файл .env.example также не найден!${NC}"
        exit 1
    fi
fi

# Установка зависимостей
echo -e "${CYAN}Проверка зависимостей...${NC}"
go mod download
if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: Не удалось загрузить зависимости!${NC}"
    exit 1
fi

echo -e "${GREEN}Зависимости установлены${NC}"
echo ""

# Запуск сервера
echo -e "${CYAN}Запуск сервера...${NC}"
echo -e "${YELLOW}Для остановки нажмите Ctrl+C${NC}"
echo ""
echo -e "${GRAY}===========================================${NC}"
echo ""

go run .
