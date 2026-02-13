#!/bin/bash

BASE_URL="http://127.0.0.1:8000/api/v2"
POLICYHOLDER_ID=1

echo "=== v2 Health Check ==="
curl -s -X POST "$BASE_URL/health" | jq
echo -e "\n"

echo "=== Create Record ==="
curl -s -X POST "$BASE_URL/records/$POLICYHOLDER_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "Rainbow Corp", "country": "US", "email": "contact@rainbow.com"}' | jq
echo -e "\n"

echo "=== Update Record ==="
curl -s -X POST "$BASE_URL/records/$POLICYHOLDER_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "Rainbow Corp", "country": "US", "email": "contact@rainbow.com", "status": "active"}' | jq
echo -e "\n"

echo "=== Get Record ==="
curl -s -X GET "$BASE_URL/records/$POLICYHOLDER_ID" | jq
echo -e "\n"

echo "=== Get Non-existent Record (Expect Error) ==="
curl -s -X GET "$BASE_URL/records/9999" | jq
echo -e "\n"

echo "=== Test Completed ==="
