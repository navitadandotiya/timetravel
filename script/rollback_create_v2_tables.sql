PRAGMA foreign_keys = OFF;

BEGIN TRANSACTION;

--------------------------------------------------
-- 1️⃣ Restore v1 policyholder_records structure
--------------------------------------------------

-- backup current v2 table
ALTER TABLE policyholder_records RENAME TO policyholder_records_v2_backup;

-- recreate original v1 table structure
CREATE TABLE policyholder_records (
    record_id INTEGER PRIMARY KEY,
    data TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------
-- 2️⃣ Copy latest state back to v1 table
--------------------------------------------------

INSERT INTO policyholder_records (record_id, data, created_at, updated_at)
SELECT
    record_id,
    data,
    created_at,
    updated_at
FROM policyholder_records_v2_backup;

--------------------------------------------------
-- 3️⃣ Drop v2-specific tables
--------------------------------------------------

DROP TABLE IF EXISTS audit_history;
DROP TABLE IF EXISTS event_logs;
DROP TABLE IF EXISTS observability_metrics;
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS policyholders;

--------------------------------------------------
-- 4️⃣ Remove version index (v2 only)
--------------------------------------------------

DROP INDEX IF EXISTS idx_records_version;

--------------------------------------------------
-- 5️⃣ Cleanup backup table
--------------------------------------------------

DROP TABLE IF EXISTS policyholder_records_v2_backup;

COMMIT;

PRAGMA foreign_keys = ON;
