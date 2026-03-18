# Source Tree Analysis

**Generated:** 2026-03-06

---

## Project Root

```
Billing-Note/
├── backend/                 # Go API server (Part: backend)
├── frontend/                # React SPA (Part: frontend)
├── _bmad/                   # BMad Method configuration
├── _bmad-output/            # BMad generated artifacts
├── docs/                    # Project documentation (this folder)
├── spec.md                  # Master technical specification
├── docker-compose.yml       # Docker orchestration (3 services)
├── test-and-push.sh         # Auto-test and git push script
├── start.sh                 # Quick start script
└── README.md                # Project readme
```

## Backend Structure (Go + Gin)

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # [ENTRY POINT] Server bootstrap, DI wiring, route setup
├── internal/
│   ├── handlers/                # HTTP request handlers (controller layer)
│   │   ├── auth.go              # Register, Login, Me endpoints
│   │   ├── auth_test.go
│   │   ├── transaction.go       # CRUD + stats endpoints
│   │   ├── transaction_test.go
│   │   ├── category.go          # Category listing endpoints
│   │   ├── category_test.go
│   │   ├── upload.go            # PDF upload & import endpoints
│   │   └── pdf_password.go      # PDF password CRUD endpoints
│   ├── services/                # Business logic layer
│   │   ├── auth_service.go      # Registration, login, JWT generation
│   │   ├── auth_service_test.go
│   │   ├── transaction_service.go # Transaction CRUD + stats logic
│   │   ├── upload_service.go    # PDF save, parse, import orchestration
│   │   └── pdf_password_service.go # AES encrypt/decrypt password management
│   ├── models/                  # GORM data models
│   │   ├── user.go              # User model with bcrypt password hashing
│   │   ├── user_test.go
│   │   ├── transaction.go       # Transaction model
│   │   ├── transaction_test.go
│   │   ├── category.go          # Category model
│   │   ├── category_test.go
│   │   └── pdf_password.go      # PDFPassword model with response DTO
│   ├── repository/              # Database access layer
│   │   ├── user_repo.go         # User CRUD operations
│   │   ├── user_repo_test.go
│   │   ├── transaction_repo.go  # Transaction queries with filtering
│   │   └── category_repo.go     # Category queries
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go              # JWT validation middleware
│   │   ├── auth_test.go
│   │   ├── cors.go              # CORS configuration
│   │   └── logging.go           # Request/response logging
│   └── pdf/                     # PDF parsing subsystem
│       ├── parser.go            # ParserRegistry, text extraction, password attempts
│       ├── parser_test.go
│       └── bank_parsers/        # Bank-specific parsers
│           ├── registry.go      # RegisterAllParsers factory
│           ├── cathay.go        # Cathay United Bank parser
│           ├── cathay_test.go
│           ├── taishin.go       # Taishin Bank parser
│           ├── taishin_test.go
│           ├── fubon.go         # Fubon Bank parser
│           └── fubon_test.go
├── pkg/                         # Reusable packages
│   ├── config/
│   │   └── config.go            # Environment-based configuration loader
│   ├── database/
│   │   └── database.go          # PostgreSQL connection via GORM
│   ├── crypto/
│   │   ├── aes.go               # AES-256-GCM encryption utilities
│   │   └── aes_test.go
│   ├── errors/
│   │   └── errors.go            # Custom AppError type with HTTP mapping
│   ├── logger/
│   │   └── logger.go            # Logrus-based structured logger
│   └── utils/
│       └── jwt.go               # JWT generation and validation
├── migrations/                  # SQL migration files
│   ├── 001_init.sql             # Users, categories, transactions tables
│   └── 002_pdf_passwords.sql    # PDF password table, source column
├── tests/
│   └── integration/             # Integration tests
│       ├── database_test.go
│       ├── auth_api_test.go
│       └── transaction_api_test.go
├── testdata/
│   ├── pdfs/                    # Sample PDF files for testing
│   │   ├── cathay statement PDF
│   │   ├── taishin statement PDF
│   │   └── fubon statement PDF
│   └── file_name_map.json       # PDF filename-to-bank mapping rules
├── uploads/                     # Runtime upload storage (gitignored)
├── .env.example                 # Environment variable template
├── Dockerfile                   # Backend Docker image
├── Makefile                     # Build/test commands
├── go.mod                       # Go module definition
└── go.sum                       # Go dependency checksums
```

## Frontend Structure (React + TypeScript)

```
frontend/
├── src/
│   ├── App.tsx                  # [ENTRY POINT] Route definitions, auth guard
│   ├── main.tsx                 # React DOM mount point
│   ├── index.css                # Global styles (Tailwind imports)
│   ├── components/
│   │   ├── common/              # Reusable UI components
│   │   │   ├── Button.tsx       # Styled button with variants
│   │   │   ├── Button.test.tsx
│   │   │   ├── Input.tsx        # Form input with label/error
│   │   │   └── Modal.tsx        # Dialog modal component
│   │   ├── transaction/
│   │   │   ├── TransactionList.tsx      # Transaction table with pagination
│   │   │   ├── TransactionList.test.tsx
│   │   │   └── TransactionModal.tsx     # Add/edit transaction dialog
│   │   ├── charts/
│   │   │   ├── PieChart.tsx     # Category distribution pie chart
│   │   │   ├── PieChart.test.tsx
│   │   │   ├── BarChart.tsx     # Monthly income/expense bar chart
│   │   │   └── BarChart.test.tsx
│   │   └── Layout.tsx           # App shell with navigation
│   ├── pages/
│   │   ├── Login.tsx            # Login form page
│   │   ├── Register.tsx         # Registration form page
│   │   ├── Dashboard.tsx        # Overview with stats and charts
│   │   ├── Transactions.tsx     # Transaction management page
│   │   ├── Upload.tsx           # PDF upload page
│   │   ├── Charts.tsx           # Analytics page
│   │   ├── Settings.tsx         # Settings page
│   │   └── Settings.test.tsx
│   ├── hooks/
│   │   ├── useAuth.ts           # Authentication hook
│   │   └── useTransactions.ts   # Transaction data hook
│   ├── api/
│   │   ├── client.ts            # Axios instance with JWT interceptor
│   │   ├── auth.ts              # Auth API calls
│   │   ├── transactions.ts      # Transaction API calls
│   │   └── categories.ts        # Category API calls
│   ├── types/
│   │   ├── transaction.ts       # Transaction, Category, Filter types
│   │   ├── auth.ts              # User, LoginRequest, AuthResponse types
│   │   └── api.ts               # MonthlyStats, CategoryStats types
│   ├── utils/
│   │   ├── format.ts            # Number/date formatting utilities
│   │   ├── format.test.ts
│   │   ├── validation.ts        # Form validation helpers
│   │   └── validation.test.ts
│   ├── store/
│   │   └── authStore.ts         # Zustand auth state (token, user, localStorage)
│   └── tests/
│       └── setup.ts             # Vitest test setup
├── tests/
│   └── e2e/                     # Playwright E2E tests
│       ├── transaction-list.spec.ts
│       └── charts.spec.ts
├── public/                      # Static assets
├── index.html                   # HTML entry point
├── package.json                 # Dependencies and scripts
├── tsconfig.json                # TypeScript configuration
├── tsconfig.node.json           # Node TypeScript config
├── vite.config.ts               # Vite build configuration
├── vitest.config.ts             # Vitest test configuration
├── tailwind.config.js           # Tailwind CSS configuration
├── postcss.config.js            # PostCSS configuration
├── playwright.config.ts         # Playwright E2E configuration
├── .env.example                 # Frontend environment template
├── Dockerfile                   # Frontend Docker image (nginx)
├── nginx.conf                   # Nginx reverse proxy config
└── .gitignore
```

## Critical Folders Summary

| Folder | Purpose | Importance |
|--------|---------|------------|
| `backend/internal/handlers/` | HTTP request handling, input validation | High - API surface |
| `backend/internal/services/` | Core business logic | High - Business rules |
| `backend/internal/models/` | Data structures, GORM models | High - Data contracts |
| `backend/internal/pdf/` | PDF parsing subsystem | High - Core feature |
| `backend/pkg/config/` | Configuration management | Medium - Environment setup |
| `backend/migrations/` | Database schema | High - Data structure |
| `frontend/src/api/` | Backend API client layer | High - Integration point |
| `frontend/src/components/` | UI components | High - User interface |
| `frontend/src/store/` | Client-side state | Medium - Auth state |
| `frontend/src/pages/` | Route-level pages | High - User flows |
