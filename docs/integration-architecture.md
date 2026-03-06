# Integration Architecture

**Generated:** 2026-03-06

---

## Part Communication

```
┌──────────────────┐         ┌──────────────────┐
│     Frontend     │  HTTPS  │     Backend      │
│   (React SPA)    │ ------> │   (Go + Gin)     │
│   Port: 5173     │  REST   │   Port: 8080     │
│                  │ <------ │                  │
└──────────────────┘   JSON  └────────┬─────────┘
                                      │
                                      │ GORM
                                      v
                              ┌──────────────────┐
                              │   PostgreSQL     │
                              │   Port: 5432     │
                              └──────────────────┘
```

## Integration Points

### 1. Frontend -> Backend (REST API)

| From | To | Type | Details |
|------|-----|------|---------|
| `frontend/src/api/client.ts` | `backend :8080/api/*` | REST/JSON | Axios with JWT Bearer token |
| `frontend/src/api/auth.ts` | `POST /api/auth/*` | REST/JSON | Login, register, me |
| `frontend/src/api/transactions.ts` | `GET/POST/PUT/DELETE /api/transactions/*` | REST/JSON | Transaction CRUD |
| `frontend/src/api/categories.ts` | `GET /api/categories/*` | REST/JSON | Category listing |
| Upload page | `POST /api/upload/pdf` | multipart/form-data | PDF file upload |

### 2. Backend -> PostgreSQL (GORM)

| From | To | Type | Details |
|------|-----|------|---------|
| `pkg/database/database.go` | PostgreSQL | TCP/5432 | GORM connection with DSN from config |
| Repository layer | DB tables | SQL via GORM | Auto-migration not used; manual SQL migrations |

## Authentication Flow

```
Frontend                        Backend                    Database
   |                              |                          |
   |-- POST /api/auth/login ----->|                          |
   |                              |-- Find user by email --->|
   |                              |<-- User record ----------|
   |                              |-- Verify bcrypt hash     |
   |                              |-- Generate JWT           |
   |<-- { token, user } ---------|                          |
   |                              |                          |
   |-- Store in localStorage      |                          |
   |                              |                          |
   |-- GET /api/transactions ---->|                          |
   |   Authorization: Bearer JWT  |                          |
   |                              |-- AuthMiddleware         |
   |                              |   Validate JWT           |
   |                              |   Extract user_id        |
   |                              |-- Query with user_id --->|
   |                              |<-- Transaction data -----|
   |<-- { data, total, page } ---|                          |
```

## PDF Upload Flow

```
Frontend                        Backend                    Disk
   |                              |                          |
   |-- POST /api/upload/pdf ----->|                          |
   |   (multipart/form-data)      |                          |
   |                              |-- Save PDF to ---------->|
   |                              |   /uploads/{user}/{file}  |
   |                              |                          |
   |                              |-- Get user passwords     |
   |                              |   (decrypt from DB)      |
   |                              |                          |
   |                              |-- Try decrypt PDF ------>|
   |                              |   (password #1, #2, ...) |
   |                              |                          |
   |                              |-- Extract text           |
   |                              |-- Match bank parser      |
   |                              |-- Parse transactions     |
   |                              |                          |
   |<-- { results: [...] } ------|                          |
   |                              |                          |
   |-- POST /api/transactions/ -->|                          |
   |   import (confirmed txns)    |-- Insert transactions -->|
   |<-- { imported: N } ---------|                          |
```

## Docker Compose Topology

```yaml
Services:
  frontend (port 80)  --> nginx serving built React app
      |
      | (proxy_pass /api -> backend:8080)
      v
  backend (port 8080) --> Go Gin server
      |
      v
  db (port 5432)      --> PostgreSQL 15
```

**Networks:** `billing_network` (bridge)
**Volumes:** `postgres_data` (DB persistence), `backend_uploads` (PDF storage)

## CORS Configuration

Backend allows origins configured via `ALLOWED_ORIGINS` env var.
Default development: `http://localhost:5173` (Vite dev server)
Docker: `http://localhost` (nginx on port 80)
