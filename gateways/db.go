package gateways

import (
	"database/sql"
	"log"
	"os"
	"fmt"
	"io/ioutil"
	"path/filepath"
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

func RunMigrations(db *sql.DB, migrationsDir string) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}

	for _, file := range files {
		version := filepath.Base(file)

		var exists string
		err := db.QueryRow(
			"SELECT version FROM schema_migrations WHERE version = ?",
			version,
		).Scan(&exists)

		if err == nil {
			log.Printf("migration already applied: %s", version)
			continue
		}

		sqlBytes, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		log.Printf("applying migration: %s", version)

		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			return fmt.Errorf("failed migration %s: %w", version, err)
		}
	}

	return nil
}
