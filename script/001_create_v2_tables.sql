PRAGMA foreign_keys=off;

--------------------------------------------------
-- 1️⃣ Rename old table
--------------------------------------------------
DROP TABLE IF EXISTS policyholder_records_old;
ALTER TABLE policyholder_records RENAME TO policyholder_records_old;


--------------------------------------------------
-- 5️⃣ Drop old table (optional, after verifying copy)
--------------------------------------------------
DROP TABLE IF EXISTS policyholder_records;

---------------------------------------------------
-- POLICYHOLDER RECORD (LATEST STATE)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS policyholder_records (
    record_id INTEGER PRIMARY KEY,
    policyholder_id INTEGER,            -- add FK reference if needed
    data TEXT NOT NULL,                 -- JSON blob
    version INTEGER NOT NULL DEFAULT 1, -- current version
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


--------------------------------------------------
-- 4️⃣ Copy data from old table
--------------------------------------------------
-- Copy identity info
INSERT INTO policyholder_records (record_id, policyholder_id, created_at)
SELECT record_id, policyholder_id, created_at
FROM policyholder_records_old;


--------------------------------------------------
-- 5️⃣ Drop old table (optional, after verifying copy)
--------------------------------------------------
DROP TABLE policyholder_records_old;


PRAGMA foreign_keys = ON;

--------------------------------------------------
-- POLICYHOLDER
--------------------------------------------------
CREATE TABLE IF NOT EXISTS policyholders (
    policyholder_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT,
    country_code TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);



CREATE INDEX IF NOT EXISTS idx_records_version
ON policyholder_records(record_id, version);

--------------------------------------------------
-- AUDIT HISTORY (IMMUTABLE SNAPSHOTS)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_history (
    audit_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    version INTEGER NOT NULL,
    data TEXT NOT NULL,                 -- snapshot JSON
    changed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    event_type TEXT NOT NULL,           -- create/update/delete
    FOREIGN KEY(record_id) REFERENCES policyholder_records(record_id)
);

CREATE INDEX IF NOT EXISTS idx_audit_record
ON audit_history(record_id);

CREATE INDEX IF NOT EXISTS idx_audit_time
ON audit_history(record_id, changed_at);

--------------------------------------------------
-- EVENT LOG (TRACEABILITY & DEBUGGING)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS event_logs (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER,
    action TEXT NOT NULL,               -- create/update/delete/read
    details TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_record
ON event_logs(record_id);

--------------------------------------------------
-- OBSERVABILITY METRICS (PLATFORM + BUSINESS)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS observability_metrics (
    metric_id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_type TEXT NOT NULL,   -- platform | business
    metric_name TEXT NOT NULL,
    value REAL,
    region TEXT,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_metric_type
ON observability_metrics(metric_type, metric_name);

--------------------------------------------------
-- FEATURE FLAGS (GLOBAL ROLLOUT CONTROL)
--------------------------------------------------
CREATE TABLE IF NOT EXISTS feature_flags (
    flag_key TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL,
    description TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
