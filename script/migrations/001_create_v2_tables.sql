PRAGMA foreign_keys = OFF;

--------------------------------------------------
-- FEATURE FLAGS
--------------------------------------------------
CREATE TABLE IF NOT EXISTS feature_flags (
    flag_key TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL,
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    rollout_percentage INTEGER DEFAULT 100
);

INSERT INTO feature_flags(flag_key, enabled, description, updated_at, rollout_percentage)
VALUES
('enable_v2_api', 1, 'Enable version 2 API', CURRENT_TIMESTAMP, 100),
('enable_audit_logging', 1, 'Audit history logging', CURRENT_TIMESTAMP, 100),
('enable_metrics', 1, 'Observability metrics', CURRENT_TIMESTAMP, 100)
ON CONFLICT(flag_key) DO UPDATE SET
    enabled = excluded.enabled,
    description = excluded.description,
    updated_at = CURRENT_TIMESTAMP,
    rollout_percentage = excluded.rollout_percentage;

--------------------------------------------------
-- POLICYHOLDER TABLE
--------------------------------------------------
CREATE TABLE IF NOT EXISTS policyholders (
    policyholder_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    country_code TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------
-- POLICYHOLDER RECORDS (V2)
--------------------------------------------------
-- migrate data if old table exists
DROP TABLE IF EXISTS policyholder_records_old;
ALTER TABLE policyholder_records RENAME TO policyholder_records_old;

CREATE TABLE IF NOT EXISTS policyholder_records (
    record_id INTEGER PRIMARY KEY AUTOINCREMENT,
    policyholder_id INTEGER NOT NULL,
    data TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(policyholder_id) REFERENCES policyholders(policyholder_id) ON DELETE CASCADE
);

-- copy data from old table if exists
INSERT INTO policyholder_records (record_id, policyholder_id, data, version, created_at, updated_at)
SELECT record_id, policyholder_id, data, 1, created_at, updated_at
FROM policyholder_records_old;

DROP TABLE IF EXISTS policyholder_records_old;

CREATE INDEX IF NOT EXISTS idx_records_version
ON policyholder_records(record_id, version);

--------------------------------------------------
-- AUDIT HISTORY TABLE
--------------------------------------------------
DROP TABLE IF EXISTS audit_history_old;
ALTER TABLE audit_history RENAME TO audit_history_old;

CREATE TABLE IF NOT EXISTS audit_history (
    audit_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    version INTEGER NOT NULL,
    data TEXT NOT NULL,
    event_type TEXT NOT NULL,
    changed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(record_id) REFERENCES policyholder_records(record_id) ON DELETE CASCADE
);

INSERT INTO audit_history (audit_id, record_id, version, data, event_type, changed_at)
SELECT audit_id, record_id, audit_id, data, event_type, changed_at
FROM audit_history_old;

DROP TABLE IF EXISTS audit_history_old;

CREATE INDEX IF NOT EXISTS idx_audit_record
ON audit_history(record_id);

CREATE INDEX IF NOT EXISTS idx_audit_time
ON audit_history(record_id, changed_at);

--------------------------------------------------
-- EVENT LOG
--------------------------------------------------
CREATE TABLE IF NOT EXISTS event_logs (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER,
    action TEXT NOT NULL,
    details TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(record_id) REFERENCES policyholder_records(record_id) ON DELETE CASCADE
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
-- SCHEMA MIGRATION TRACKING
--------------------------------------------------
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY);

INSERT INTO schema_migrations(version)
VALUES('001_create_users.sql')
ON CONFLICT(version) DO NOTHING;


PRAGMA foreign_keys = ON;
