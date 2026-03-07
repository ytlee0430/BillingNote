-- Add invoice carrier to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS invoice_carrier VARCHAR(10);

-- Invoices table
CREATE TABLE IF NOT EXISTS invoices (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invoice_number VARCHAR(10) NOT NULL,
    invoice_date TIMESTAMP NOT NULL,
    seller_name VARCHAR(255),
    seller_ban VARCHAR(8),
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50),
    items JSONB,
    is_duplicated BOOLEAN DEFAULT FALSE,
    duplicated_transaction_id INT REFERENCES transactions(id),
    confidence_score DECIMAL(3, 2),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, invoice_number)
);

CREATE INDEX IF NOT EXISTS idx_invoices_user_date ON invoices(user_id, invoice_date);
