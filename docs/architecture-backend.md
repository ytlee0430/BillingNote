# Architecture - Backend

**Generated:** 2026-03-06
**Part:** backend
**Type:** Go REST API

---

## Architecture Pattern

**Layered Architecture** with clean separation of concerns:

```
HTTP Request
    |
    v
[Middleware] --> Auth (JWT), CORS, Logging
    |
    v
[Handlers]  --> Input validation, response formatting
    |
    v
[Services]  --> Business logic, orchestration
    |
    v
[Repository]--> Database queries (GORM)
    |
    v
[PostgreSQL]
```

## Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.23 |
| Web Framework | Gin | 1.9.1 |
| ORM | GORM | 1.25.5 |
| Database | PostgreSQL | 15+ |
| Auth | golang-jwt/jwt | v5.2.0 |
| Password Hashing | bcrypt | (crypto stdlib) |
| PDF Parsing | pdfcpu | 0.6.0 |
| Encryption | AES-256-GCM | (crypto stdlib) |
| Logging | logrus | 1.9.3 |
| Config | godotenv | 1.5.1 |
| Testing | testify + go-sqlmock | 1.8.4 / 1.5.2 |

## Entry Point and Bootstrap

**File:** `cmd/server/main.go`

Bootstrap sequence:
1. Initialize logger (level from `LOG_LEVEL` env)
2. Load configuration from `.env` file
3. Connect to PostgreSQL via GORM
4. Initialize repositories (User, Category, Transaction)
5. Initialize services (Auth, Transaction, PDFPassword, Upload)
6. Initialize handlers
7. Configure Gin with middleware (Recovery, Logging, CORS)
8. Register routes (public auth + protected API)
9. Start HTTP server on configured port

## Dependency Injection

Manual constructor-based DI (no framework):

```
Config -> Database -> Repositories -> Services -> Handlers -> Routes
```

All services accept interfaces, enabling mock injection for testing.

## Authentication Flow

1. **Registration:** `POST /api/auth/register` -> bcrypt hash -> store user -> return JWT
2. **Login:** `POST /api/auth/login` -> verify bcrypt hash -> generate JWT
3. **Protected routes:** `AuthMiddleware` extracts JWT from `Authorization: Bearer <token>`, validates, sets `user_id` and `user_email` in Gin context

## PDF Parsing Subsystem

```
Upload -> Save to disk -> Extract text (with password attempts) -> Match bank parser -> Parse transactions
```

**Strategy Pattern:**
- `BankParser` interface: `CanParse(content) bool`, `Parse(content) ([]Transaction, error)`
- `ParserRegistry`: Holds all parsers, tries each via `CanParse`
- Current parsers: Cathay, Taishin, Fubon

**Password handling:**
- 4 password slots per user, AES-256-GCM encrypted in DB
- Passwords tried in priority order (1-4)
- File name rules can auto-match passwords to specific banks

## Error Handling

Custom `AppError` type with:
- Error code (string)
- HTTP status code
- User-friendly message
- Optional details
- Trace ID (request_id)

Factory functions: `NewValidationError`, `NewUnauthorizedError`, `NewNotFoundError`, `NewInternalError`

## Middleware Stack

1. **Recovery** - Panic recovery (Gin built-in)
2. **LoggingMiddleware** - Request/response logging with request_id
3. **CORSMiddleware** - Configurable CORS origins
4. **AuthMiddleware** - JWT validation (applied to `/api/*` routes)

## Data Architecture

See [Data Models](./data-models.md) for full schema.

Key tables: `users`, `categories`, `transactions`, `user_pdf_passwords`

## Testing Strategy

- **Unit tests:** Per-file `_test.go` files using testify assertions
- **DB mocking:** go-sqlmock for repository tests
- **Integration tests:** `tests/integration/` directory with real DB tests
- **Test data:** `testdata/pdfs/` for PDF parser tests
- **Coverage target:** >= 80%
