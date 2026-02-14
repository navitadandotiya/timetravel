package service

import (
	"database/sql"
	"sync"
	"github.com/rainbowmga/timetravel/observability"
)

type FeatureFlagService struct {
	db   *sql.DB
	mu   sync.RWMutex
	cache map[string]bool
}


// NewFeatureFlagService initializes and loads flags into memory
func NewFeatureFlagService(dbPath string) (*FeatureFlagService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enforce foreign keys
	_, _ = db.Exec("PRAGMA foreign_keys = ON;")

	s := &FeatureFlagService{
		db:    db,
		cache: make(map[string]bool),
	}
	if err := s.reload(); err != nil {
		return nil, err
	}
	return s, nil
}

// reload loads flags from DB into memory
func (s *FeatureFlagService) reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rows, err := s.db.Query("SELECT flag_key, enabled FROM feature_flags")
	if err != nil {
		return err
	}
	defer rows.Close()

	s.cache = make(map[string]bool)
	for rows.Next() {
		var key string
		var enabled bool
		if err := rows.Scan(&key, &enabled); err != nil {
			return err
		}
		s.cache[key] = enabled
	}
	return nil
}

// IsEnabled checks if a flag is enabled
func (s *FeatureFlagService) IsEnabled(flag string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	observability.DefaultLogger.Info("IsEnabled ",flag, s.cache[flag])
	return s.cache[flag]
}

// Refresh forces reload from DB at runtime
func (s *FeatureFlagService) Refresh() error {
	observability.DefaultLogger.Info("Refresh ...... ")
	observability.DefaultLogger.Info("Refresh ", s.cache["enable_v2_api"])
	return s.reload()
}
