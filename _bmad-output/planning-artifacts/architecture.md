---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - docs/project-overview.md
  - docs/architecture-backend.md
  - docs/architecture-frontend.md
  - docs/api-contracts.md
  - docs/data-models.md
  - docs/integration-architecture.md
  - spec.md
workflowType: 'architecture'
project_name: 'Billing-Note'
user_name: 'Bruce'
date: '2026-03-06'
---

# Architecture Decision Document

_Billing Note - Personal Accounting System_

---

## 1. Project Context

### 1.1 System Overview

Billing Note is a **brownfield** personal accounting web application with existing Phase 1 (MVP) and Phase 2A (PDF import) implementation. Remaining phases (2B-5) need to be built on top of the existing architecture.

### 1.2 Existing Architecture Summary

- **Backend:** Go 1.23 + Gin + GORM + PostgreSQL (layered: handler -> service -> repository)
- **Frontend:** React 18 + TypeScript + Vite + Zustand + TanStack Query + Tailwind CSS
- **Deployment:** Docker Compose (3 services: backend, frontend/nginx, PostgreSQL)
- **Auth:** JWT with bcrypt password hashing
- **Integration:** REST/JSON over HTTPS, Axios with JWT interceptor

### 1.3 Key Constraint

All new features MUST be additive to the existing codebase. No breaking changes to existing API contracts or database schema.

---

## 2. Architecture Style Decision

**Decision: Maintain existing Layered Monolith architecture**

**Rationale:**
- Single-user personal application - microservices would be over-engineering
- Existing codebase follows clean layered pattern that works well
- All new features (Gmail, invoices, sharing) fit naturally into the existing layer structure
- Docker Compose deployment sufficient for personal use

**Architecture layers (maintained):**
```
[Middleware] -> [Handler] -> [Service] -> [Repository] -> [Database]
                                |
                          [External APIs]
                          (Gmail, MOF Invoice)
```

---

## 3. Technology Decisions

### 3.1 Retained Technologies (No Changes)

| Component | Technology | Decision |
|-----------|-----------|----------|
| Backend Framework | Gin 1.9.1 | Keep - performant and well-suited |
| ORM | GORM 1.25.5 | Keep - handles all DB needs |
| Database | PostgreSQL 15+ | Keep - JSONB for invoice items |
| Frontend Framework | React 18 + TS | Keep - mature ecosystem |
| Build Tool | Vite 5 | Keep - fast HMR |
| State Management | Zustand | Keep - minimal and effective |
| Server State | TanStack Query | Keep - caching and refetching |
| Styling | Tailwind CSS 3 | Keep - utility-first works well |
| Charts | Recharts | Keep - expand for Phase 4 |

### 3.2 New Technology Additions

| Component | Technology | Phase | Decision Rationale |
|-----------|-----------|-------|-------------------|
| Gmail API | google.golang.org/api/gmail/v1 | 2B | Official Google Go client |
| OAuth2 | golang.org/x/oauth2 | 2B | Standard Go OAuth2 library |
| Cron Jobs | github.com/robfig/cron/v3 | 2B | Lightweight Go cron scheduler |
| String Similarity | github.com/agnivade/levenshtein | 3 | Efficient Levenshtein distance for dedup |
| Excel Export | github.com/xuri/excelize/v2 | 4 | Go Excel file generation |
| CSV Export | encoding/csv (stdlib) | 4 | Standard library CSV writer |

---

## 4. Data Architecture Decisions

### 4.1 New Tables (Additive)

#### Phase 2B - Gmail Integration

```sql
CREATE TABLE gmail_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    access_token_encrypted TEXT NOT NULL,
    refresh_token_encrypted TEXT NOT NULL,
    token_expiry TIMESTAMP,
    scopes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE gmail_scan_rules (
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

CREATE TABLE gmail_scan_history (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scan_at TIMESTAMP DEFAULT NOW(),
    emails_found INT DEFAULT 0,
    pdfs_downloaded INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'completed',
    error_message TEXT
);
```

#### Phase 3 - Invoices

```sql
ALTER TABLE users ADD COLUMN invoice_carrier VARCHAR(10);

CREATE TABLE invoices (
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

CREATE INDEX idx_invoices_user_date ON invoices(user_id, invoice_date);
```

#### Phase 5 - Sharing

```sql
CREATE TABLE shared_access (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(20) DEFAULT 'read',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(owner_id, shared_with)
);

CREATE TABLE user_pairing_codes (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(10) UNIQUE NOT NULL,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 4.2 Migration Strategy

- Continue using plain SQL migration files in `backend/migrations/`
- Naming convention: `{NNN}_{description}.sql`
- All migrations MUST be idempotent (`IF NOT EXISTS`)
- One migration file per phase feature set

---

## 5. API Design Decisions

### 5.1 Routing Convention (Maintained)

```
Public:     /api/auth/*
Protected:  /api/*  (requires JWT via AuthMiddleware)
```

### 5.2 New Endpoints by Phase

#### Phase 2B - Gmail
```
GET    /api/gmail/auth         -> GmailHandler.GetAuthURL
POST   /api/gmail/callback     -> GmailHandler.HandleCallback
POST   /api/gmail/scan         -> GmailHandler.TriggerScan
GET    /api/gmail/status       -> GmailHandler.GetStatus
PUT    /api/gmail/settings     -> GmailHandler.UpdateSettings
DELETE /api/gmail/disconnect   -> GmailHandler.Disconnect
```

#### Phase 3 - Invoices
```
POST   /api/invoice/sync               -> InvoiceHandler.Sync
GET    /api/invoice/list                -> InvoiceHandler.List
POST   /api/invoice/confirm-duplicate   -> InvoiceHandler.ConfirmDuplicate
DELETE /api/invoice/:id                 -> InvoiceHandler.Delete
PUT    /api/invoice/settings            -> InvoiceHandler.UpdateSettings
```

#### Phase 4 - Advanced
```
GET    /api/export/csv          -> ExportHandler.ExportCSV
GET    /api/export/excel        -> ExportHandler.ExportExcel
GET    /api/search/transactions -> SearchHandler.Search
POST   /api/budget              -> BudgetHandler.Create
GET    /api/budget              -> BudgetHandler.List
PUT    /api/budget/:id          -> BudgetHandler.Update
```

#### Phase 5 - Sharing
```
GET    /api/shared/my-code           -> SharedHandler.GetMyCode
POST   /api/shared/regenerate-code   -> SharedHandler.RegenerateCode
POST   /api/shared/pair              -> SharedHandler.Pair
GET    /api/shared/connections       -> SharedHandler.ListConnections
DELETE /api/shared/connections/:uid  -> SharedHandler.RevokeAccess
```

All existing transaction/stats endpoints gain `?view_as=<user_id>` support in Phase 5.

### 5.3 Error Handling (Maintained)

Continue using existing `AppError` pattern with consistent JSON response format.

---

## 6. Security Architecture Decisions

### 6.1 Encryption Standards

| Data | Method | Key |
|------|--------|-----|
| User passwords | bcrypt (default cost) | N/A (one-way) |
| PDF passwords | AES-256-GCM | ENCRYPTION_KEY env |
| Gmail tokens | AES-256-GCM | ENCRYPTION_KEY env |
| JWT | HS256 | JWT_SECRET env |

### 6.2 Phase 5 Security Rules

**CRITICAL: Backend-enforced permission checks**

1. New middleware: `CheckViewPermission` on all data-access endpoints
2. When `view_as` parameter present and != current user:
   - Verify `shared_access` record exists with `status = 'active'`
   - Set `read_only = true` in context
   - Reject ALL write operations (create/update/delete)
3. Frontend read-only mode is UI-only - backend MUST enforce independently
4. Permission revocation takes effect immediately (no caching)

### 6.3 OAuth Token Security

- Store tokens encrypted (AES-256-GCM) in database
- Refresh token rotation on each use
- Auto-revoke on disconnect
- Never log token values

---

## 7. Integration Architecture Decisions

### 7.1 Gmail API Integration (Phase 2B)

```
[Backend Cron/Manual] -> [Gmail Service] -> [Gmail API]
                              |
                        [Download PDFs]
                              |
                        [Upload Service] (reuse existing)
                              |
                        [PDF Parser] (reuse existing)
```

**Decision:** Server-side OAuth flow. Backend holds tokens and makes API calls. Frontend only triggers scan and displays results.

### 7.2 MOF Invoice API Integration (Phase 3)

```
[Backend] -> [Invoice Service] -> [MOF API]
                   |
             [Dedup Service] -> [Transaction Repository]
                   |
             [Invoice Repository]
```

**Decision:** Backend-only integration. Frontend displays results and handles duplicate confirmation UI.

### 7.3 Deduplication Algorithm (Phase 3)

```
For each invoice:
  1. Find transactions within +/- 3 days of invoice date
  2. Filter: |transaction.amount - invoice.amount| <= 1.0
  3. For remaining: Calculate Levenshtein similarity of merchant names
  4. If similarity > 0.8: Mark as potential duplicate with confidence score
  5. Return matches sorted by confidence (highest first)
```

**Decision:** Conservative matching with user confirmation. No auto-merge.

---

## 8. Frontend Architecture Decisions

### 8.1 New Pages/Routes

| Route | Page | Phase |
|-------|------|-------|
| `/invoices` | Invoice list with duplicate handling | 3 |
| `/budget` | Budget management | 4 |
| existing `/charts` | Add new chart types | 4 |
| existing `/settings` | Add Gmail, Invoice, Sharing sections | 2B, 3, 5 |

### 8.2 State Management (Maintained)

- **Zustand:** Auth state only (keep minimal)
- **TanStack Query:** All server data (transactions, invoices, gmail status, etc.)
- **No new Zustand stores needed** - TanStack Query handles all new data

### 8.3 Component Strategy

- Continue co-locating components by feature domain
- New component folders: `components/gmail/`, `components/invoice/`, `components/sharing/`
- Reuse existing `common/` components (Button, Input, Modal)

---

## 9. Testing Architecture Decisions

### 9.1 Testing Strategy (Maintained)

| Layer | Tool | Coverage Target |
|-------|------|----------------|
| Backend Unit | testify + go-sqlmock | >= 80% |
| Backend Integration | httptest + test DB | Key flows |
| Frontend Unit | Vitest + RTL | >= 75% |
| Frontend E2E | Playwright | Critical paths |

### 9.2 New Test Requirements per Phase

**Phase 2B:** Mock Gmail API, test OAuth flow, test scan filtering
**Phase 3:** Test dedup algorithm exhaustively (edge cases critical), mock MOF API
**Phase 4:** Test export formats, test search queries
**Phase 5:** Security tests (unauthorized access, permission escalation, revocation)

### 9.3 Critical Test Scenarios (Phase 5)

- Unauthorized access prevention
- Write rejection in read-only mode
- Immediate effect of permission revocation
- SQL injection protection on view_as parameter
- Self-authorization prevention

---

## 10. Deployment Architecture (Maintained)

```yaml
# Docker Compose - no changes to structure
services:
  frontend:  # nginx + React build (port 80)
  backend:   # Go binary (port 8080)
  db:        # PostgreSQL 15 (port 5432)
```

**New environment variables needed:**
- `GOOGLE_CLIENT_ID` - OAuth client ID (Phase 2B)
- `GOOGLE_CLIENT_SECRET` - OAuth client secret (Phase 2B)
- `GOOGLE_REDIRECT_URI` - OAuth callback URL (Phase 2B)
- `EINVOICE_APP_ID` - MOF API key (Phase 3)
- `EINVOICE_API_URL` - MOF API base URL (Phase 3)

---

## 11. File/Folder Structure Decisions

### Backend New Files

```
backend/internal/
├── handlers/
│   ├── gmail.go              # Phase 2B
│   ├── invoice.go            # Phase 3
│   ├── export.go             # Phase 4
│   ├── search.go             # Phase 4
│   ├── budget.go             # Phase 4
│   └── shared.go             # Phase 5
├── services/
│   ├── gmail_service.go      # Phase 2B
│   ├── invoice_service.go    # Phase 3
│   ├── deduplication.go      # Phase 3
│   ├── export_service.go     # Phase 4
│   ├── search_service.go     # Phase 4
│   ├── budget_service.go     # Phase 4
│   ├── pairing_service.go    # Phase 5
│   └── shared_access_service.go # Phase 5
├── models/
│   ├── gmail_token.go        # Phase 2B
│   ├── invoice.go            # Phase 3
│   ├── budget.go             # Phase 4
│   ├── shared_access.go      # Phase 5
│   └── pairing_code.go       # Phase 5
├── repository/
│   ├── gmail_repo.go         # Phase 2B
│   ├── invoice_repo.go       # Phase 3
│   ├── budget_repo.go        # Phase 4
│   └── shared_access_repo.go # Phase 5
├── middleware/
│   └── permission.go         # Phase 5
└── migrations/
    ├── 003_gmail.sql          # Phase 2B
    ├── 004_invoices.sql       # Phase 3
    ├── 005_budgets_tags.sql   # Phase 4
    └── 006_sharing.sql        # Phase 5
```

### Frontend New Files

```
frontend/src/
├── components/
│   ├── gmail/
│   │   ├── GmailConnect.tsx     # Phase 2B
│   │   └── ScanHistory.tsx      # Phase 2B
│   ├── invoice/
│   │   ├── InvoiceList.tsx      # Phase 3
│   │   └── DuplicateHandler.tsx # Phase 3
│   ├── sharing/
│   │   ├── PairingCode.tsx      # Phase 5
│   │   └── ConnectionList.tsx   # Phase 5
│   └── charts/
│       └── LineChart.tsx        # Phase 4
├── pages/
│   ├── Invoices.tsx             # Phase 3
│   └── Budget.tsx               # Phase 4
├── api/
│   ├── gmail.ts                 # Phase 2B
│   ├── invoices.ts              # Phase 3
│   ├── budget.ts                # Phase 4
│   ├── export.ts                # Phase 4
│   └── sharing.ts               # Phase 5
└── types/
    ├── gmail.ts                 # Phase 2B
    ├── invoice.ts               # Phase 3
    ├── budget.ts                # Phase 4
    └── sharing.ts               # Phase 5
```

---

## 12. Decision Log

| # | Decision | Rationale | Date |
|---|----------|-----------|------|
| 1 | Maintain layered monolith | Single user app, existing architecture works | 2026-03-06 |
| 2 | Server-side OAuth for Gmail | Tokens stay on backend, more secure | 2026-03-06 |
| 3 | Conservative dedup with user confirm | Avoid false positive auto-merges | 2026-03-06 |
| 4 | Backend-enforced permissions (Phase 5) | Frontend can be bypassed | 2026-03-06 |
| 5 | Plain SQL migrations, no migration tool | Simple, existing pattern works | 2026-03-06 |
| 6 | TanStack Query for all new server state | Consistent with existing pattern | 2026-03-06 |
| 7 | AES-256-GCM for all token encryption | Reuse existing crypto/aes package | 2026-03-06 |
| 8 | Reuse existing PDF pipeline for Gmail PDFs | Avoid code duplication | 2026-03-06 |
