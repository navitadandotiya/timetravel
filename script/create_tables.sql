-- Policyholder table
CREATE TABLE IF NOT EXISTS policyholders (
    policyholder_id INTEGER PRIMARY KEY,
    name TEXT,
    email TEXT,
    country_code TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Policy table
CREATE TABLE IF NOT EXISTS policies (
    policy_id INTEGER PRIMARY KEY,
    policyholder_id INTEGER NOT NULL,
    status TEXT,
    effective_date TIMESTAMP,
    expiration_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(policyholder_id) REFERENCES policyholders(policyholder_id)
);

-- EventLog table
CREATE TABLE IF NOT EXISTS event_logs (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    event_type TEXT,
    payload TEXT,
    source_service TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
