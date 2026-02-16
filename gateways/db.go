package gateways

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// ConnectDB opens a SQLite connection and creates tables.
// If sqlFilePath is empty, skips reading a file (for in-memory testing)
func ConnectDB(path, sqlFilePath string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open SQLite DB: %v", err)
	}

	// If a SQL file path is provided, read and execute it
	if sqlFilePath != "" {
		content, err := os.ReadFile(sqlFilePath)
		if err != nil {
			log.Fatalf("failed to read %s: %v", sqlFilePath, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("failed to create tables: %v", err)
		}
	}

	return db
}

// RunMigrations executes all SQL files in the directory, sorted by name
func RunMigrations(db *sql.DB, migrationsPath string) error {
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		version := file.Name()

		_, err := db.Exec(`
							CREATE TABLE IF NOT EXISTS schema_migrations (
    							version TEXT PRIMARY KEY,
    							applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
							);
						`)
		if err != nil {
   			 return fmt.Errorf("failed to create schema_migrations: %w", err)
		}

		// Skip already applied migrations
		var exists int
		err = db.QueryRow("SELECT 1 FROM schema_migrations WHERE version = ?", version).Scan(&exists)
		if err == nil {
			continue // already applied
		} else if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check migration %s: %w", version, err)
		}

		path := filepath.Join(migrationsPath, version)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", version, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}
	}

	return nil
}
