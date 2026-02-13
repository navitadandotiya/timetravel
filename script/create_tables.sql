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

-- RiskProfile table
CREATE TABLE IF NOT EXISTS risk_profiles (
    risk_profile_id INTEGER PRIMARY KEY,
    policy_id INTEGER NOT NULL,
    liability_limit REAL,
    workforce_size INTEGER,
    risk_classification TEXT,
    premium_estimate REAL,
    version_number INTEGER NOT NULL,
    effective_from TIMESTAMP,
    effective_to TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(policy_id) REFERENCES policies(policy_id)
);

-- RiskProfileHistory table
CREATE TABLE IF NOT EXISTS risk_profile_history (
    history_id INTEGER PRIMARY KEY AUTOINCREMENT,
    risk_profile_id INTEGER NOT NULL,
    change_type TEXT,
    old_value TEXT,
    new_value TEXT,
    changed_by TEXT,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    valid_from TIMESTAMP,
    valid_to TIMESTAMP,
    FOREIGN KEY(risk_profile_id) REFERENCES risk_profiles(risk_profile_id)
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
