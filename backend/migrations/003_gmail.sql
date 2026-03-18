-- Gmail OAuth tokens (encrypted)
CREATE TABLE IF NOT EXISTS gmail_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    access_token_encrypted TEXT NOT NULL,
    refresh_token_encrypted TEXT NOT NULL,
    token_expiry TIMESTAMP,
    scopes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Gmail scan rules (per-user configuration)
CREATE TABLE IF NOT EXISTS gmail_scan_rules (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    sender_keywords TEXT[] DEFAULT '{"credit","信用卡","帳單","statement"}',
    subject_keywords TEXT[] DEFAULT '{"帳單","電子帳單","statement"}',
    require_attachment BOOLEAN DEFAULT TRUE,
    last_scan_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Gmail scan history log
CREATE TABLE IF NOT EXISTS gmail_scan_history (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scan_at TIMESTAMP DEFAULT NOW(),
    emails_found INT DEFAULT 0,
    pdfs_downloaded INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'completed',
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_gmail_scan_history_user ON gmail_scan_history(user_id, scan_at DESC);
