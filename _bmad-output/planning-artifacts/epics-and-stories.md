---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/architecture.md
  - docs/project-overview.md
  - spec.md
workflowType: 'epics-and-stories'
project_name: 'Billing-Note'
date: '2026-03-06'
---

# Epics and Stories - Billing Note

---

## Epic 1: Gmail Auto-Scan Integration (Phase 2B)

**Goal:** Auto-download credit card PDF statements from Gmail and feed them into the existing PDF parsing pipeline.

### Story 1.1: Gmail OAuth Backend Service
**Priority:** High | **Points:** 5

**Description:** As a user, I want to connect my Gmail account so the system can access my emails to find credit card statements.

**Acceptance Criteria:**
- [ ] Google OAuth2 flow implemented (authorization URL generation, token exchange)
- [ ] Refresh token stored encrypted (AES-256-GCM) in `gmail_tokens` table
- [ ] Access token auto-refreshed when expired
- [ ] Migration `003_gmail.sql` creates `gmail_tokens`, `gmail_scan_rules`, `gmail_scan_history` tables
- [ ] `GET /api/gmail/auth` returns OAuth authorization URL
- [ ] `POST /api/gmail/callback` exchanges code for tokens and stores them
- [ ] `GET /api/gmail/status` returns connection status
- [ ] `DELETE /api/gmail/disconnect` removes tokens and revokes access
- [ ] Unit tests for gmail_service.go with mocked Google API
- [ ] Integration tests for OAuth flow

**Tasks:**
1. Create migration `003_gmail.sql` with all 3 tables
2. Create `models/gmail_token.go` GORM model
3. Create `services/gmail_service.go` with OAuth flow methods
4. Create `handlers/gmail.go` with auth/callback/status/disconnect endpoints
5. Register routes in `main.go`
6. Write unit tests for service layer
7. Write integration tests for API endpoints

---

### Story 1.2: Gmail Email Scanning Service
**Priority:** High | **Points:** 5

**Description:** As a user, I want the system to scan my Gmail for credit card statement emails and download PDF attachments.

**Acceptance Criteria:**
- [ ] Gmail API client queries emails matching scan rules
- [ ] Default scan rules: sender contains credit/信用卡/帳單, subject contains 帳單/statement, has attachment
- [ ] PDF attachments extracted and saved to `/uploads/{user_id}/gmail/`
- [ ] Already-downloaded emails tracked to prevent re-download
- [ ] `POST /api/gmail/scan` triggers scan and returns results
- [ ] Scan history recorded in `gmail_scan_history` table
- [ ] Downloaded PDFs auto-passed to existing `UploadService.ParsePDF()`
- [ ] Unit tests with mocked Gmail API responses
- [ ] Error handling for API rate limits

**Tasks:**
1. Implement Gmail API email search with configurable query
2. Implement attachment download and PDF filtering
3. Create scan deduplication (track processed email IDs)
4. Integrate with existing PDF parsing pipeline
5. Create scan history logging
6. Write unit tests with mocked Gmail API
7. Write integration test for full scan flow

---

### Story 1.3: Gmail Settings Frontend
**Priority:** Medium | **Points:** 3

**Description:** As a user, I want a settings UI to connect/disconnect Gmail and configure scan rules.

**Acceptance Criteria:**
- [ ] Settings page has "Gmail Integration" section
- [ ] "Connect Gmail" button triggers OAuth popup/redirect
- [ ] Connected state shows email and last scan time
- [ ] "Disconnect" button with confirmation dialog
- [ ] Scan rules editor (sender keywords, subject keywords)
- [ ] "Scan Now" button with loading state and results display
- [ ] Scan history table showing recent scans
- [ ] `frontend/src/api/gmail.ts` API client
- [ ] `frontend/src/types/gmail.ts` TypeScript types
- [ ] Component tests for Gmail settings section

**Tasks:**
1. Create `api/gmail.ts` and `types/gmail.ts`
2. Create `components/gmail/GmailConnect.tsx`
3. Create `components/gmail/ScanHistory.tsx`
4. Add Gmail section to Settings page
5. Handle OAuth callback redirect
6. Write component tests

---

## Epic 2: Cloud Invoice Integration (Phase 3)

**Goal:** Sync e-invoices from Taiwan MOF API and detect duplicates with credit card transactions.

### Story 2.1: Invoice Sync Backend Service
**Priority:** High | **Points:** 5

**Description:** As a user, I want to sync my cloud invoices from the MOF e-invoice platform using my mobile barcode carrier.

**Acceptance Criteria:**
- [ ] Migration `004_invoices.sql` creates `invoices` table and adds `invoice_carrier` to users
- [ ] `models/invoice.go` GORM model with JSONB items field
- [ ] MOF API integration: fetch invoices by carrier code and date range
- [ ] Invoice data stored with all fields (number, date, seller, amount, items)
- [ ] Duplicate invoice numbers rejected (UNIQUE constraint)
- [ ] `POST /api/invoice/sync` triggers sync for date range
- [ ] `GET /api/invoice/list` returns invoices with pagination
- [ ] `PUT /api/invoice/settings` updates carrier code
- [ ] Unit tests with mocked MOF API responses
- [ ] Handle MOF API errors gracefully

**Tasks:**
1. Create migration `004_invoices.sql`
2. Create `models/invoice.go` GORM model
3. Create `repository/invoice_repo.go`
4. Create `services/invoice_service.go` with MOF API client
5. Create `handlers/invoice.go` with all endpoints
6. Register routes in `main.go`
7. Write unit tests with mocked MOF API
8. Write integration tests

---

### Story 2.2: Deduplication Engine
**Priority:** High | **Points:** 8

**Description:** As a user, I want the system to automatically detect when an invoice matches an existing credit card transaction to avoid double-counting.

**Acceptance Criteria:**
- [ ] `services/deduplication.go` implements matching algorithm
- [ ] Match rules: date within +/-3 days, amount within +/-1 TWD, merchant similarity > 80%
- [ ] Levenshtein distance used for merchant name comparison
- [ ] Duplicate matches stored with confidence score (0.0-1.0)
- [ ] Multiple potential matches ranked by confidence
- [ ] `POST /api/invoice/confirm-duplicate` marks confirmed pair
- [ ] `DELETE /api/invoice/:id` deletes invoice
- [ ] Edge cases tested: timezone boundaries, decimal precision, multiple matches
- [ ] Dedup accuracy > 95% (validated with test data)
- [ ] Comprehensive unit tests including all edge cases from spec

**Tasks:**
1. Create `services/deduplication.go` with core algorithm
2. Implement Levenshtein similarity function
3. Implement date range and amount tolerance matching
4. Create confidence scoring logic
5. Integrate with invoice sync (auto-run after sync)
6. Create confirm/dismiss duplicate API
7. Write exhaustive unit tests (ExactMatch, SameAmountDiffDate, SimilarMerchant80%, etc.)
8. Write integration tests for full dedup flow

---

### Story 2.3: Invoice Frontend UI
**Priority:** Medium | **Points:** 5

**Description:** As a user, I want to view my invoices, see duplicate suggestions, and confirm or dismiss them.

**Acceptance Criteria:**
- [ ] Invoice list page at `/invoices` route
- [ ] Invoices display with duplicate indicator badges
- [ ] Duplicate detail view shows matched transaction
- [ ] Confirm/dismiss buttons for duplicate pairs
- [ ] Settings section for carrier code and sync configuration
- [ ] "Sync Now" button with loading state
- [ ] Pagination and date filtering
- [ ] `api/invoices.ts` and `types/invoice.ts` created
- [ ] Component tests for InvoiceList and DuplicateHandler
- [ ] E2E test for invoice sync flow

**Tasks:**
1. Create `api/invoices.ts` and `types/invoice.ts`
2. Create `components/invoice/InvoiceList.tsx`
3. Create `components/invoice/DuplicateHandler.tsx`
4. Create `pages/Invoices.tsx`
5. Add route to App.tsx
6. Add invoice section to Settings page
7. Write component tests
8. Write E2E test

---

## Epic 3: Advanced Features (Phase 4)

### Story 3.1: Advanced Charts
**Priority:** Medium | **Points:** 3

**Description:** As a user, I want advanced chart visualizations for deeper spending analysis.

**Acceptance Criteria:**
- [ ] Monthly trend line chart (income + expense over 12 months)
- [ ] Category year-over-year grouped bar chart
- [ ] Income/expense ratio stacked bar chart
- [ ] Charts page updated with new chart types
- [ ] API endpoint for trend data: `GET /api/stats/trend?months=12`
- [ ] Component tests for new chart components

**Tasks:**
1. Add `LineChart.tsx` component
2. Add trend stats backend endpoint
3. Update Charts page with new visualizations
4. Write component tests

---

### Story 3.2: Budget Management
**Priority:** Medium | **Points:** 5

**Description:** As a user, I want to set monthly budgets per category and receive over-budget warnings.

**Acceptance Criteria:**
- [ ] Budget CRUD API (`POST/GET/PUT /api/budget`)
- [ ] Budget model with user_id, category_id, monthly_amount
- [ ] Budget vs actual comparison endpoint
- [ ] Budget management page at `/budget`
- [ ] Over-budget visual indicator on dashboard
- [ ] Unit tests for budget service
- [ ] Component tests for budget UI

**Tasks:**
1. Create migration for budgets table
2. Create budget model, repository, service, handler
3. Create budget comparison endpoint
4. Create Budget page frontend
5. Add over-budget indicator to Dashboard
6. Write tests

---

### Story 3.3: Data Export
**Priority:** Medium | **Points:** 3

**Description:** As a user, I want to export my transactions as CSV or Excel files.

**Acceptance Criteria:**
- [ ] `GET /api/export/csv?start_date=...&end_date=...` returns CSV file
- [ ] `GET /api/export/excel?start_date=...&end_date=...` returns .xlsx file
- [ ] Export includes all transaction fields + category name
- [ ] Export modal in frontend with format and date range selection
- [ ] Unit tests for export service

**Tasks:**
1. Create `services/export_service.go` with CSV and Excel generation
2. Create `handlers/export.go`
3. Create export modal component
4. Write unit tests

---

### Story 3.4: Advanced Search and Tags
**Priority:** Low | **Points:** 5

**Description:** As a user, I want to search transactions by keyword and tag them for organization.

**Acceptance Criteria:**
- [ ] Full-text search on transaction descriptions
- [ ] Multi-condition filtering (type + category + date + amount range)
- [ ] Tag CRUD (add/remove tags to transactions)
- [ ] Filter by tags
- [ ] Add `tags TEXT[]` column to transactions table
- [ ] Search API: `GET /api/search/transactions?q=...&tags=...`
- [ ] Updated transaction form with tag input
- [ ] Unit tests for search service

**Tasks:**
1. Create migration adding tags column
2. Create search service with full-text query
3. Update transaction model and API for tags
4. Create search UI and tag input components
5. Write tests

---

## Epic 4: Multi-User Sharing (Phase 5)

### Story 4.1: Pairing Code System
**Priority:** Low | **Points:** 3

**Description:** As a user, I want a pairing code so I can share my financial data with family members.

**Acceptance Criteria:**
- [ ] Migration `006_sharing.sql` creates `shared_access` and `user_pairing_codes` tables
- [ ] Pairing code generated in `AB12-CD34` format
- [ ] `GET /api/shared/my-code` returns or generates code
- [ ] `POST /api/shared/regenerate-code` creates new code
- [ ] `POST /api/shared/pair` links accounts using code
- [ ] Self-pairing prevention
- [ ] Unit tests for pairing service

**Tasks:**
1. Create migration `006_sharing.sql`
2. Create models (shared_access.go, pairing_code.go)
3. Create pairing service with code generation
4. Create shared handler with endpoints
5. Write unit tests

---

### Story 4.2: Permission Middleware and View-As
**Priority:** Low | **Points:** 5

**Description:** As a user, I want to view my family member's financial data in read-only mode after they grant me access.

**Acceptance Criteria:**
- [ ] `middleware/permission.go` checks `view_as` parameter on all data endpoints
- [ ] Read-only mode enforced: all write operations rejected when viewing others' data
- [ ] `GET /api/shared/connections` lists connected accounts
- [ ] `DELETE /api/shared/connections/:uid` revokes access immediately
- [ ] All existing transaction/stats/invoice endpoints support `?view_as=`
- [ ] Security tests: unauthorized access, privilege escalation, revocation
- [ ] Permission middleware coverage >= 95%

**Tasks:**
1. Create `middleware/permission.go`
2. Apply middleware to all data routes
3. Update existing handlers to use `view_as_user_id` from context
4. Create connection management endpoints
5. Write comprehensive security tests

---

### Story 4.3: Sharing Frontend UI
**Priority:** Low | **Points:** 3

**Description:** As a user, I want a UI to manage sharing connections and switch between accounts.

**Acceptance Criteria:**
- [ ] Settings > Sharing section with pairing code and connection management
- [ ] Nav bar dropdown for account switching
- [ ] Read-only banner when viewing shared data
- [ ] Add/edit/delete buttons hidden in read-only mode
- [ ] `api/sharing.ts` and `types/sharing.ts` created
- [ ] E2E test for pairing flow
- [ ] E2E test for read-only mode enforcement

**Tasks:**
1. Create `api/sharing.ts` and `types/sharing.ts`
2. Create `components/sharing/PairingCode.tsx`
3. Create `components/sharing/ConnectionList.tsx`
4. Add account switcher to nav bar (Layout.tsx)
5. Implement read-only mode in all data pages
6. Add sharing section to Settings
7. Write E2E tests

---

## Story Priority Summary

| Story | Epic | Priority | Points |
|-------|------|----------|--------|
| 1.1 Gmail OAuth Backend | Gmail | High | 5 |
| 1.2 Gmail Scanning Service | Gmail | High | 5 |
| 1.3 Gmail Settings Frontend | Gmail | Medium | 3 |
| 2.1 Invoice Sync Backend | Invoice | High | 5 |
| 2.2 Deduplication Engine | Invoice | High | 8 |
| 2.3 Invoice Frontend UI | Invoice | Medium | 5 |
| 3.1 Advanced Charts | Advanced | Medium | 3 |
| 3.2 Budget Management | Advanced | Medium | 5 |
| 3.3 Data Export | Advanced | Medium | 3 |
| 3.4 Advanced Search & Tags | Advanced | Low | 5 |
| 4.1 Pairing Code System | Sharing | Low | 3 |
| 4.2 Permission Middleware | Sharing | Low | 5 |
| 4.3 Sharing Frontend UI | Sharing | Low | 3 |

**Total: 4 Epics, 13 Stories, 58 Story Points**
