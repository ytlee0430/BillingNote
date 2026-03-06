# Billing Note - Project Documentation Index

**Generated:** 2026-03-06
**Scan Level:** Deep
**Workflow Mode:** initial_scan

---

## Project Overview

- **Type:** Multi-part (monorepo) with 2 parts
- **Primary Language:** Go 1.23 (backend), TypeScript 5.3 (frontend)
- **Architecture:** REST API + SPA, layered architecture
- **Database:** PostgreSQL 15+
- **Deployment:** Docker Compose

### Quick Reference

#### Backend (backend/)
- **Type:** Go REST API
- **Framework:** Gin 1.9.1 + GORM 1.25.5
- **Entry Point:** `cmd/server/main.go`
- **Port:** 8080

#### Frontend (frontend/)
- **Type:** React SPA
- **Framework:** React 18 + Vite 5 + TypeScript
- **Entry Point:** `src/App.tsx`
- **Port:** 5173 (dev) / 80 (Docker)

---

## Generated Documentation

- [Project Overview](./project-overview.md) - Executive summary, tech stack, status
- [Source Tree Analysis](./source-tree-analysis.md) - Annotated directory structure
- [Architecture - Backend](./architecture-backend.md) - Go API architecture and patterns
- [Architecture - Frontend](./architecture-frontend.md) - React SPA architecture and patterns
- [API Contracts](./api-contracts.md) - REST API endpoints and schemas
- [Data Models](./data-models.md) - Database schema and entity relationships
- [Integration Architecture](./integration-architecture.md) - Frontend-Backend communication
- [Development Guide](./development-guide.md) - Setup, commands, testing

## Existing Documentation

- [README.md](../README.md) - Project readme with quick start
- [spec.md](../spec.md) - Master technical specification (Phase 1-5)

---

## Getting Started

1. Review this index for navigation
2. Read [Project Overview](./project-overview.md) for executive summary
3. Read [Development Guide](./development-guide.md) for setup instructions
4. Reference [API Contracts](./api-contracts.md) for endpoint details
5. Reference [Data Models](./data-models.md) for database schema

### For AI-Assisted Development

When planning new features or creating PRDs, provide this index as context input. Key references:
- **Architecture decisions:** `architecture-backend.md` + `architecture-frontend.md`
- **Existing API surface:** `api-contracts.md`
- **Database schema:** `data-models.md`
- **Integration patterns:** `integration-architecture.md`
- **Full spec:** `../spec.md`
