# Development Guide

**Generated:** 2026-03-06

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.23+ | Backend runtime |
| Node.js | 18+ | Frontend build/dev |
| PostgreSQL | 15+ | Database |
| npm | 9+ | Package manager |
| Docker & Docker Compose | Latest | Containerized deployment |

## Quick Start

### Option 1: Local Development

```bash
# 1. Clone and enter project
git clone <repo-url>
cd Billing-Note

# 2. Setup database
createdb billing_note
cd backend
psql -d billing_note -f migrations/001_init.sql
psql -d billing_note -f migrations/002_pdf_passwords.sql

# 3. Configure backend
cp .env.example .env
# Edit .env with your DB credentials

# 4. Start backend
make run
# Server runs on http://localhost:8080

# 5. In another terminal, setup frontend
cd frontend
npm install
npm run dev
# Dev server runs on http://localhost:5173
```

### Option 2: Docker Compose

```bash
docker-compose up --build
# Frontend: http://localhost
# Backend:  http://localhost:8080
# Database: localhost:5432
```

## Environment Variables

### Backend (.env)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Server port |
| `GIN_MODE` | debug | Gin mode (debug/release) |
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | DB username |
| `DB_PASSWORD` | postgres | DB password |
| `DB_NAME` | billing_note | Database name |
| `DB_SSLMODE` | disable | SSL mode |
| `JWT_SECRET` | (change this) | JWT signing key |
| `JWT_EXPIRY` | 24h | Token expiration |
| `ALLOWED_ORIGINS` | http://localhost:5173 | CORS allowed origins |
| `UPLOAD_DIR` | ./uploads | PDF upload directory |
| `MAX_UPLOAD_SIZE` | 10485760 | Max upload size (10MB) |
| `ENCRYPTION_KEY` | (change this) | AES-256 key for PDF passwords |
| `LOG_LEVEL` | info | Logging level |
| `LOG_FORMAT` | text | Logging format (text/json) |

### Frontend (.env)

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_URL` | http://localhost:8080 | Backend API base URL |

## Development Commands

### Backend

```bash
cd backend

make run              # Start dev server
make build            # Build binary to bin/server
make test             # Run all tests
make test-coverage    # Generate coverage report
make clean            # Remove build artifacts
make deps             # Download and tidy dependencies
make migrate-up       # Run SQL migrations
make test-push        # Test + auto commit/push
```

### Frontend

```bash
cd frontend

npm run dev           # Start Vite dev server (HMR)
npm run build         # Build for production
npm run preview       # Preview production build
npm run test          # Run Vitest (watch mode)
npm run test:ui       # Vitest with UI
npm run test:coverage # Coverage report
npm run e2e           # Run Playwright E2E tests
npm run e2e:headed    # E2E with visible browser
npm run e2e:debug     # E2E in debug mode
npm run lint          # ESLint check
npm run test:push     # Test + auto commit/push
```

### Root Level

```bash
./test-and-push.sh                    # Run all tests + auto push
./test-and-push.sh "commit message"   # With custom message
```

## Testing

### Backend Tests

Tests are co-located with source files (`*_test.go`).

```bash
# All tests
go test ./... -v

# Specific package
go test ./internal/services -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

**Key test files:**
- `internal/models/*_test.go` - Model validation
- `internal/handlers/*_test.go` - Handler HTTP tests
- `internal/services/*_test.go` - Service logic tests
- `internal/repository/*_test.go` - Repository tests (go-sqlmock)
- `internal/pdf/*_test.go` - PDF parser tests
- `tests/integration/*_test.go` - Integration tests

### Frontend Tests

```bash
# Unit/component tests (Vitest)
npm run test

# E2E tests (Playwright)
npm run e2e
```

**Key test files:**
- `src/utils/*.test.ts` - Utility function tests
- `src/components/**/*.test.tsx` - Component tests
- `tests/e2e/*.spec.ts` - E2E user flow tests

## Database Migrations

Migrations are plain SQL files in `backend/migrations/`.

```bash
# Apply all migrations
cd backend
psql -h localhost -U postgres -d billing_note -f migrations/001_init.sql
psql -h localhost -U postgres -d billing_note -f migrations/002_pdf_passwords.sql
```

No migration tool is used; migrations are idempotent (`IF NOT EXISTS`).

## Code Conventions

- **Backend:** Standard Go project layout (`cmd/`, `internal/`, `pkg/`)
- **Naming:** Go standard (PascalCase exports, camelCase private)
- **Error handling:** Return `error` from functions, use `AppError` in handlers
- **Frontend:** TypeScript strict mode, React functional components with hooks
- **Styling:** Tailwind utility classes, no CSS modules
- **API types:** Defined in `frontend/src/types/`, matching backend JSON contracts
