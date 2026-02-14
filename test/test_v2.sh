#!/bin/bash

BASE_URL="http://127.0.0.1:8000/api/v2"
POLICYHOLDER_ID=1

echo "=== v2 Health Check ==="
curl -s -X POST "$BASE_URL/health" | jq
echo -e "\n"



echo "=== Feature Flags Refresh ==="
curl -s -X POST http://localhost:8000/api/v2/admin/refresh-flags | jq
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

echo "=== Get specific version ==="
curl -s "$BASE_URL/records/$POLICYHOLDER_ID/versions/1" | jq
echo -e "\n"

echo "=== List all versions ==="
curl -s "$BASE_URL/records/$POLICYHOLDER_ID/versions" | jq
echo -e "\n"


echo "=== Test Completed ==="

