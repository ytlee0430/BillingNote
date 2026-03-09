-- Add tags column to transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';

-- Create GIN index for array containment queries
CREATE INDEX IF NOT EXISTS idx_transactions_tags ON transactions USING GIN (tags);

-- Enable pg_trgm extension for trigram similarity search (must be before index that uses it)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create GIN index for full-text search on description
CREATE INDEX IF NOT EXISTS idx_transactions_description_trgm ON transactions USING GIN (description gin_trgm_ops);
