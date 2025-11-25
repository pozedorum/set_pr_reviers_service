#!/bin/bash

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

test_endpoint() {
    local name=$1
    local method=$2
    local url=$3
    local data=$4
    
    echo -n "Testing $name... "
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "%{http_code}" "$BASE_URL$url")
    else
        response=$(curl -s -w "%{http_code}" -X $method -H "Content-Type: application/json" -d "$data" "$BASE_URL$url")
    fi
    
    http_code=${response: -3}
    body=${response%???}
    
    if [ $http_code -ge 200 ] && [ $http_code -lt 300 ]; then
        echo -e "${GREEN}✓ ($http_code)${NC}"
    else
        echo -e "${RED}✗ ($http_code)${NC}"
        echo "Response: $body"
    fi
}

echo "🚀 Starting API tests..."

# Test all endpoints
test_endpoint "Health Check" "GET" "/health" ""

test_endpoint "Create Team" "POST" "/team/add" '{
    "team_name": "test-team",
    "members": [
        {"user_id": "test1", "username": "Test User 1", "is_active": true},
        {"user_id": "test2", "username": "Test User 2", "is_active": true}
    ]
}'

test_endpoint "Get Team" "GET" "/team/get?team_name=test-team" ""

test_endpoint "Create PR" "POST" "/pullRequest/create" '{
    "pull_request_id": "test-pr-001",
    "pull_request_name": "Test PR",
    "author_id": "test1"
}'

test_endpoint "Get User Reviews" "GET" "/users/getReview?user_id=test2" ""

test_endpoint "Merge PR" "POST" "/pullRequest/merge" '{
    "pull_request_id": "test-pr-001"
}'

echo "🎉 All tests completed!"