PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;

--------------------------------------------------
-- MIGRATION VERSION TRACKING
--------------------------------------------------
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------
-- POLICYHOLDER (NEW)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS policyholders (
    policyholder_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT,
    country_code TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------
-- EXTEND POLICYHOLDER RECORDS (V1 → V2)
--------------------------------------------------

--ALTER TABLE policyholder_records ADD COLUMN policyholder_id INTEGER;
--ALTER TABLE policyholder_records ADD COLUMN version INTEGER DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_records_version
ON policyholder_records(record_id, version);

--------------------------------------------------
-- AUDIT HISTORY (VERSION SNAPSHOTS)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_history (
    audit_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    version INTEGER NOT NULL,
    data TEXT NOT NULL,
    changed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    event_type TEXT NOT NULL,
    FOREIGN KEY(record_id) REFERENCES policyholder_records(record_id)
);

CREATE INDEX IF NOT EXISTS idx_audit_record
ON audit_history(record_id);

CREATE INDEX IF NOT EXISTS idx_audit_time
ON audit_history(record_id, changed_at);

--------------------------------------------------
-- EVENT LOG (TRACEABILITY)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS event_logs (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER,
    action TEXT NOT NULL,
    details TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------
-- OBSERVABILITY METRICS
--------------------------------------------------
CREATE TABLE IF NOT EXISTS observability_metrics (
    metric_id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_type TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    value REAL,
    region TEXT,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------
-- FEATURE FLAGS (ROLL OUT CONTROL)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS feature_flags (
    flag_key TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL,
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    rollout_percentage INTEGER DEFAULT 100
);


--------------------------------------------------
-- API VERSION CONFIGURATION
--------------------------------------------------
CREATE TABLE IF NOT EXISTS api_version_config (
    version TEXT PRIMARY KEY,
    is_active BOOLEAN NOT NULL,
    rollout_percentage INTEGER DEFAULT 100,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------
-- RECORD MIGRATION APPLIED
--------------------------------------------------
INSERT OR IGNORE INTO schema_migrations(version)
VALUES ('001_time_travel_v2');

COMMIT;
PRAGMA foreign_keys = ON;


INSERT INTO feature_flags(flag_key, enabled, description, updated_at)
VALUES
('enable_v2_api', 1, 'Enable version 2 API', CURRENT_TIMESTAMP),
('enable_audit_logging', 1, 'Audit history logging', CURRENT_TIMESTAMP),
('enable_metrics', 1, 'Observability metrics', CURRENT_TIMESTAMP)
ON CONFLICT(flag_key)
DO UPDATE SET
    enabled = excluded.enabled,
    description = excluded.description,
    updated_at = CURRENT_TIMESTAMP;

--verification SELECT enabled FROM feature_flags WHERE flag_key = 'enable_v2_api';