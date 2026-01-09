-- Create user_pdf_passwords table
CREATE TABLE IF NOT EXISTS user_pdf_passwords (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_encrypted TEXT NOT NULL,
    priority INTEGER NOT NULL CHECK (priority >= 1 AND priority <= 4),
    label VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, priority)
);

-- Create index for faster lookups
CREATE INDEX idx_pdf_passwords_user_id ON user_pdf_passwords(user_id);

-- Add source column to transactions if not exists (for tracking PDF imports)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'transactions' AND column_name = 'source'
    ) THEN
        ALTER TABLE transactions ADD COLUMN source VARCHAR(50) DEFAULT 'manual';
    END IF;
END
$$;
