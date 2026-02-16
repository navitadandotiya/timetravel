#!/bin/sh
set -e

BASE_URL="http://timetravel-timetravel-1:8000/api/v2"
POLICYHOLDER_ID=3
USER_ID="10"

echo "=== v2 Health Check ==="
curl -sS -X POST "$BASE_URL/health" \
  -H "X-User-ID: $USER_ID" 
echo "\n"

echo "=== Feature Flags Refresh ==="
curl -sS -X POST "$BASE_URL/admin/refresh-flags" \
  -H "X-User-ID: $USER_ID" 
echo "\n"

echo "=== Create Record ==="
curl -sS -X POST "$BASE_URL/records/$POLICYHOLDER_ID" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d '{"name": "Rainbow Corp", "country": "US", "email": "contact@rainbow.com"}' 
echo "\n"

echo "=== Update Record ==="
curl -sS -X POST "$BASE_URL/records/$POLICYHOLDER_ID" \
  -H "Content-Type: application/json" \
  -H "X-User-ID: $USER_ID" \
  -d '{"name": "Rainbow Corp", "country": "US", "email": "contact@rainbow.com", "status": "active"}' 
echo "\n"

echo "=== Get Record ==="
curl -sS -X GET "$BASE_URL/records/$POLICYHOLDER_ID" \
  -H "X-User-ID: $USER_ID" 
echo "\n"

echo "=== Get Non-existent Record (Expect Error) ==="
curl -sS -X GET "$BASE_URL/records/9999" \
  -H "X-User-ID: $USER_ID" 
echo "\n"

echo "=== Get Specific Version ==="
curl -sS "$BASE_URL/records/$POLICYHOLDER_ID/versions/1" \
  -H "X-User-ID: $USER_ID" 
echo "\n"

echo "=== List All Versions ==="
curl -sS "$BASE_URL/records/$POLICYHOLDER_ID/versions" \
  -H "X-User-ID: $USER_ID" 
echo "\n"

echo "=== Test Completed ==="


curl -X GET "http://localhost:8000/metrics"  | grep -v '^#' | grep 'http_requests_total'

echo "=== metrics for SLA ==="