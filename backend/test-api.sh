#!/bin/bash

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
USER="test-$(date +%s)"

PASSED=0
FAILED=0

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}🧪 FinHelper API Test Suite${NC}"
echo -e "${BLUE}=========================================${NC}"
echo -e "Base URL: ${BASE_URL}"
echo -e "Test User: ${USER}"
echo ""

function test_endpoint() {
    local METHOD=$1
    local ENDPOINT=$2
    local DESCRIPTION=$3
    
    echo -n "Testing: $DESCRIPTION... "
    
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X $METHOD "$BASE_URL$ENDPOINT")
    
    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "201" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $HTTP_CODE)"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC} (HTTP $HTTP_CODE)"
        ((FAILED++))
        return 1
    fi
}

# Test 1: Health Check
echo -e "${YELLOW}Test Suite 1: Health Check${NC}"
test_endpoint "GET" "/healthz" "Health endpoint"
echo ""

# Test 2: Consent Management
echo -e "${YELLOW}Test Suite 2: Consent Management${NC}"
test_endpoint "POST" "/api/consents?bank=vbank&user=$USER" "Create consent for VBank"
test_endpoint "POST" "/api/consents?bank=abank&user=$USER" "Create consent for ABank"
test_endpoint "POST" "/api/consents?bank=sbank&user=$USER" "Create consent for SBank"
echo ""

# Test 3: Account Operations
echo -e "${YELLOW}Test Suite 3: Account Operations${NC}"
test_endpoint "GET" "/api/accounts?user=$USER" "Get accounts from all banks"
test_endpoint "GET" "/api/accounts?user=$USER&bank=vbank" "Get accounts from VBank"
echo ""

# Test 4: Transaction Operations
echo -e "${YELLOW}Test Suite 4: Transaction Operations${NC}"
FROM=$(date -d '30 days ago' -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -v-30d -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "2025-10-01T00:00:00Z")
test_endpoint "GET" "/api/transactions?user=$USER&from=$FROM" "Get transactions from all banks"
test_endpoint "GET" "/api/transactions?user=$USER&bank=vbank&from=$FROM" "Get transactions from VBank"
echo ""

# Test 5: Legacy Endpoints
echo -e "${YELLOW}Test Suite 5: Legacy Endpoints${NC}"
LEGACY_USER="legacy-$(date +%s)"
test_endpoint "POST" "/api/banks/vbank/connect?user=$LEGACY_USER" "Legacy bank connect"
echo ""

# Summary
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}📊 Test Results${NC}"
echo -e "${BLUE}=========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "${BLUE}=========================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed! Backend is working correctly! 🎉${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed. Check the logs above.${NC}"
    exit 1
fi