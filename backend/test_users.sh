#!/bin/bash

# Тестирование работы с разными типами пользователей
# team053-1 до team053-10

BASE_URL="http://localhost:8080"
BANK="vbank"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
echo "========================================"
echo "Тестирование пользователей team053-X"
echo "========================================"
echo -e "${NC}"

# Типы пользователей согласно документации
declare -A userTypes=(
    [1]="employee (обычный)"
    [2]="vip (депозит 500k-2MP)"
    [3]="entrepreneur (кредит 500k-2MP)"
    [4]="business/ООО"
    [5]="student"
    [6]="pensioner (депозит)"
    [7]="employee (дебетовая карта)"
    [8]="employee (кредитная карта)"
    [9]="entrepreneur (много продуктов)"
    [10]="vip premium (ВСЕ продукты)"
)

# Тестируем первых 3 пользователей (можно изменить на 10)
for i in {1..3}; do
    userId="team053-$i"
    userType="${userTypes[$i]}"

    echo -e "${GREEN}"
    echo "================================================"
    echo "Пользователь: $userId - $userType"
    echo "================================================"
    echo -e "${NC}"

    # 1. Создаем consent
    echo -e "${CYAN}[1] Создание consent для $userId...${NC}"
    consentResponse=$(curl -s -X POST "$BASE_URL/api/consents?bank=$BANK&user=$userId")
    echo "$consentResponse" | jq '.'

    consentId=$(echo "$consentResponse" | jq -r '.consent_id')
    echo -e "${GREEN}Consent ID: $consentId${NC}"

    # 2. Проверяем статус consent
    echo -e "${CYAN}[2] Проверка статуса consent...${NC}"
    sleep 1
    curl -s "$BASE_URL/api/consents/$consentId?bank=$BANK" | jq '.'

    # 3. Получаем счета
    echo -e "${CYAN}[3] Получение счетов для $userId...${NC}"
    sleep 1
    accountsResponse=$(curl -s "$BASE_URL/api/accounts?user=$userId&bank=$BANK")
    echo "$accountsResponse" | jq '.'

    accountId=$(echo "$accountsResponse" | jq -r '.accounts[0].id // empty')

    if [ -n "$accountId" ]; then
        echo -e "${GREEN}Первый счет ID: $accountId${NC}"

        # 4. Получаем баланс первого счета
        echo -e "${CYAN}[4] Получение баланса счета $accountId...${NC}"
        sleep 1
        curl -s "$BASE_URL/api/accounts/$accountId/balances?bank=$BANK&user=$userId" | jq '.'

        # 5. Получаем транзакции
        echo -e "${CYAN}[5] Получение транзакций счета $accountId...${NC}"
        sleep 1
        curl -s "$BASE_URL/api/accounts/$accountId/transactions?bank=$BANK&user=$userId" | jq '.'
    else
        echo -e "${YELLOW}Нет счетов для этого пользователя${NC}"
    fi

    # 6. Получаем продукты (вклады/кредиты/карты)
    echo -e "${CYAN}[6] Получение доступных продуктов...${NC}"
    sleep 1

    for productType in DEPOSIT LOAN CARD; do
        echo -e "${MAGENTA}  Продукты типа $productType:${NC}"
        productsResponse=$(curl -s "$BASE_URL/api/products?bank=$BANK&user=$userId&type=$productType")
        productsCount=$(echo "$productsResponse" | jq '.products | length')

        if [ "$productsCount" -gt 0 ]; then
            echo -e "    Найдено: $productsCount шт."
            echo "$productsResponse" | jq -r '.products[] | "    - \(.name) (ID: \(.product_id))"'
        fi
    done

    echo -e "${GREEN}✅ Тестирование $userId завершено успешно${NC}"
    echo ""
    sleep 2
done

echo -e "${CYAN}"
echo "========================================"
echo "Тестирование завершено!"
echo "========================================"
echo -e "${NC}"
