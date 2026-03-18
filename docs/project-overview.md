# Billing Note - Project Overview

**Generated:** 2026-03-06
**Scan Level:** Deep
**Project Type:** Multi-part (Backend + Frontend)

---

## Executive Summary

Billing Note is a personal web-based accounting system designed to automate financial data import from credit card statements (PDF), cloud invoices (Taiwan MOF API), and Gmail. The project follows a phased development approach (Phase 1-5) and currently has Phase 1 (MVP) and Phase 2A (PDF import) substantially implemented.

## Project Purpose

- Reduce manual data entry for personal finance tracking
- Auto-import credit card bills from encrypted PDFs (Cathay, Taishin, Fubon banks)
- Integrate with Taiwan's Ministry of Finance e-invoice API
- Auto-scan Gmail for credit card statement attachments
- Provide visual spending analytics (charts, category breakdowns)
- Support multi-user read-only sharing (Phase 5)

## Repository Structure

| Aspect | Details |
|--------|---------|
| **Repository Type** | Multi-part (monorepo) |
| **Parts** | `backend/` (Go API), `frontend/` (React SPA) |
| **Primary Language** | Go 1.23 (backend), TypeScript 5.3 (frontend) |
| **Architecture** | REST API + SPA, layered architecture |
| **Database** | PostgreSQL 15+ |
| **Deployment** | Docker Compose (3 services: backend, frontend, db) |

## Current Implementation Status

### Phase 1 - Core MVP (Mostly Complete)
- User registration/login with JWT
- Manual transaction CRUD
- Transaction listing with pagination and filtering
- Basic charts (monthly bar chart, category pie chart)
- Settings page
- Category management (14 default categories)

### Phase 2A - PDF Import (Partially Complete)
- PDF password management (AES-256 encrypted, 4 slots)
- PDF upload and parsing pipeline
- Bank parsers: Cathay, Taishin, Fubon
- Transaction preview and import from parsed PDFs

### Phase 2B-5 (Not Started)
- Gmail auto-scan for PDF attachments
- Cloud invoice integration (MOF API)
- Advanced features (budgets, export, tags)
- Multi-user sharing with read-only mode

## Technology Decision Table

| Category | Technology | Version | Justification |
|----------|-----------|---------|---------------|
| Backend Framework | Gin | 1.9.1 | Lightweight, high-performance Go web framework |
| ORM | GORM | 1.25.5 | Full-featured Go ORM with PostgreSQL support |
| Database | PostgreSQL | 15+ | Robust relational DB, JSONB support for flexibility |
| Auth | golang-jwt/jwt | 5.2.0 | Industry standard JWT implementation |
| PDF Parsing | pdfcpu | 0.6.0 | Pure Go PDF processor with encryption support |
| Logging | logrus | 1.9.3 | Structured logging with field support |
| Encryption | crypto/aes | stdlib | AES-256 for PDF password storage |
| Frontend Framework | React | 18.2 | Component-based UI with large ecosystem |
| Build Tool | Vite | 5.0 | Fast HMR, native TypeScript support |
| State Management | Zustand | 4.4.7 | Minimal, hook-based state management |
| Data Fetching | TanStack Query | 5.17.9 | Server state management with caching |
| Routing | React Router | 6.21.1 | Standard React routing solution |
| Styling | Tailwind CSS | 3.4.1 | Utility-first CSS framework |
| Charts | Recharts | 2.10.3 | React-native chart library |
| Forms | React Hook Form | 7.49.3 | Performant form handling |
| HTTP Client | Axios | 1.6.5 | HTTP client with interceptors |
| Unit Testing (BE) | testify + go-sqlmock | 1.8.4 | Go standard test assertions and DB mocking |
| Unit Testing (FE) | Vitest | 1.2.1 | Vite-native test runner, Jest-compatible |
| Component Testing | React Testing Library | 14.1.2 | User-behavior focused component testing |
| E2E Testing | Playwright | 1.41.0 | Multi-browser E2E automation |

## Key Architectural Patterns

1. **Layered Architecture (Backend):** Handler -> Service -> Repository -> Database
2. **Interface-based Design:** Services use interfaces for testability (mock injection)
3. **Strategy Pattern (PDF):** `BankParser` interface with registry for bank-specific parsers
4. **JWT Middleware:** Route-level authentication with `AuthMiddleware`
5. **Structured Error Handling:** Custom `AppError` type with HTTP status mapping
6. **Zustand Store (Frontend):** Centralized auth state with localStorage persistence
7. **TanStack Query (Frontend):** Server-state caching for API data

## Links to Detailed Documentation

- [Source Tree Analysis](./source-tree-analysis.md)
- [Architecture - Backend](./architecture-backend.md)
- [Architecture - Frontend](./architecture-frontend.md)
- [API Contracts](./api-contracts.md)
- [Data Models](./data-models.md)
- [Development Guide](./development-guide.md)
- [Integration Architecture](./integration-architecture.md)
