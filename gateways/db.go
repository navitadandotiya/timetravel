package gateways

import (
    "database/sql"
    "log"

    _ "github.com/mattn/go-sqlite3"
)

// ConnectDB opens a persistent SQLite database
func ConnectDB(dbPath string) *sql.DB {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatalf("Failed to open DB: %v", err)
    }

    if err := db.Ping(); err != nil {
        log.Fatalf("Failed to ping DB: %v", err)
    }

    return db
}
