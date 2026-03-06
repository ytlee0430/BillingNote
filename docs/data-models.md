# Data Models

**Generated:** 2026-03-06
**Database:** PostgreSQL 15+
**ORM:** GORM 1.25.5

---

## Entity Relationship Diagram

```
users (1) ──────< transactions (N)
  |                    |
  |                    v
  |              categories (1)
  |
  └──────< user_pdf_passwords (N)
```

---

## Tables

### users

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-increment ID |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email (login identifier) |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt hashed password |
| name | VARCHAR(100) | - | Display name |
| created_at | TIMESTAMP | DEFAULT NOW() | Creation timestamp |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update timestamp |

**GORM Model:** `internal/models/user.go`
- `BeforeCreate` hook: auto-hashes password via bcrypt
- `CheckPassword(pwd)`: compares bcrypt hash
- `SetPassword(pwd)`: hashes and sets password

---

### categories

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-increment ID |
| name | VARCHAR(50) | UNIQUE, NOT NULL | Category name |
| type | VARCHAR(20) | NOT NULL, CHECK (income/expense) | Category type |
| icon | VARCHAR(50) | - | Emoji icon |
| color | VARCHAR(20) | - | Hex color code |
| created_at | TIMESTAMP | DEFAULT NOW() | Creation timestamp |

**Default categories (14):**
- Expense: Dining, Transport, Shopping, Entertainment, Medical, Education, Housing, Telecom, Insurance, Other Expense
- Income: Salary, Investment, Bonus, Other Income

---

### transactions

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-increment ID |
| user_id | INTEGER | NOT NULL, FK -> users(id) ON DELETE CASCADE | Owner |
| category_id | INTEGER | FK -> categories(id) ON DELETE SET NULL | Category reference |
| amount | DECIMAL(15,2) | NOT NULL, CHECK >= 0 | Transaction amount |
| type | VARCHAR(20) | NOT NULL, CHECK (income/expense) | Transaction type |
| description | TEXT | - | Description/memo |
| transaction_date | DATE | NOT NULL | Date of transaction |
| source | VARCHAR(50) | DEFAULT 'manual' | Source: manual/pdf/gmail/invoice |
| created_at | TIMESTAMP | DEFAULT NOW() | Creation timestamp |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update timestamp |

**Indexes:**
- `idx_transactions_user_id` on (user_id)
- `idx_transactions_category_id` on (category_id)
- `idx_transactions_date` on (transaction_date)
- `idx_transactions_type` on (type)
- `idx_transactions_user_date` on (user_id, transaction_date) - composite

---

### user_pdf_passwords

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PRIMARY KEY | Auto-increment ID |
| user_id | INTEGER | NOT NULL, FK -> users(id) ON DELETE CASCADE | Owner |
| password_encrypted | TEXT | NOT NULL | AES-256-GCM encrypted password |
| priority | INTEGER | NOT NULL, CHECK 1-4 | Try order (1 = first) |
| label | VARCHAR(100) | - | User-defined label |
| created_at | TIMESTAMP | DEFAULT NOW() | Creation timestamp |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update timestamp |

**Constraints:**
- UNIQUE (user_id, priority) - one password per priority slot per user

**Index:**
- `idx_pdf_passwords_user_id` on (user_id)

---

## Migrations

| File | Description |
|------|-------------|
| `001_init.sql` | Creates users, categories, transactions tables with indexes and default categories |
| `002_pdf_passwords.sql` | Creates user_pdf_passwords table, adds source column to transactions |

---

## Planned Tables (from spec.md)

These tables are defined in the spec but not yet implemented:

| Table | Phase | Purpose |
|-------|-------|---------|
| `invoices` | Phase 3 | Cloud invoice records with deduplication |
| `gmail_tokens` | Phase 2B | Gmail OAuth tokens (encrypted) |
| `shared_access` | Phase 5 | Multi-user read permission grants |
| `user_pairing_codes` | Phase 5 | Pairing code for user linking |
