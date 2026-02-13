-- Policyholder table
CREATE TABLE IF NOT EXISTS policyholders (
    policyholder_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    country_code TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Policyholder Records table
CREATE TABLE IF NOT EXISTS policyholder_records (
    record_id INTEGER PRIMARY KEY AUTOINCREMENT,
    policyholder_id INTEGER NOT NULL,
    data TEXT NOT NULL, -- JSON stored as TEXT
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (policyholder_id) REFERENCES policyholders(policyholder_id)
);

-- Audit history table
CREATE TABLE IF NOT EXISTS audit_history (
    audit_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    data TEXT NOT NULL,
    event_type TEXT NOT NULL, -- create/update/delete
    changed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (record_id) REFERENCES policyholder_records(record_id)
);

-- Event log table
CREATE TABLE IF NOT EXISTS event_log (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    action TEXT NOT NULL,
    details TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (record_id) REFERENCES policyholder_records(record_id)
);
