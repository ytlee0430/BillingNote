---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]
inputDocuments:
  - docs/index.md
  - docs/project-overview.md
  - docs/architecture-backend.md
  - docs/architecture-frontend.md
  - docs/api-contracts.md
  - docs/data-models.md
  - docs/integration-architecture.md
  - docs/development-guide.md
  - spec.md
workflowType: 'prd'
lastStep: 11
documentCounts:
  briefs: 0
  research: 0
  projectDocs: 8
  spec: 1
---

# Product Requirements Document - Billing Note

**Author:** Bruce
**Date:** 2026-03-06
**Version:** 1.0
**Status:** Approved

---

## 1. Executive Summary

### 1.1 Product Vision

Billing Note is a personal web-based accounting system that automates financial data import from multiple sources (credit card PDF statements, cloud invoices, Gmail attachments) to minimize manual data entry. It provides visual spending analytics and supports multi-user read-only sharing for family financial transparency.

### 1.2 Problem Statement

Managing personal finances across multiple banks and payment methods requires tedious manual entry. Credit card statements arrive as encrypted PDFs, invoices are scattered across government platforms, and consolidating this data into a single view is time-consuming and error-prone.

### 1.3 Target Users

- **Primary:** Individual user (Bruce) for personal finance tracking
- **Phase 5 Extension:** Family members with read-only access to each other's data

### 1.4 Success Metrics

| Metric | Target |
|--------|--------|
| Manual data entry reduction | > 80% via automated imports |
| PDF parsing accuracy | > 95% for supported banks |
| Invoice deduplication accuracy | > 95% |
| Backend test coverage | >= 80% |
| Frontend test coverage | >= 75% |

---

## 2. Existing System Overview (Brownfield)

### 2.1 Current Implementation Status

| Phase | Feature | Status |
|-------|---------|--------|
| Phase 1 | User auth (JWT) | Implemented |
| Phase 1 | Manual transaction CRUD | Implemented |
| Phase 1 | Transaction list (pagination, filtering) | Implemented |
| Phase 1 | Basic charts (monthly bar, category pie) | Implemented |
| Phase 1 | Category management (14 defaults) | Implemented |
| Phase 2A | PDF password management (4 slots, AES-256) | Implemented |
| Phase 2A | PDF upload & parsing pipeline | Implemented |
| Phase 2A | Bank parsers (Cathay, Taishin, Fubon) | Implemented |
| Phase 2A | Transaction preview & import from PDF | Implemented |
| Phase 2B | Gmail auto-scan | Not started |
| Phase 3 | Cloud invoice integration | Not started |
| Phase 4 | Advanced features | Not started |
| Phase 5 | Multi-user sharing | Not started |

### 2.2 Current Architecture

- **Backend:** Go 1.23 + Gin 1.9.1 + GORM 1.25.5 + PostgreSQL 15+
- **Frontend:** React 18 + TypeScript 5.3 + Vite 5 + Tailwind CSS 3
- **Database:** 4 tables (users, categories, transactions, user_pdf_passwords)
- **API:** 17 REST endpoints with JWT auth
- **Deployment:** Docker Compose (3 services)

### 2.3 Existing API Surface

See [API Contracts](../../docs/api-contracts.md) for full endpoint documentation.

Key existing endpoints:
- Auth: register, login, me
- Transactions: CRUD + monthly stats + category stats
- Categories: list, filter by type
- PDF: upload/parse, import, password management

---

## 3. Functional Requirements

### 3.1 Phase 2B - Gmail Auto-Scan (NEW)

**Priority:** High
**Dependencies:** Phase 2A (PDF parsing - done)

#### FR-2B-1: Google OAuth Integration
- User can connect Gmail account via OAuth 2.0
- System stores encrypted refresh_token
- User can disconnect Gmail at any time
- Required scopes: `gmail.readonly`, `gmail.metadata`

#### FR-2B-2: Gmail Email Scanning
- Scan Gmail for credit card statement emails
- Configurable scan rules (sender keywords, subject keywords)
- Default rules: sender contains "credit/信用卡/帳單", subject contains "帳單/statement", has attachment
- Track last scan timestamp to avoid re-scanning

#### FR-2B-3: PDF Attachment Download
- Auto-download PDF attachments from matched emails
- Save to user's upload directory
- Skip already-downloaded files (dedup by email ID)
- Filter: only `.pdf` files

#### FR-2B-4: Auto-Parse Pipeline
- Downloaded PDFs automatically enter Phase 2A parsing pipeline
- Reuse existing bank parsers and password management
- Present results in same preview UI as manual upload

#### FR-2B-5: Scan Scheduling
- Manual trigger: "Scan Now" button
- Optional: Daily auto-scan via cron job (8:00 AM)

**New API Endpoints:**
- `GET /api/gmail/auth` - Get OAuth authorization URL
- `POST /api/gmail/callback` - Handle OAuth callback
- `POST /api/gmail/scan` - Trigger manual scan
- `GET /api/gmail/status` - Get connection status
- `PUT /api/gmail/settings` - Update scan rules
- `DELETE /api/gmail/disconnect` - Remove Gmail connection

**New Database Tables:**
- `gmail_tokens` - Encrypted OAuth tokens per user
- `gmail_scan_rules` - User's scan configuration
- `gmail_scan_history` - Scan log with results

---

### 3.2 Phase 3 - Cloud Invoice Integration (NEW)

**Priority:** High
**Dependencies:** Phase 1 (transactions)

#### FR-3-1: Invoice Carrier Setup
- User inputs mobile barcode carrier (`/XXXXXXX` format)
- Validate carrier format
- Store in user profile

#### FR-3-2: Invoice Sync
- Pull invoices from Taiwan MOF e-invoice API
- API endpoint: `https://api.einvoice.nat.gov.tw/PB2CAPIVAN/invapp/InvApp`
- Support date range queries
- Store invoice details including line items (JSONB)

#### FR-3-3: Deduplication Engine
- Compare invoices against existing credit card transactions
- Matching rules:
  - Date: within +/- 3 days
  - Amount: exact match or +/- 1 TWD tolerance
  - Merchant: Levenshtein similarity > 80%
- Auto-mark duplicates with confidence score
- User can confirm or dismiss duplicate suggestions

#### FR-3-4: Invoice Management
- List invoices with duplicate indicators
- Confirm/dismiss duplicate pairs
- Delete invoices
- Sync settings (manual / daily auto)

**New API Endpoints:**
- `POST /api/invoice/sync` - Sync cloud invoices
- `GET /api/invoice/list` - List invoices (with pagination)
- `POST /api/invoice/confirm-duplicate` - Confirm duplicate pair
- `DELETE /api/invoice/:id` - Delete invoice
- `PUT /api/invoice/settings` - Update sync settings

**New Database Tables:**
- `invoices` - Invoice records with JSONB items
- Add `invoice_carrier` column to `users` table

---

### 3.3 Phase 4 - Advanced Features (NEW)

**Priority:** Medium
**Dependencies:** Phase 1-3

#### FR-4-1: Electronic Passbook Import
- Upload CSV/PDF bank statements
- Parse and import transactions
- Mark source as "bank"

#### FR-4-2: Advanced Charts
- Monthly trend line chart
- Category year-over-year grouped bar chart
- Income/expense stacked bar chart

#### FR-4-3: Budget Management
- Set monthly budget per category
- Over-budget warnings
- Budget vs actual comparison view

#### FR-4-4: Data Export
- Export transactions as CSV
- Export as Excel (.xlsx)
- Date range selection for export

#### FR-4-5: Advanced Search
- Full-text keyword search on descriptions
- Multi-condition filters (combine type + category + date + amount range)
- Save frequently used filter presets

#### FR-4-6: Tag System
- Custom tags per transaction
- Multi-tag support
- Filter by tags

#### FR-4-7: Backup/Restore
- Full data export (all user data as JSON)
- Import historical data from JSON backup

---

### 3.4 Phase 5 - Multi-User Sharing (NEW)

**Priority:** Low
**Dependencies:** Phase 1-4

#### FR-5-1: Pairing Code Mechanism
- Each user gets a unique pairing code (`AB12-CD34` format)
- Code can be regenerated
- Optional expiration

#### FR-5-2: Account Linking
- Enter another user's pairing code to request access
- Pairing creates `shared_access` record with `read` permission
- Both users must pair independently (bidirectional)

#### FR-5-3: Account Switching UI
- Dropdown in nav bar to switch between "My Account" and shared accounts
- Clear visual indicator when viewing someone else's data
- Read-only banner when viewing shared data

#### FR-5-4: Read-Only Mode Enforcement
- Frontend: Hide add/edit/delete buttons when viewing shared data
- Backend: Reject all write operations when `view_as` != current user
- API: All transaction endpoints support `?view_as=<user_id>` parameter

#### FR-5-5: Permission Management
- View connected accounts
- Revoke access at any time (immediate effect)
- View who has access to your data

**New API Endpoints:**
- `GET /api/shared/my-code` - Get pairing code
- `POST /api/shared/regenerate-code` - Generate new code
- `POST /api/shared/pair` - Link with another user
- `GET /api/shared/connections` - List connections
- `DELETE /api/shared/connections/:user_id` - Revoke access

**New Database Tables:**
- `shared_access` - Permission grants (owner_id, shared_with, permission, status)
- `user_pairing_codes` - Pairing codes per user

---

## 4. Non-Functional Requirements

### 4.1 Performance
- API response time: < 200ms for CRUD operations
- PDF parsing: < 30 seconds per file
- Page load time: < 3 seconds
- Support 1000+ transactions per user efficiently

### 4.2 Security
- Passwords: bcrypt with default cost
- JWT: HS256 with configurable expiry
- PDF passwords: AES-256-GCM encryption at rest
- Gmail tokens: AES-256-GCM encryption at rest
- All API endpoints require authentication (except auth routes)
- CORS: Configurable allowed origins
- SQL injection: Prevented via GORM parameterized queries
- XSS: React auto-escaping + no dangerouslySetInnerHTML

### 4.3 Reliability
- Database: PostgreSQL with proper indexing
- Error handling: Consistent AppError format with trace IDs
- Structured logging: logrus with request context

### 4.4 Testing
- Backend unit test coverage: >= 80%
- Frontend unit test coverage: >= 75%
- E2E tests for all critical user flows
- All tests must pass before deployment (CI gate)

### 4.5 Deployment
- Docker Compose for local and production
- Three services: frontend (nginx), backend (Go binary), database (PostgreSQL)
- Environment-based configuration via .env files

---

## 5. UI/UX Requirements

### 5.1 Existing Pages
- Login / Register
- Dashboard (stats overview + charts)
- Transactions (list + CRUD modal)
- Upload (PDF upload)
- Charts (analytics)
- Settings

### 5.2 New Pages/Sections Needed

#### Phase 2B
- Settings > Gmail Integration section
  - Connect/disconnect button
  - Scan rules configuration
  - Scan history log
  - "Scan Now" button

#### Phase 3
- Settings > Cloud Invoice section
  - Carrier code input and validation
  - Sync settings (manual/auto)
  - "Sync Now" button
- Invoice list page (or tab within Transactions)
  - Duplicate indicators
  - Confirm/dismiss duplicate UI

#### Phase 4
- Advanced Charts page (additional chart types)
- Budget Management page
  - Category budget settings
  - Budget vs actual view
- Export modal (format + date range selection)
- Advanced search filters panel
- Tag management in transaction form

#### Phase 5
- Settings > Sharing section
  - Pairing code display
  - Add connection via code
  - Manage connections list
- Nav bar: Account switcher dropdown
- Read-only banner when viewing shared data

---

## 6. Technical Constraints

### 6.1 Must Maintain
- Existing Go + Gin backend architecture (layered: handler -> service -> repository)
- Existing React + TypeScript frontend with Zustand + TanStack Query
- Existing database schema (additive migrations only)
- Existing API contract (no breaking changes to existing endpoints)
- Docker Compose deployment model

### 6.2 External Dependencies
- Taiwan MOF e-invoice API (Phase 3) - requires APP ID registration
- Google Cloud Console OAuth credentials (Phase 2B) - requires project setup
- Gmail API access (Phase 2B) - requires user authorization

### 6.3 Browser Support
- Modern browsers: Chrome, Firefox, Safari, Edge (latest 2 versions)
- Mobile responsive: Tailwind CSS responsive utilities

---

## 7. Implementation Phasing

| Phase | Scope | Dependencies |
|-------|-------|-------------|
| Phase 2B | Gmail auto-scan + PDF download | Phase 2A (done) |
| Phase 3 | Cloud invoice sync + dedup | Phase 1 (done) |
| Phase 4 | Advanced features (charts, budget, export, search, tags) | Phase 1-3 |
| Phase 5 | Multi-user sharing | Phase 1-4 |

**Recommended order:** Phase 2B -> Phase 3 -> Phase 4 -> Phase 5

Phase 2B and 3 can potentially be parallelized since they are independent features.

---

## 8. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| MOF API changes or rate limits | Phase 3 blocked | Cache responses, implement retry logic |
| Google OAuth complexity | Phase 2B delayed | Use well-documented google.golang.org/api |
| PDF format changes by banks | Parsing failures | Modular parser registry, easy to add/update parsers |
| Dedup false positives | User frustration | Conservative matching + user confirmation UI |
| Multi-user security holes | Data leakage | Backend-enforced permissions, comprehensive security tests |

---

## 9. Acceptance Criteria Summary

### Phase 2B
- [ ] User can connect Gmail via OAuth
- [ ] System scans and downloads PDF attachments
- [ ] Downloaded PDFs auto-enter parsing pipeline
- [ ] User can configure scan rules
- [ ] Gmail connection can be disconnected

### Phase 3
- [ ] User can set mobile barcode carrier
- [ ] System syncs invoices from MOF API
- [ ] Duplicates auto-detected with > 95% accuracy
- [ ] User can confirm/dismiss duplicates

### Phase 4
- [ ] Advanced charts render correctly
- [ ] Budget alerts trigger on overspend
- [ ] CSV/Excel export works for any date range
- [ ] Keyword search finds matching transactions
- [ ] Tags can be created and filtered

### Phase 5
- [ ] Users can pair via codes
- [ ] Shared data is read-only (frontend + backend enforced)
- [ ] Access can be revoked immediately
- [ ] Permission middleware covers all write endpoints

---

## 10. Glossary

| Term | Definition |
|------|-----------|
| Carrier Code | Taiwan mobile barcode for e-invoice aggregation (/XXXXXXX format) |
| MOF | Taiwan Ministry of Finance (財政部) |
| Deduplication | Process of identifying duplicate entries between invoices and credit card transactions |
| Pairing Code | Random code used to link two user accounts for shared viewing |
| BankParser | Interface for bank-specific PDF parsing strategies |
