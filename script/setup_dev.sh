#!/bin/bash

# ===============================
# Developer Environment Setup Script
# ===============================

set -e

echo "=== Starting Developer Environment Setup ==="

# -------------------------------
# 1. Check for Go installation
# -------------------------------
if ! command -v go &> /dev/null
then
    echo "Go not found. Please install Go: https://golang.org/dl/"
else
    echo "Go found: $(go version)"
fi

# -------------------------------
# 2. Check for SQLite installation
# -------------------------------
if ! command -v sqlite3 &> /dev/null
then
    echo "SQLite not found."
    # macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "Installing SQLite via Homebrew..."
        brew install sqlite
    # Linux (Debian/Ubuntu)
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "Installing SQLite via apt..."
        sudo apt update
        sudo apt install -y sqlite3
    else
        echo "Please install SQLite manually: https://www.sqlite.org/download.html"
    fi
else
    echo "SQLite found: $(sqlite3 --version)"
fi

# -------------------------------
# 3. Check for jq installation (used in test scripts)
# -------------------------------
if ! command -v jq &> /dev/null
then
    echo "jq not found."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "Installing jq via Homebrew..."
        brew install jq
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "Installing jq via apt..."
        sudo apt update
        sudo apt install -y jq
    else
        echo "Please install jq manually: https://stedolan.github.io/jq/"
    fi
else
    echo "jq found: $(jq --version)"
fi

# -------------------------------
# 4. Check for Git
# -------------------------------
if ! command -v git &> /dev/null
then
    echo "Git not found. Please install Git: https://git-scm.com/downloads"
else
    echo "Git found: $(git --version)"
fi

# -------------------------------
# 5. Initialize Go modules
# -------------------------------
echo "Initializing Go modules..."
go mod tidy

# -------------------------------
# 6. Install SQLite Go driver
# -------------------------------
echo "Installing SQLite Go driver..."
go get github.com/mattn/go-sqlite3

echo "=== Developer Environment Setup Complete ==="
echo "Next steps:"
echo "1. Run SQL script to create tables: ./script/create_tables.sql"
echo "2. Run the project: go run main.go"
echo "3. Run test scripts: ./test/test_v1.sh"
