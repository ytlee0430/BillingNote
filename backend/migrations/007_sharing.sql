-- Pairing codes for user sharing
CREATE TABLE IF NOT EXISTS user_pairing_codes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(9) NOT NULL UNIQUE, -- AB12-CD34 format
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pairing_codes_user_id ON user_pairing_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_pairing_codes_code ON user_pairing_codes(code);

-- Shared access between users
CREATE TABLE IF NOT EXISTS shared_access (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    viewer_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(owner_id, viewer_id)
);

CREATE INDEX IF NOT EXISTS idx_shared_access_owner_id ON shared_access(owner_id);
CREATE INDEX IF NOT EXISTS idx_shared_access_viewer_id ON shared_access(viewer_id);
