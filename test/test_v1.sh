#!/bin/bash

# Base URL
BASE_URL="http://127.0.0.1:8000/api/v1"

# Colors for pretty output
GREEN="\033[0;32m"
RED="\033[0;31m"
NC="\033[0m"

# --- Health Check ---
echo -e "${GREEN}Testing Health Check...${NC}"
curl -s -X POST "$BASE_URL/health" | jq .
echo -e "\n"

# --- Create / Update Record ---
RECORD_ID=1
echo -e "${GREEN}Testing POST /records/$RECORD_ID...${NC}"
curl -s -X POST "$BASE_URL/records/$RECORD_ID" \
    -H "Content-Type: application/json" \
    -d '{"hello": "world"}' | jq .
echo -e "\n"

# --- Update Record (add field) ---
echo -e "${GREEN}Testing POST update /records/$RECORD_ID...${NC}"
curl -s -X POST "$BASE_URL/records/$RECORD_ID" \
    -H "Content-Type: application/json" \
    -d '{"status": "ok"}' | jq .
echo -e "\n"

# --- Delete a field from Record ---
echo -e "${GREEN}Testing POST delete field /records/$RECORD_ID...${NC}"
curl -s -X POST "$BASE_URL/records/$RECORD_ID" \
    -H "Content-Type: application/json" \
    -d '{"hello": ""}' | jq .
echo -e "\n"

# --- Get Record ---
echo -e "${GREEN}Testing GET /records/$RECORD_ID...${NC}"
curl -s -X GET "$BASE_URL/records/$RECORD_ID" | jq .
echo -e "\n"

# --- Get Non-existent Record ---
NON_EXISTENT_ID=999
echo -e "${RED}Testing GET non-existent /records/$NON_EXISTENT_ID...${NC}"
curl -s -X GET "$BASE_URL/records/$NON_EXISTENT_ID" | jq .
echo -e "\n"

echo -e "${GREEN}All v1 tests completed!${NC}"
