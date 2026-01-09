-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
    icon VARCHAR(50),
    color VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    amount DECIMAL(15, 2) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
    description TEXT,
    transaction_date DATE NOT NULL,
    source VARCHAR(50) DEFAULT 'manual',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT amount_positive CHECK (amount >= 0)
);

-- Create indexes for better query performance
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_category_id ON transactions(category_id);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_user_date ON transactions(user_id, transaction_date);

-- Insert default categories
INSERT INTO categories (name, type, icon, color) VALUES
-- Expense categories
('餐飲', 'expense', '🍔', '#FF6B6B'),
('交通', 'expense', '🚗', '#4ECDC4'),
('購物', 'expense', '🛍️', '#95E1D3'),
('娛樂', 'expense', '🎮', '#F38181'),
('醫療', 'expense', '🏥', '#AA96DA'),
('教育', 'expense', '📚', '#FCBAD3'),
('居家', 'expense', '🏠', '#FFFFD2'),
('通訊', 'expense', '📱', '#A8D8EA'),
('保險', 'expense', '🛡️', '#C7CEEA'),
('其他支出', 'expense', '📦', '#B4B4B4'),
-- Income categories
('薪資', 'income', '💰', '#48C774'),
('投資', 'income', '📈', '#3273DC'),
('獎金', 'income', '🎁', '#FFDD57'),
('其他收入', 'income', '💵', '#00D1B2')
ON CONFLICT (name) DO NOTHING;
