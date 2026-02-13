package gateways

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectDB opens a SQLite connection and ensures tables exist
func ConnectDB(path string) *sql.DB {
	// Create DB dir if missing
	if _, err := os.Stat("db"); os.IsNotExist(err) {
		os.Mkdir("db", os.ModePerm)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open SQLite DB: %v", err)
	}

	// Run table creation statements
	tableSQL, err := os.ReadFile("script/create_v2_tables.sql")
	if err != nil {
		log.Fatalf("failed to read create_v2_tables.sql: %v", err)
	}

	_, err = db.Exec(string(tableSQL))
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	return db
}
