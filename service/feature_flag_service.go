package service

import (
	"database/sql"
	"sync"
	"hash/fnv"
	"fmt"
	"strconv"
	"github.com/rainbowmga/timetravel/observability"
)

// service/feature_flag_service.go
type FeatureFlagServiceInterface interface {
	IsEnabled(flagKey string, userID int64) bool
	Refresh() error
}

// Ensure FeatureFlagService implements the interface
var _ FeatureFlagServiceInterface = (*FeatureFlagService)(nil)

type FeatureFlagService struct {
	db   *sql.DB
	mu   sync.RWMutex
	cache map[string]FeatureFlag
}

type FeatureFlag struct {
	Key               string
	Enabled           bool
	RolloutPercentage int
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
		cache: make(map[string]FeatureFlag),
	}
	if err := s.Refresh(); err != nil {
		return nil, err
	}
	return s, nil
}

// reload loads flags from DB into memory
func (s *FeatureFlagService) Refresh() error {
	if s == nil || s.db == nil {
        return fmt.Errorf("database not initialized")
    }

	rows, err := s.db.Query(`
		SELECT flag_key, enabled, rollout_percentage
		FROM feature_flags`)
	if err != nil {
		return err
	}
	defer rows.Close()

	tmp := make(map[string]FeatureFlag)

	for rows.Next() {
		var f FeatureFlag
		if err := rows.Scan(&f.Key, &f.Enabled, &f.RolloutPercentage); err != nil {
			return err
		}
		tmp[f.Key] = f
	}

	s.mu.Lock()
	s.cache = tmp
	s.mu.Unlock()

	observability.DefaultLogger.Info("Refresh ", s.cache["enable_v2_api"])
	return nil
}

// IsEnabled checks if a flag is enabled
func (s *FeatureFlagService) IsEnabled(flagKey string, userID int64) bool {
	s.mu.RLock()
	flag, ok := s.cache[flagKey]
	s.mu.RUnlock()

	if !ok || !flag.Enabled {
		observability.FlagEvaluated(flagKey, false)
		return false
	}

	// 100% rollout
	if flag.RolloutPercentage >= 100 {
		observability.FlagEvaluated(flagKey, true)
		return true
	}

	// deterministic hash
	h := fnv.New32a()
	h.Write([]byte(strconv.FormatInt(userID, 10)))
	value := h.Sum32() % 100

	enabled := int(value) < flag.RolloutPercentage

	observability.FlagEvaluated(flagKey, enabled)
	observability.DefaultLogger.Info("IsEnabled ", flagKey, s.cache[flagKey])
	return enabled
}



