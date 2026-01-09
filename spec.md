# 記帳系統技術規格文件 (Billing Note System)

**版本：** 1.2
**日期：** 2026-01-08
**作者：** BMAD Team
**核心價值：** 自動化減少手動輸入，整合多來源財務資料

---

## 目錄

1. [專案概述](#專案概述)
2. [技術架構](#技術架構)
3. [測試策略](#測試策略)
4. [Phase 劃分](#phase-劃分)
5. [資料模型](#資料模型)
6. [API 設計](#api-設計)
7. [PDF 處理流程](#pdf-處理流程)
8. [Gmail 整合流程](#gmail-整合流程)
9. [部署規劃](#部署規劃)

---

## 專案概述

### 目標
建立一個個人使用的 Web 記帳系統，透過自動化匯入信用卡帳單、雲端發票、電子存摺等資料來源，減少手動輸入工作，並提供視覺化的消費分析。

### 核心功能
1. ✅ **PDF 信用卡帳單解析**（加密 PDF 支援）
2. ✅ **雲端發票整合**（財政部 API）
3. ⚠️ **電子存摺匯入**（Phase 4）
4. ✅ **多來源資料去重複**
5. ✅ **消費/收入分類與圖表**
6. ✅ **手動輸入資料**

### 用戶角色
- **Phase 1-4：** 單一用戶（個人使用）
- **Phase 5：** 多用戶互相檢視（每人有自己的交易資料，可授權他人檢視）

---

## 技術架構

### 技術棧

#### 後端 - GO
- **Framework:** Gin (https://github.com/gin-gonic/gin)
- **ORM:** GORM (https://gorm.io)
- **Database:** PostgreSQL 15+
- **PDF 解析:** unipdf (https://github.com/unidoc/unipdf) 或 pdfcpu
- **JWT 認證:** golang-jwt/jwt (https://github.com/golang-jwt/jwt)
- **Gmail API:** google.golang.org/api/gmail/v1
- **加密:** crypto/aes (標準庫)

**專案結構：**
```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # 入口
├── internal/
│   ├── handlers/                   # HTTP handlers
│   │   ├── auth.go
│   │   ├── transaction.go
│   │   ├── upload.go
│   │   └── gmail.go
│   ├── services/                   # 業務邏輯
│   │   ├── pdf_parser.go
│   │   ├── invoice_service.go
│   │   ├── deduplication.go
│   │   └── gmail_service.go
│   ├── models/                     # 資料模型
│   │   ├── transaction.go
│   │   ├── user.go
│   │   └── pdf_password.go
│   ├── repository/                 # 資料庫操作
│   │   ├── transaction_repo.go
│   │   └── user_repo.go
│   └── pdf/                        # PDF 解析器
│       ├── parser.go
│       └── bank_parsers/
│           ├── cathay.go           # 國泰世華
│           ├── esun.go             # 玉山
│           └── chinatrust.go       # 中信
├── pkg/                            # 可重用套件
│   ├── crypto/                     # 加密工具
│   └── utils/
├── migrations/                     # DB migrations
│   └── 001_init.sql
├── uploads/                        # 上傳檔案儲存 (gitignore)
├── config/
│   └── config.yaml                 # 配置檔
├── go.mod
└── go.sum
```

#### 前端 - React + TypeScript
- **Build Tool:** Vite 5+
- **UI Framework:** React 18+
- **語言:** TypeScript 5+
- **狀態管理:** Zustand
- **資料獲取:** TanStack Query (React Query)
- **路由:** React Router v6
- **UI 樣式:** Tailwind CSS 3+
- **圖表庫:** Recharts
- **HTTP Client:** Axios
- **表單處理:** React Hook Form
- **日期處理:** date-fns

**專案結構：**
```
frontend/
├── src/
│   ├── components/                 # 可重用元件
│   │   ├── common/
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   └── Modal.tsx
│   │   ├── transaction/
│   │   │   ├── TransactionList.tsx
│   │   │   ├── TransactionForm.tsx
│   │   │   └── TransactionPreview.tsx
│   │   └── charts/
│   │       ├── PieChart.tsx
│   │       ├── BarChart.tsx
│   │       └── LineChart.tsx
│   ├── pages/                      # 頁面
│   │   ├── Dashboard.tsx
│   │   ├── Upload.tsx
│   │   ├── Settings.tsx
│   │   ├── Transactions.tsx
│   │   └── Charts.tsx
│   ├── hooks/                      # Custom hooks
│   │   ├── useTransactions.ts
│   │   ├── useUpload.ts
│   │   └── useAuth.ts
│   ├── api/                        # API 呼叫
│   │   ├── client.ts
│   │   ├── transactions.ts
│   │   ├── upload.ts
│   │   └── gmail.ts
│   ├── types/                      # TypeScript 型別
│   │   ├── transaction.ts
│   │   ├── api.ts
│   │   └── chart.ts
│   ├── utils/                      # 工具函數
│   │   ├── format.ts
│   │   └── validation.ts
│   ├── store/                      # Zustand store
│   │   └── authStore.ts
│   ├── App.tsx
│   └── main.tsx
├── public/
├── index.html
├── package.json
├── tsconfig.json
├── tailwind.config.js
└── vite.config.ts
```

### 系統架構圖

```
┌─────────────┐
│   Browser   │
│  (React)    │
└──────┬──────┘
       │ HTTPS
       ▼
┌─────────────────┐
│  GO Backend     │
│  (Gin Server)   │
├─────────────────┤
│ - PDF Parser    │
│ - Gmail API     │
│ - Invoice API   │
│ - Dedup Logic   │
└────┬────┬───────┘
     │    │
     │    └──────────┐
     ▼               ▼
┌──────────┐   ┌─────────────┐
│PostgreSQL│   │File Storage │
│          │   │  /uploads/  │
└──────────┘   └─────────────┘
```

---

## 測試策略

### 測試原則

**核心要求：**
- ✅ 每個 function 都必須有 unit test（100% 函式覆蓋）
- ✅ 每個 Phase 交付前所有測試必須 PASS
- ✅ 測試覆蓋率目標：後端 ≥ 80%、前端 ≥ 75%
- ✅ CI/CD 自動化測試（測試失敗 = 無法部署）
- ✅ 關鍵功能必須有 E2E 測試

### 測試金字塔

```
           ╱ ╲
          ╱E2E ╲         10% - UI 自動化測試（Playwright）
         ╱------╲
        ╱ 整合   ╲       20% - API/整合測試
       ╱  測試   ╲
      ╱----------╲
     ╱   Unit     ╲     70% - 單元測試（每個 function）
    ╱   Tests     ╲
   ╱---------------╲
```

**分層測試職責：**
- **Unit Tests (70%):** 測試單一函式、方法的邏輯正確性
- **Integration Tests (20%):** 測試 API、資料庫、外部服務整合
- **E2E Tests (10%):** 測試完整用戶流程

---

### 測試工具與套件

#### 後端測試（GO）

**1. Unit Test Framework:**
- **testing** - GO 標準庫（內建）
- **testify** (https://github.com/stretchr/testify) - 斷言庫
  ```bash
  go get github.com/stretchr/testify/assert
  go get github.com/stretchr/testify/mock
  ```

**2. Mock Framework:**
- **gomock** (https://github.com/golang/mock) - 官方 mock 工具
  ```bash
  go install github.com/golang/mock/mockgen@latest
  ```
- **testify/mock** - 簡單 mock

**3. Database Testing:**
- **go-sqlmock** (https://github.com/DATA-DOG/go-sqlmock) - SQL mock
- **testcontainers-go** - 真實 PostgreSQL 容器測試

**4. HTTP Testing:**
- **httptest** (GO 標準庫) - HTTP handler 測試
- **Gin 測試工具** - 內建測試支援

**執行指令：**
```bash
# 執行所有測試
go test ./... -v

# 執行測試 + 覆蓋率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 執行特定測試
go test ./internal/services -v

# 平行執行測試
go test ./... -v -parallel=4
```

---

#### 前端測試（React + TypeScript）

**1. Unit Test Framework:**
- **Vitest** (https://vitest.dev/) - Vite 原生，超快速
  - 與 Vite 完美整合
  - 兼容 Jest API
  - 內建 TypeScript 支援
  ```bash
  npm install -D vitest @vitest/ui
  ```

**2. React Testing:**
- **React Testing Library** (https://testing-library.com/react)
  - 組件測試
  - 鼓勵測試用戶行為而非實作細節
  ```bash
  npm install -D @testing-library/react @testing-library/jest-dom
  npm install -D @testing-library/user-event
  ```

**3. E2E / UI 自動化測試：**
- **Playwright** ✅ (https://playwright.dev/) - **強烈推薦**
  - 微軟官方維護
  - 支援多瀏覽器（Chrome、Firefox、Safari、Edge）
  - 自動等待機制（減少 flaky tests）
  - 內建截圖、錄影
  - 平行執行
  - TypeScript 原生支援
  ```bash
  npm install -D @playwright/test
  npx playwright install
  ```

**為什麼選 Playwright？**

| 特性 | Playwright ✅ | Cypress | Selenium |
|------|--------------|---------|----------|
| 速度 | 非常快 | 中等 | 慢 |
| 多瀏覽器 | 全支援 | 有限 | 全支援 |
| 自動等待 | 內建 | 內建 | 需手動 |
| 平行執行 | 支援 | 付費版 | 支援 |
| 學習曲線 | 低 | 低 | 高 |
| 穩定性 | 高 | 中 | 中 |
| TypeScript | 原生支援 | 支援 | 需配置 |

**執行指令：**
```bash
# Unit tests
npm run test              # 執行所有單元測試
npm run test:ui           # UI 模式
npm run test:coverage     # 覆蓋率報告

# E2E tests
npx playwright test                    # 執行所有 E2E
npx playwright test --headed           # 顯示瀏覽器
npx playwright test --debug            # Debug 模式
npx playwright test upload-pdf.spec.ts # 執行特定測試
npx playwright show-report             # 查看報告
```

---

### 測試專案結構

#### 後端測試結構
```
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── auth_test.go              ← Unit test
│   │   ├── transaction.go
│   │   ├── transaction_test.go       ← Unit test
│   │   ├── upload.go
│   │   └── upload_test.go            ← Unit test
│   ├── services/
│   │   ├── pdf_parser.go
│   │   ├── pdf_parser_test.go        ← Unit test
│   │   ├── gmail_service.go
│   │   ├── gmail_service_test.go     ← Unit test
│   │   ├── deduplication.go
│   │   └── deduplication_test.go     ← Unit test (critical!)
│   ├── repository/
│   │   ├── transaction_repo.go
│   │   ├── transaction_repo_test.go  ← Unit test
│   │   ├── user_repo.go
│   │   └── user_repo_test.go         ← Unit test
│   ├── models/
│   │   ├── transaction.go
│   │   ├── transaction_test.go       ← Model validation test
│   │   ├── user.go
│   │   └── user_test.go              ← Model validation test
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── auth_test.go              ← Unit test
│   │   ├── permission.go
│   │   └── permission_test.go        ← Unit test
│   └── pdf/
│       ├── parser.go
│       ├── parser_test.go            ← Unit test
│       └── bank_parsers/
│           ├── cathay.go
│           ├── cathay_test.go        ← Bank-specific test
│           ├── esun.go
│           └── esun_test.go          ← Bank-specific test
├── tests/
│   ├── integration/                   ← Integration tests
│   │   ├── api_test.go               # API 整合測試
│   │   ├── auth_api_test.go
│   │   ├── transaction_api_test.go
│   │   ├── upload_api_test.go
│   │   ├── gmail_api_test.go
│   │   ├── invoice_api_test.go
│   │   ├── shared_access_test.go
│   │   └── database_test.go          # DB 整合測試
│   ├── fixtures/                      ← 測試資料
│   │   ├── pdfs/
│   │   │   ├── sample-cathay.pdf
│   │   │   ├── sample-esun-encrypted.pdf
│   │   │   └── invalid.pdf
│   │   ├── test_users.sql
│   │   └── test_transactions.sql
│   └── helpers/
│       ├── test_db.go                # 測試資料庫工具
│       ├── mock_gmail.go             # Gmail API mock
│       └── fixtures.go               # 測試資料生成
├── coverage/                          ← 覆蓋率報告
│   ├── coverage.out
│   └── coverage.html
└── Makefile
    # 測試指令快捷方式
```

#### 前端測試結構
```
frontend/
├── src/
│   ├── components/
│   │   ├── common/
│   │   │   ├── Button.tsx
│   │   │   ├── Button.test.tsx       ← Component test
│   │   │   ├── Input.tsx
│   │   │   ├── Input.test.tsx        ← Component test
│   │   │   ├── Modal.tsx
│   │   │   └── Modal.test.tsx        ← Component test
│   │   ├── transaction/
│   │   │   ├── TransactionList.tsx
│   │   │   ├── TransactionList.test.tsx
│   │   │   ├── TransactionForm.tsx
│   │   │   └── TransactionForm.test.tsx
│   │   └── charts/
│   │       ├── PieChart.tsx
│   │       ├── PieChart.test.tsx
│   │       ├── BarChart.tsx
│   │       └── BarChart.test.tsx
│   ├── hooks/
│   │   ├── useTransactions.ts
│   │   ├── useTransactions.test.ts   ← Hook test
│   │   ├── useAuth.ts
│   │   ├── useAuth.test.ts           ← Hook test
│   │   ├── useUpload.ts
│   │   └── useUpload.test.ts         ← Hook test
│   ├── utils/
│   │   ├── format.ts
│   │   ├── format.test.ts            ← Utils test
│   │   ├── validation.ts
│   │   └── validation.test.ts        ← Utils test
│   ├── api/
│   │   ├── client.ts
│   │   ├── client.test.ts            ← API client test
│   │   ├── transactions.ts
│   │   └── transactions.test.ts      ← API test
│   └── pages/
│       ├── Dashboard.tsx
│       ├── Dashboard.test.tsx        ← Page test
│       ├── Upload.tsx
│       ├── Upload.test.tsx           ← Page test
│       └── Settings.tsx
│           └── Settings.test.tsx     ← Page test
├── tests/
│   ├── e2e/                           ← Playwright E2E tests
│   │   ├── auth.spec.ts              # 認證流程
│   │   ├── manual-transaction.spec.ts # 手動新增交易
│   │   ├── transaction-list.spec.ts  # 交易列表
│   │   ├── upload-pdf.spec.ts        # PDF 上傳
│   │   ├── pdf-password.spec.ts      # 密碼設定
│   │   ├── gmail-connect.spec.ts     # Gmail 連結
│   │   ├── invoice-sync.spec.ts      # 發票同步
│   │   ├── charts.spec.ts            # 圖表
│   │   ├── pairing.spec.ts           # 配對功能
│   │   └── read-only-mode.spec.ts    # 唯讀模式
│   ├── setup.ts                       # 測試設定
│   ├── helpers/
│   │   ├── test-utils.tsx            # 測試工具（render with providers）
│   │   └── mock-data.ts              # Mock 資料
│   └── fixtures/
│       └── sample-pdfs/              # E2E 測試用 PDF
├── coverage/                          ← 覆蓋率報告
├── playwright-report/                 ← Playwright 報告
├── playwright.config.ts               ← Playwright 配置
├── vitest.config.ts                   ← Vitest 配置
└── package.json
```

---

### 測試覆蓋率目標

**後端（GO）：**
- **總體覆蓋率：≥ 80%**
- **關鍵模組要求：**
  - `services/` - ≥ 85%（業務邏輯核心）
  - `handlers/` - ≥ 80%
  - `repository/` - ≥ 75%
  - `middleware/` - ≥ 90%（安全相關）
  - `pdf/` - ≥ 85%（PDF 解析）

**前端（React）：**
- **總體覆蓋率：≥ 75%**
- **關鍵模組要求：**
  - `hooks/` - ≥ 80%
  - `utils/` - ≥ 85%
  - `components/common/` - ≥ 75%
  - `api/` - ≥ 70%

**E2E 覆蓋率：**
- **關鍵用戶流程覆蓋率：100%**
  - 註冊/登入
  - 手動新增交易
  - PDF 上傳與匯入
  - 圖表查看
  - 多人共享（Phase 5）

---

### CI/CD 測試流程

**GitHub Actions Pipeline：**

```yaml
# .github/workflows/test.yml
name: Test Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  backend-unit-test:
    name: Backend Unit Tests
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Cache Go Modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: Install Dependencies
        run: |
          cd backend
          go mod download

      - name: Run Unit Tests
        run: |
          cd backend
          go test ./... -v -coverprofile=coverage.out -covermode=atomic

      - name: Generate Coverage Report
        run: |
          cd backend
          go tool cover -html=coverage.out -o coverage.html

      - name: Check Coverage Threshold
        run: |
          cd backend
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Total coverage: $coverage%"
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "❌ Coverage $coverage% is below 80% threshold"
            exit 1
          fi
          echo "✅ Coverage $coverage% meets threshold"

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./backend/coverage.out

  backend-integration-test:
    name: Backend Integration Tests
    runs-on: ubuntu-latest
    needs: backend-unit-test

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: billing_note_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Integration Tests
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/billing_note_test?sslmode=disable
        run: |
          cd backend
          go test ./tests/integration/... -v

  frontend-unit-test:
    name: Frontend Unit Tests
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install Dependencies
        run: |
          cd frontend
          npm ci

      - name: Run Unit Tests
        run: |
          cd frontend
          npm run test -- --coverage

      - name: Check Coverage Threshold
        run: |
          cd frontend
          npm run test:coverage-check

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./frontend/coverage/coverage-final.json

  e2e-test:
    name: E2E Tests (Playwright)
    runs-on: ubuntu-latest
    needs: [backend-integration-test, frontend-unit-test]

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: billing_note_test
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '20'

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Backend Dependencies
        run: |
          cd backend
          go mod download

      - name: Build Backend
        run: |
          cd backend
          go build -o server cmd/server/main.go

      - name: Start Backend
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/billing_note_test?sslmode=disable
          JWT_SECRET: test_secret
        run: |
          cd backend
          ./server &
          sleep 5

      - name: Install Frontend Dependencies
        run: |
          cd frontend
          npm ci

      - name: Install Playwright Browsers
        run: |
          cd frontend
          npx playwright install --with-deps

      - name: Build Frontend
        run: |
          cd frontend
          npm run build

      - name: Start Frontend
        run: |
          cd frontend
          npm run preview &
          sleep 3

      - name: Run Playwright Tests
        run: |
          cd frontend
          npx playwright test

      - name: Upload Playwright Report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: frontend/playwright-report/
          retention-days: 30

      - name: Upload Test Videos
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: test-videos
          path: frontend/test-results/
          retention-days: 7

  test-summary:
    name: Test Summary
    runs-on: ubuntu-latest
    needs: [backend-unit-test, backend-integration-test, frontend-unit-test, e2e-test]
    if: always()

    steps:
      - name: Summary
        run: |
          echo "## 🧪 Test Results Summary" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "✅ All tests passed!" >> $GITHUB_STEP_SUMMARY
```

**本地測試指令（Makefile）：**

```makefile
# backend/Makefile
.PHONY: test test-unit test-integration test-coverage

test:
	go test ./... -v

test-unit:
	go test ./internal/... -v

test-integration:
	go test ./tests/integration/... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total

test-watch:
	# 需要安裝 https://github.com/cespare/reflex
	reflex -r '\.go$$' -s -- make test

clean:
	rm -f coverage.out coverage.html
```

```makefile
# frontend/Makefile
.PHONY: test test-unit test-e2e test-coverage

test:
	npm run test

test-unit:
	npm run test -- --run

test-e2e:
	npx playwright test

test-e2e-ui:
	npx playwright test --ui

test-coverage:
	npm run test -- --coverage

test-watch:
	npm run test -- --watch

clean:
	rm -rf coverage/ playwright-report/ test-results/
```

---

## Phase 劃分

### Phase 1 - 核心基礎 (MVP)

**目標：** 建立基本記帳功能，驗證核心流程

**功能清單：**
- [ ] 用戶註冊/登入（JWT）
- [ ] 手動輸入交易記錄
  - 日期、金額、類別（收入/支出）
  - 子類別（餐飲、交通、娛樂等）
  - 備註
- [ ] 交易列表檢視
  - 分頁
  - 基本篩選（日期範圍、類別）
- [ ] 基礎圖表
  - 月度收支圖（Bar Chart）
  - 類別分布圓餅圖（Pie Chart）
- [ ] 基礎設定頁面

**資料庫 Schema：**
- `users` 表
- `transactions` 表
- `categories` 表

**交付成果：**
- 可運作的 Web App
- 手動輸入並查看交易記錄
- 基本圖表呈現

**測試要求：**

**後端 Unit Tests：**
- [ ] `models/transaction_test.go` - Transaction 模型驗證
- [ ] `models/user_test.go` - User 模型驗證
- [ ] `models/category_test.go` - Category 模型驗證
- [ ] `services/transaction_service_test.go` - 交易業務邏輯
- [ ] `repository/transaction_repo_test.go` - 交易資料庫操作
- [ ] `repository/user_repo_test.go` - 用戶資料庫操作
- [ ] `handlers/auth_test.go` - 認證 handler
- [ ] `handlers/transaction_test.go` - 交易 handler
- [ ] `middleware/auth_test.go` - JWT 認證中介層

**後端 Integration Tests：**
- [ ] `tests/integration/auth_api_test.go` - 註冊/登入 API 測試
- [ ] `tests/integration/transaction_api_test.go` - 交易 CRUD API 測試
- [ ] `tests/integration/category_api_test.go` - 分類 API 測試
- [ ] `tests/integration/database_test.go` - 資料庫整合測試

**前端 Unit Tests：**
- [ ] 所有 `components/**/*.test.tsx` - 所有組件測試
- [ ] 所有 `hooks/*.test.ts` - 所有 hooks 測試
- [ ] 所有 `utils/*.test.ts` - 所有工具函式測試
- [ ] 所有 `api/*.test.ts` - API 客戶端測試

**E2E Tests (Playwright)：**
- [ ] `tests/e2e/auth.spec.ts` - 註冊/登入流程測試
- [ ] `tests/e2e/manual-transaction.spec.ts` - 手動新增交易測試
- [ ] `tests/e2e/transaction-list.spec.ts` - 交易列表、篩選、分頁測試
- [ ] `tests/e2e/charts.spec.ts` - 圖表顯示測試

**覆蓋率目標：**
- 後端：≥ 80%
- 前端：≥ 75%
- 所有測試必須 PASS 才能進入 Phase 2

---

### Phase 2A - PDF 自動匯入（手動上傳）

**目標：** 實現 PDF 信用卡帳單解析，減少手動輸入

#### 功能清單

**1. 密碼設定介面**
- [ ] 設定頁面新增「PDF 密碼管理」區塊
- [ ] 提供 4 組密碼輸入框
  - 密碼 #1（優先順序：1）
  - 密碼 #2（優先順序：2）
  - 密碼 #3（優先順序：3）
  - 密碼 #4（優先順序：4）
- [ ] 密碼加密儲存（AES-256）
- [ ] 儲存/更新密碼功能

**2. PDF 上傳流程**

**前端流程：**
```
1. 用戶點擊「上傳 PDF 帳單」
2. 選擇 PDF 檔案（支援多檔案）
3. 顯示檔案列表（檔名、大小）
4. 點擊「開始解析」
5. 顯示解析進度
6. 解析完成 → 顯示交易預覽表格
7. 用戶確認/編輯後點擊「匯入」
8. 匯入完成 → 跳轉交易列表
```

**後端流程：**
```
1. 接收 PDF 檔案（multipart/form-data）
2. 儲存到：/uploads/{user_id}/pdfs/{year}/{month}/{filename}
3. 讀取用戶設定的 4 組密碼
4. 依序嘗試解密 PDF
   - 成功：繼續步驟 5
   - 全部失敗：回傳錯誤「無法解密，請檢查密碼設定」
5. 解析 PDF 文字內容
6. 識別銀行類型（檔名關鍵字或內容）
7. 套用對應銀行解析器
8. 提取交易記錄（日期、金額、商家、類別）
9. 回傳 JSON 格式的交易預覽資料
```

**API：**
- `POST /api/upload/pdf` - 上傳並解析 PDF
- `POST /api/transactions/import` - 確認匯入交易

**3. PDF 解析器設計**

**支援銀行（初期）：**
根據 Bruce 提供的 PDF 範本決定，預留擴充性

**解析器架構：**
```go
// internal/pdf/parser.go

type BankParser interface {
    CanParse(content string) bool
    Parse(content string) ([]Transaction, error)
}

type ParserRegistry struct {
    parsers []BankParser
}

func (r *ParserRegistry) Parse(pdfPath string, passwords []string) ([]Transaction, error) {
    // 1. 嘗試解密
    reader := tryDecrypt(pdfPath, passwords)

    // 2. 提取文字
    content := extractText(reader)

    // 3. 找到適合的 parser
    for _, parser := range r.parsers {
        if parser.CanParse(content) {
            return parser.Parse(content)
        }
    }

    return nil, errors.New("unsupported bank format")
}
```

**4. 密碼管理**

**資料庫 Schema：**
```sql
CREATE TABLE user_pdf_passwords (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    password_encrypted TEXT NOT NULL,  -- AES-256 加密
    priority INT NOT NULL,              -- 嘗試順序 1-4
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**密碼嘗試邏輯：**
```go
func TryDecrypt(pdfPath string, passwords []string) (*pdf.Reader, error) {
    for i, pwd := range passwords {
        reader, err := pdf.NewReader(pdfPath, pwd)
        if err == nil {
            log.Printf("PDF decrypted with password #%d", i+1)
            return reader, nil
        }
    }
    return nil, errors.New("all passwords failed")
}
```

**交付成果：**
- 可上傳加密 PDF 並自動解析
- 支援多組密碼自動嘗試
- 交易預覽與確認機制

**測試要求：**

**後端 Unit Tests（新增）：**
- [ ] `pdf/parser_test.go` - PDF 解析核心邏輯
- [ ] `pdf/decryption_test.go` - **密碼嘗試機制（critical）**
- [ ] `pdf/bank_parsers/cathay_test.go` - 國泰解析器（根據 PDF 範本）
- [ ] `pdf/bank_parsers/esun_test.go` - 玉山解析器（根據 PDF 範本）
- [ ] `services/upload_service_test.go` - 上傳服務邏輯
- [ ] `services/pdf_password_service_test.go` - 密碼管理服務
- [ ] `handlers/upload_test.go` - 上傳 handler
- [ ] `handlers/pdf_password_test.go` - 密碼設定 handler

**後端 Integration Tests（新增）：**
- [ ] `tests/integration/upload_api_test.go` - PDF 上傳 API 完整流程
- [ ] `tests/integration/pdf_password_api_test.go` - 密碼設定/更新 API
- [ ] `tests/integration/pdf_import_test.go` - 解析後匯入測試

**測試資料準備：**
- [ ] 準備各銀行 PDF 範本（加密版 + 未加密版）
- [ ] 準備異常 PDF（損壞檔案、非銀行帳單、空白 PDF）
- [ ] 準備重複交易測試資料

**E2E Tests（新增）：**
- [ ] `tests/e2e/upload-pdf.spec.ts` - 完整上傳流程（選檔→解析→預覽→匯入）
- [ ] `tests/e2e/pdf-password-setup.spec.ts` - 密碼設定流程（設定4組密碼→測試解密）
- [ ] `tests/e2e/pdf-preview-import.spec.ts` - 預覽表格編輯與匯入測試
- [ ] `tests/e2e/pdf-error-handling.spec.ts` - 錯誤處理（密碼錯誤、檔案損壞）

**關鍵測試場景：**
```go
// 必須測試的邊界情況
- 密碼嘗試順序正確（1→2→3→4）
- 第3組密碼成功時，不嘗試第4組
- 所有密碼失敗時回傳明確錯誤
- 相同交易重複上傳檢測
- 多檔案同時上傳
- 大型 PDF（>5MB）處理
```

**覆蓋率目標：**
- `pdf/` 模組：≥ 85%（PDF 解析核心）
- 整體後端：≥ 80%
- 整體前端：≥ 75%
- 所有測試必須 PASS 才能進入 Phase 2B

---

### Phase 2B - Gmail 自動抓取

**目標：** 從 Gmail 自動下載信用卡帳單 PDF，完全自動化

#### Gmail API 整合

**1. Google Cloud 設定步驟**

**必要設定：**
```
1. 前往 Google Cloud Console (https://console.cloud.google.com)
2. 建立新專案「Billing-Note」
3. 啟用 API：
   - Gmail API
4. 建立 OAuth 2.0 憑證：
   - 應用程式類型：Web 應用程式
   - 名稱：Billing Note Web App
   - 授權的重新導向 URI：
     * http://localhost:3000/auth/google/callback （開發環境）
     * https://your-domain.com/auth/google/callback （正式環境）
5. 下載憑證 JSON 或複製：
   - Client ID
   - Client Secret
6. 將憑證寫入 backend/config/google_oauth.json
```

**所需權限 Scope：**
- `https://www.googleapis.com/auth/gmail.readonly` - 讀取郵件
- `https://www.googleapis.com/auth/gmail.metadata` - 讀取郵件 metadata

**2. OAuth 認證流程**

```
┌──────────┐                                   ┌──────────────┐
│  前端    │                                   │ Google OAuth │
└────┬─────┘                                   └──────┬───────┘
     │                                                │
     │ 1. 用戶點擊「連結 Gmail」                       │
     ├────────────────────────────────────────────────▶
     │ 2. 導向 Google 授權頁面                         │
     │    (含 client_id, scope, redirect_uri)         │
     │                                                │
     │ 3. 用戶授權                                    │
     │                                                │
     │ 4. Google 重新導向回系統                        │
     ◀────────────────────────────────────────────────┤
     │    (帶 authorization_code)                     │
     │                                                │
     │ 5. 前端將 code 傳給後端                         │
     ├──────────────▶ 後端                            │
     │                  │                             │
     │                  │ 6. 用 code 換 token         │
     │                  ├─────────────────────────────▶
     │                  │                             │
     │                  │ 7. 取得 access_token +      │
     │                  │    refresh_token            │
     │                  ◀─────────────────────────────┤
     │                  │                             │
     │                  │ 8. 加密儲存 refresh_token   │
     │                  │    到資料庫                  │
     │                  │                             │
     │ 9. 回傳成功      │                             │
     ◀────────────────  │                             │
     │                                                │
```

**3. Gmail 郵件掃描功能**

**掃描規則設定介面：**
```
設定頁面 > Gmail 整合

□ 啟用 Gmail 自動掃描

掃描規則：
- 寄件者包含關鍵字：
  [credit, 信用卡, 帳單, statement]
- 主旨包含關鍵字：
  [帳單, 電子帳單, statement]
- 必須有附件

[測試掃描] [儲存設定]

上次掃描時間：2026-01-07 10:30
[立即掃描]
```

**掃描邏輯：**
```go
// internal/services/gmail_service.go

func (s *GmailService) ScanForBills(userID int) ([]EmailBill, error) {
    // 1. 取得用戶的 refresh_token
    token := s.repo.GetGmailToken(userID)

    // 2. 建立 Gmail client
    client := s.createGmailClient(token)

    // 3. 搜尋郵件
    query := "has:attachment (from:credit OR subject:帳單)"
    messages := client.Users.Messages.List("me").Q(query).Do()

    // 4. 下載附件
    var bills []EmailBill
    for _, msg := range messages.Messages {
        attachments := s.getAttachments(client, msg.Id)
        for _, att := range attachments {
            if strings.HasSuffix(att.Filename, ".pdf") {
                // 下載 PDF
                data := s.downloadAttachment(client, msg.Id, att.AttachmentId)

                // 儲存檔案
                path := s.savePDF(userID, att.Filename, data)

                bills = append(bills, EmailBill{
                    Filename: att.Filename,
                    Path: path,
                    Date: msg.Date,
                })
            }
        }
    }

    return bills, nil
}
```

**4. 自動化排程**

**選項 A - 手動觸發：**
- 用戶點擊「掃描 Gmail」按鈕
- 即時掃描並顯示結果

**選項 B - 定期自動掃描（進階）：**
- 使用 GO cron job（github.com/robfig/cron）
- 每天固定時間掃描（例如：每天早上 8:00）
- 掃描後發送通知

**API：**
- `GET /api/gmail/auth` - 取得 OAuth 授權 URL
- `POST /api/gmail/callback` - 處理 OAuth callback
- `POST /api/gmail/scan` - 手動觸發掃描
- `GET /api/gmail/status` - 查詢連結狀態

**交付成果：**
- Gmail OAuth 連結功能
- 自動掃描並下載信用卡帳單 PDF
- 自動觸發 PDF 解析流程（複用 Phase 2A）

**測試要求：**

**後端 Unit Tests（新增）：**
- [ ] `services/gmail_service_test.go` - Gmail API 呼叫邏輯
- [ ] `services/gmail_scanner_test.go` - 郵件掃描與過濾邏輯
- [ ] `services/gmail_oauth_test.go` - OAuth 流程處理
- [ ] `handlers/gmail_test.go` - Gmail handler 測試

**後端 Integration Tests（新增）：**
- [ ] `tests/integration/gmail_oauth_test.go` - OAuth 完整流程（mock Google）
- [ ] `tests/integration/gmail_scan_test.go` - 掃描測試（mock Gmail API）
- [ ] `tests/integration/gmail_download_test.go` - 附件下載測試

**Mock 策略：**
- [ ] 建立 Gmail API mock server
- [ ] 準備 mock 郵件資料（含附件）
- [ ] 模擬 OAuth callback

**E2E Tests（新增）：**
- [ ] `tests/e2e/gmail-connect.spec.ts` - Gmail 連結流程（mock OAuth）
- [ ] `tests/e2e/gmail-scan.spec.ts` - 掃描並下載郵件測試
- [ ] `tests/e2e/gmail-disconnect.spec.ts` - 取消連結測試

**關鍵測試場景：**
```go
// 必須測試的情境
- OAuth token 過期自動更新
- 郵件搜尋規則正確性
- 附件過濾（只下載 PDF）
- 重複郵件不重複下載
- Gmail API 速率限制處理
```

**覆蓋率目標：**
- `services/gmail_*`：≥ 85%
- 整體後端：≥ 80%
- 整體前端：≥ 75%
- 所有測試必須 PASS 才能進入 Phase 3

---

### Phase 3 - 雲端發票整合

**目標：** 整合財政部電子發票 API，實現發票自動匯入與去重複

#### 財政部電子發票 API

**1. API 申請與設定**

**申請流程：**
```
1. 前往財政部電子發票整合服務平台
   https://www.einvoice.nat.gov.tw/
2. 註冊開發者帳號
3. 申請 API 使用權限
4. 取得 APP ID（API Key）
5. 設定 Callback URL（如需要）
```

**手機條碼載具：**
- 用戶需先在「統一發票兌獎 APP」或財政部網站申請手機條碼
- 格式：`/XXXXXXX`（7 碼）

**2. 發票資料拉取**

**API Endpoint：**
```
GET https://api.einvoice.nat.gov.tw/PB2CAPIVAN/invapp/InvApp

參數：
- version: 0.5
- action: carrierInvChk
- cardType: 3J0002 (手機條碼)
- cardNo: /XXXXXXX (用戶手機條碼)
- expTimeStamp: Unix timestamp
- timeStamp: Unix timestamp
- startDate: YYYY/MM/DD
- endDate: YYYY/MM/DD
- onlyWinningInv: N
- uuid: APP_ID
- appID: APP_ID
```

**回應範例：**
```json
{
  "v": "0.5",
  "code": 200,
  "msg": "成功",
  "details": [
    {
      "invNum": "AB12345678",
      "cardType": "3J0002",
      "cardNo": "/ABCD123",
      "sellerName": "全家便利商店",
      "invStatus": "已使用",
      "invDonatable": true,
      "amount": 150,
      "invPeriod": "11312",
      "donateMark": "0",
      "invDate": "2026/01/05 14:30:00",
      "sellerBan": "12345678",
      "sellerAddress": "台北市...",
      "invoiceTime": "14:30:00",
      "details": [
        {
          "description": "商品A",
          "quantity": "1",
          "unitPrice": "100",
          "amount": "100"
        }
      ]
    }
  ]
}
```

**3. 資料模型**

**資料庫 Schema：**
```sql
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    invoice_number VARCHAR(10) NOT NULL,     -- AB12345678
    invoice_date TIMESTAMP NOT NULL,
    seller_name VARCHAR(255),
    seller_ban VARCHAR(8),                   -- 統一編號
    amount DECIMAL(10, 2),
    status VARCHAR(50),                      -- 已使用/未使用
    items JSONB,                             -- 發票明細
    is_duplicated BOOLEAN DEFAULT FALSE,     -- 是否與信用卡重複
    duplicated_transaction_id INT,           -- 重複的交易 ID
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, invoice_number)
);
```

**4. 去重複邏輯**

**比對規則：**
```go
// internal/services/deduplication.go

type DeduplicationRule struct {
    AmountTolerance  float64  // 金額容差（元）
    TimeTolerance    int      // 時間容差（分鐘）
    MatchFields      []string // 比對欄位
}

func (s *DeduplicationService) FindDuplicates(invoice Invoice) (*Transaction, bool) {
    // 規則：
    // 1. 發票日期 ± 3 天內
    // 2. 金額完全相符（或 ±1 元容差）
    // 3. 商家名稱相似度 > 80%

    startDate := invoice.Date.AddDate(0, 0, -3)
    endDate := invoice.Date.AddDate(0, 0, 3)

    transactions := s.repo.FindTransactionsByDateRange(
        invoice.UserID,
        startDate,
        endDate,
    )

    for _, txn := range transactions {
        // 比對金額
        if math.Abs(txn.Amount - invoice.Amount) <= 1.0 {
            // 比對商家名稱（使用 Levenshtein distance）
            similarity := calculateSimilarity(txn.Description, invoice.SellerName)
            if similarity > 0.8 {
                return &txn, true  // 找到重複
            }
        }
    }

    return nil, false  // 無重複
}
```

**去重複策略：**
- **自動標記：** 系統自動標記可能重複的發票
- **用戶確認：** 前端顯示「可能重複」標籤，用戶可手動確認或取消
- **不匯入重複：** 預設不匯入已標記重複的發票

**5. 前端介面**

**雲端發票設定頁面：**
```
設定 > 雲端發票

□ 啟用雲端發票自動同步

手機條碼載具：
[/ABCD123] [驗證]

同步設定：
○ 手動同步
○ 每日自動同步（早上 9:00）

上次同步時間：2026-01-07 09:00
[立即同步]

去重複設定：
☑ 自動標記與信用卡重複的發票
  金額容差：[±1] 元
  時間容差：[±3] 天
```

**API：**
- `POST /api/invoice/sync` - 同步雲端發票
- `GET /api/invoice/list` - 發票列表
- `POST /api/invoice/confirm-duplicate` - 確認重複
- `DELETE /api/invoice/{id}` - 刪除發票

**交付成果：**
- 雲端發票自動同步
- 與信用卡交易去重複
- 重複項目人工確認機制

**測試要求：**

**後端 Unit Tests（新增）：**
- [ ] `services/invoice_service_test.go` - 發票同步邏輯
- [ ] `services/deduplication_test.go` - **去重複演算法（critical！）**
- [ ] `handlers/invoice_test.go` - 發票 handler 測試

**後端 Integration Tests（新增）：**
- [ ] `tests/integration/invoice_api_test.go` - 發票 API 測試（mock 財政部 API）
- [ ] `tests/integration/deduplication_integration_test.go` - **去重複整合測試**

**去重複測試場景（critical）：**
```go
// 必須測試的所有情境
TestDeduplication_ExactMatch()           // 完全相同
TestDeduplication_SameAmountDifferentDate()  // 金額同、日期差5天
TestDeduplication_SameDate_DifferentAmount() // 日期同、金額差2元
TestDeduplication_SimilarMerchant_80percent() // 商家名相似度80%
TestDeduplication_SimilarMerchant_79percent() // 商家名相似度79%（不應匹配）
TestDeduplication_MultipleMatches()      // 找到多個可能重複
TestDeduplication_EdgeCase_Timezone()    // 時區邊界情況
TestDeduplication_EdgeCase_DecimalPrecision() // 金額精度問題
```

**Mock 策略：**
- [ ] Mock 財政部發票 API
- [ ] 準備測試發票資料（含重複、不重複、邊界情況）

**E2E Tests（新增）：**
- [ ] `tests/e2e/invoice-sync.spec.ts` - 發票同步流程
- [ ] `tests/e2e/invoice-duplicate-handling.spec.ts` - 重複處理 UI 測試
- [ ] `tests/e2e/invoice-confirm-duplicate.spec.ts` - 確認/取消重複測試

**覆蓋率目標：**
- `services/deduplication.go`：≥ 90%（critical 邏輯）
- 整體後端：≥ 80%
- 整體前端：≥ 75%
- **去重複誤判率：< 5%**（需要實際資料驗證）
- 所有測試必須 PASS 才能進入 Phase 4

---

### Phase 4 - 進階功能

**目標：** 完善系統功能，提升使用體驗

**功能清單：**
- [ ] 電子存摺匯入（CSV/PDF）
- [ ] 進階圖表
  - 月度趨勢圖（Line Chart）
  - 類別年度對比（Grouped Bar Chart）
  - 收支比例（Stacked Bar Chart）
- [ ] 預算管理
  - 設定月度預算
  - 超支警告
- [ ] 資料匯出
  - CSV 格式
  - Excel 格式
  - 日期範圍選擇
- [ ] 進階搜尋/篩選
  - 關鍵字搜尋
  - 多條件篩選
  - 儲存常用篩選條件
- [ ] 標籤系統
  - 自訂標籤
  - 多標籤支援
- [ ] 備份/還原
  - 手動匯出所有資料
  - 匯入歷史資料

**交付成果：**
- 完整功能的記帳系統
- 優秀的使用體驗

**測試要求：**

**後端 Unit Tests（新增）：**
- [ ] `services/export_service_test.go` - 資料匯出邏輯
- [ ] `services/search_service_test.go` - 進階搜尋邏輯
- [ ] `services/budget_service_test.go` - 預算管理邏輯
- [ ] `handlers/export_test.go` - 匯出 handler
- [ ] `handlers/search_test.go` - 搜尋 handler

**後端 Integration Tests（新增）：**
- [ ] `tests/integration/export_api_test.go` - 匯出 API 測試
- [ ] `tests/integration/search_api_test.go` - 進階搜尋 API 測試
- [ ] `tests/integration/budget_api_test.go` - 預算 API 測試

**E2E Tests（新增）：**
- [ ] `tests/e2e/advanced-charts.spec.ts` - 進階圖表測試
- [ ] `tests/e2e/export-data.spec.ts` - 資料匯出測試
- [ ] `tests/e2e/advanced-search.spec.ts` - 進階搜尋與篩選測試
- [ ] `tests/e2e/budget-management.spec.ts` - 預算管理測試

**覆蓋率目標：**
- 整體後端：≥ 80%
- 整體前端：≥ 75%
- 所有測試必須 PASS 才能進入 Phase 5

---

### Phase 5 - 多用戶共享功能

**目標：** 實現多用戶互相檢視功能，支援家庭成員共享財務資訊

#### 使用場景

**典型使用流程：**
```
1. Bruce 註冊並使用系統（Phase 1-4 功能）
2. 配偶也註冊自己的帳號
3. Bruce 授權配偶「檢視」自己的交易資料
4. 配偶也授權 Bruce「檢視」
5. 雙方登入後可以切換檢視：
   - 「我的帳本」（自己的交易，可編輯）
   - 「配偶的帳本」（對方的交易，唯讀）
```

**關鍵特性：**
- 每個人的交易資料完全獨立（各自的 PDF、Gmail、發票）
- 授權檢視是雙向的（需要雙方各自授權）
- 檢視他人資料時為**唯讀模式**（不能新增/編輯/刪除）
- 可擴展到 2+ 人（家庭成員、室友等）

#### 功能清單

**1. 邀請與授權機制**

**選項 A - 配對碼（簡單，建議）：**
```
設定頁面：
我的配對碼：AB12-CD34
[重新生成]

加入其他人的帳本：
輸入對方的配對碼：[____-____]
[加入]

已連結的帳本：
- 配偶 (email@example.com) [移除授權]
```

**選項 B - Email 邀請（正式）：**
```
設定頁面：
邀請其他人檢視我的帳本：
Email: [___________]
[發送邀請]

待處理的邀請：
- spouse@example.com (已發送，等待接受)

我的邀請：
- bruce@example.com 邀請你檢視他的帳本 [接受] [拒絕]
```

**資料庫 Schema：**
```sql
CREATE TABLE shared_access (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(20) DEFAULT 'read',  -- read / write (Phase 5 只實作 read)
    status VARCHAR(20) DEFAULT 'active',    -- active / revoked
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(owner_id, shared_with)
);

-- 配對碼表（選項 A）
CREATE TABLE user_pairing_codes (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(10) UNIQUE NOT NULL,  -- AB12-CD34 格式
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**2. 帳本切換介面**

**前端 UI 設計：**
```
頂部導航欄：
┌────────────────────────────────────┐
│ 💰 Billing Note                    │
│                                    │
│ [下拉選單：我的帳本 ▼]              │
│   ├─ 我的帳本                       │
│   ├─ 配偶的帳本 👁️                 │
│   └─ 家人的帳本 👁️                 │
│                                    │
│ [Dashboard] [交易] [圖表] [設定]    │
└────────────────────────────────────┘

當檢視他人帳本時：
┌────────────────────────────────────┐
│ 📊 配偶的帳本 (唯讀模式)            │
│                                    │
│ 🔒 你正在檢視配偶的交易資料          │
│ 此模式下無法新增、編輯或刪除交易     │
└────────────────────────────────────┘
```

**3. 權限控制**

**API 層權限檢查：**
```go
// middleware/permission.go

func CheckViewPermission(c *gin.Context) {
    currentUserID := getUserIDFromToken(c)
    viewAsUserID := c.Query("view_as")  // 檢視誰的資料

    // 如果檢視自己的資料，直接通過
    if viewAsUserID == "" || viewAsUserID == currentUserID {
        c.Set("view_as_user_id", currentUserID)
        c.Next()
        return
    }

    // 檢查是否有授權
    hasAccess := checkSharedAccess(viewAsUserID, currentUserID)
    if !hasAccess {
        c.JSON(403, gin.H{"error": "無權限檢視此帳本"})
        c.Abort()
        return
    }

    c.Set("view_as_user_id", viewAsUserID)
    c.Set("read_only", true)  // 標記為唯讀模式
    c.Next()
}

func checkSharedAccess(ownerID, requesterID int) bool {
    var count int64
    db.Model(&SharedAccess{}).
        Where("owner_id = ? AND shared_with = ? AND status = 'active'",
              ownerID, requesterID).
        Count(&count)
    return count > 0
}
```

**前端唯讀模式：**
```tsx
// hooks/useReadOnlyMode.ts
const { data: currentView } = useQuery(['current-view']);

const isReadOnly = currentView.viewAsUserID !== currentView.currentUserID;

// 在交易列表頁面
{!isReadOnly && (
  <Button onClick={handleAddTransaction}>新增交易</Button>
)}

// 在交易詳情頁面
{!isReadOnly && (
  <>
    <Button onClick={handleEdit}>編輯</Button>
    <Button onClick={handleDelete}>刪除</Button>
  </>
)}
```

#### API 設計

**共享管理 API：**

**GET /api/shared/my-code**
取得我的配對碼
```json
{
  "code": "AB12-CD34",
  "expires_at": "2026-02-08T00:00:00Z"
}
```

**POST /api/shared/regenerate-code**
重新生成配對碼

**POST /api/shared/pair**
使用配對碼連結
```json
// Request
{
  "code": "XY98-ZW76"
}

// Response
{
  "success": true,
  "user": {
    "id": 2,
    "name": "配偶",
    "email": "spouse@example.com"
  }
}
```

**GET /api/shared/connections**
取得已連結的帳本
```json
{
  "connections": [
    {
      "user_id": 2,
      "name": "配偶",
      "email": "spouse@example.com",
      "permission": "read",
      "created_at": "2026-01-08T10:00:00Z"
    }
  ]
}
```

**DELETE /api/shared/connections/:user_id**
移除授權

**修改現有 API：**

所有交易相關 API 增加 `view_as` 參數：
```
GET /api/transactions?view_as=2
GET /api/charts/monthly-summary?view_as=2
GET /api/invoice/list?view_as=2
```

後端檢查：
- 如果 `view_as` 為空或等於當前用戶 → 回傳自己的資料
- 如果 `view_as` 不同 → 檢查權限 → 回傳對方的資料（唯讀）

#### 安全考量

**1. 權限驗證：**
- 所有 API 都必須通過權限中介層
- 檢查 `shared_access` 表的 `status = 'active'`
- 防止權限提升攻擊

**2. 資料隔離：**
- 所有資料庫查詢必須加上 `user_id` 過濾
- 防止跨用戶資料洩漏

**3. 唯讀模式強制：**
- 前端 UI 隱藏編輯按鈕
- **後端 API 必須再次驗證**（前端可被繞過）
- 修改/刪除 API 檢查 `view_as !== current_user` 時拒絕

**4. 授權撤銷：**
- 用戶可隨時撤銷授權
- 撤銷後對方立即無法檢視

#### 擴展性設計

**未來可擴展功能（Phase 6+）：**
- [ ] 寫入權限（`permission = 'write'`）
- [ ] 共同帳本模式（Workspace 概念）
- [ ] 細粒度權限（只分享特定類別、日期範圍）
- [ ] 群組管理（家庭群組、室友群組）
- [ ] 活動日誌（誰查看了我的帳本）

**資料表已預留：**
- `permission` 欄位（目前只用 `read`，未來可加 `write`）
- `status` 欄位（目前 `active/revoked`，未來可加 `pending`）

#### 交付成果

- [ ] 配對碼機制（或 Email 邀請）
- [ ] 帳本切換 UI
- [ ] 唯讀模式檢視他人交易
- [ ] 權限管理介面
- [ ] 前後端權限驗證
- [ ] 授權撤銷功能

#### 測試要求

**後端 Unit Tests（新增）：**
- [ ] `middleware/permission_test.go` - **權限檢查中介層（security critical）**
- [ ] `services/pairing_service_test.go` - 配對碼生成與驗證
- [ ] `services/shared_access_service_test.go` - 共享權限管理
- [ ] `handlers/shared_test.go` - 共享 handler 測試

**後端 Integration Tests（新增）：**
- [ ] `tests/integration/shared_access_api_test.go` - 授權 API 測試
- [ ] `tests/integration/permission_integration_test.go` - **權限驗證整合測試**
- [ ] `tests/integration/view_as_test.go` - `view_as` 參數測試

**安全測試（critical）：**
```go
// 必須測試的安全場景
TestPermission_PreventUnauthorizedAccess()     // 未授權存取防護
TestPermission_PreventSelfAuthorization()      // 防止自己授權給自己
TestPermission_PreventWriteInReadOnlyMode()    // 唯讀模式寫入防護
TestPermission_RevokedAccessDenied()           // 撤銷後立即拒絕存取
TestPermission_ExpiredTokenRejected()          // 過期 token 拒絕
TestPermission_SQLInjectionProtection()        // SQL 注入防護
TestPermission_PrivilegeEscalation()           // 權限提升攻擊防護
```

**E2E Tests（新增）：**
- [ ] `tests/e2e/pairing.spec.ts` - 完整配對流程（兩個瀏覽器模擬）
- [ ] `tests/e2e/view-others-account.spec.ts` - 檢視他人帳本測試
- [ ] `tests/e2e/read-only-mode.spec.ts` - **唯讀模式驗證（嘗試編輯應失敗）**
- [ ] `tests/e2e/revoke-access.spec.ts` - 撤銷授權測試
- [ ] `tests/e2e/multiple-shared-users.spec.ts` - 多人共享測試（3+ 用戶）

**覆蓋率目標：**
- `middleware/permission.go`：≥ 95%（security critical）
- `services/shared_*.go`：≥ 85%
- 整體後端：≥ 80%
- 整體前端：≥ 75%
- **安全測試通過率：100%**（不可妥協）
- 所有測試必須 PASS 才能正式上線

---

## 資料模型

### ER Diagram

```
                  ┌─────────────┐
              ┌───│   users     │───┐
              │   ├─────────────┤   │
              │   │ id          │◀──┼──────────────────┐
              │   │ email       │   │                  │
              │   │ password    │   │                  │
              │   │ name        │   │                  │
              │   │ created_at  │   │                  │
              │   └─────────────┘   │                  │
              │          │           │                  │
              │          │ user_id   │                  │
              │          │           │                  │
    ┌─────────┼──────────┼───────────┼────────────┐     │
    │         │          │           │            │     │
    ▼         ▼          ▼           ▼            ▼     │
┌─────────────────┐  ┌────────────┐ ┌──────────┐ ┌──────────────┐
│ shared_access   │  │transactions│ │invoices  │ │pdf_passwords │
│ (Phase 5)       │  ├────────────┤ ├──────────┤ ├──────────────┤
├─────────────────┤  │ id         │ │ id       │ │ id           │
│ id              │  │ user_id    │ │ user_id  │ │ user_id      │
│ owner_id        │──┘ date       │ │ inv_num  │ │ password_enc │
│ shared_with     │    amount     │ │ amount   │ │ priority     │
│ permission      │    type       │ │ is_dup   │ └──────────────┘
│ status          │    category_id├─┘dup_txn_id│
└─────────────────┘    description│  └──────────┘
         │             source     │
         │             tags       │       ┌──────────────────┐
         │             created_at │       │user_pairing_codes│
         │             └──────────┘       │ (Phase 5)        │
         │                    │           ├──────────────────┤
         │                    │           │ user_id          │
         │                    ▼           │ code             │
         │             ┌─────────────┐    │ expires_at       │
         └────────────▶│ categories  │    └──────────────────┘
                       ├─────────────┤
                       │ id          │
                       │ name        │
                       │ type        │
                       │ icon        │
                       └─────────────┘
```

**Phase 說明：**
- Phase 1-4：users, transactions, categories, invoices, pdf_passwords
- Phase 5：shared_access, user_pairing_codes（多用戶共享）

### 詳細 Schema

#### 1. users 表
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100),
    gmail_refresh_token TEXT,           -- Gmail refresh token (加密)
    invoice_carrier VARCHAR(10),        -- 手機條碼載具
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### 2. transactions 表
```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    type VARCHAR(10) NOT NULL,          -- income / expense
    category_id INT REFERENCES categories(id),
    description TEXT,
    source VARCHAR(20) NOT NULL,        -- manual / pdf / invoice / bank
    source_file VARCHAR(255),           -- PDF 檔案路徑
    tags TEXT[],                        -- 標籤陣列
    metadata JSONB,                     -- 額外資訊（彈性欄位）
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    INDEX idx_user_date (user_id, date),
    INDEX idx_category (category_id),
    INDEX idx_source (source)
);
```

#### 3. categories 表
```sql
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(10) NOT NULL,          -- income / expense
    icon VARCHAR(50),                   -- emoji 或 icon name
    parent_id INT REFERENCES categories(id),  -- 支援子分類
    user_id INT REFERENCES users(id),   -- NULL = 系統預設分類
    created_at TIMESTAMP DEFAULT NOW()
);

-- 預設分類
INSERT INTO categories (name, type, icon) VALUES
('薪資', 'income', '💰'),
('獎金', 'income', '🎁'),
('投資', 'income', '📈'),
('餐飲', 'expense', '🍜'),
('交通', 'expense', '🚗'),
('購物', 'expense', '🛍️'),
('娛樂', 'expense', '🎮'),
('醫療', 'expense', '🏥'),
('教育', 'expense', '📚'),
('房租', 'expense', '🏠'),
('水電', 'expense', '💡'),
('其他', 'expense', '📦');
```

#### 4. invoices 表
```sql
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invoice_number VARCHAR(10) NOT NULL,
    invoice_date TIMESTAMP NOT NULL,
    seller_name VARCHAR(255),
    seller_ban VARCHAR(8),
    amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50),
    items JSONB,                        -- 發票明細
    is_duplicated BOOLEAN DEFAULT FALSE,
    duplicated_transaction_id INT REFERENCES transactions(id),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, invoice_number),
    INDEX idx_user_date (user_id, invoice_date)
);
```

#### 5. pdf_passwords 表
```sql
CREATE TABLE pdf_passwords (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_encrypted TEXT NOT NULL,
    priority INT NOT NULL CHECK (priority BETWEEN 1 AND 4),
    filename_pattern VARCHAR(255),      -- 可選：檔名 pattern
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, priority)
);
```

#### 6. gmail_tokens 表
```sql
CREATE TABLE gmail_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    access_token_encrypted TEXT NOT NULL,
    refresh_token_encrypted TEXT NOT NULL,
    token_type VARCHAR(20) DEFAULT 'Bearer',
    expiry TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### 7. shared_access 表（Phase 5）
```sql
CREATE TABLE shared_access (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(20) DEFAULT 'read',  -- read / write
    status VARCHAR(20) DEFAULT 'active',    -- active / revoked
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(owner_id, shared_with),
    CHECK(owner_id != shared_with)  -- 防止自己授權給自己
);

-- 索引
CREATE INDEX idx_shared_access_owner ON shared_access(owner_id);
CREATE INDEX idx_shared_access_shared_with ON shared_access(shared_with);
```

#### 8. user_pairing_codes 表（Phase 5）
```sql
CREATE TABLE user_pairing_codes (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(10) UNIQUE NOT NULL,  -- 格式：AB12-CD34
    expires_at TIMESTAMP,              -- 可選：配對碼過期時間
    created_at TIMESTAMP DEFAULT NOW()
);

-- 自動生成配對碼的函數（PostgreSQL）
CREATE OR REPLACE FUNCTION generate_pairing_code() RETURNS VARCHAR(10) AS $$
DECLARE
    chars TEXT := 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';  -- 排除易混淆字元 I,O,0,1
    result VARCHAR(10) := '';
    i INTEGER;
BEGIN
    FOR i IN 1..4 LOOP
        result := result || substr(chars, floor(random() * length(chars) + 1)::int, 1);
    END LOOP;
    result := result || '-';
    FOR i IN 1..4 LOOP
        result := result || substr(chars, floor(random() * length(chars) + 1)::int, 1);
    END LOOP;
    RETURN result;
END;
$$ LANGUAGE plpgsql;
```

---

## API 設計

### 認證相關

#### POST /api/auth/register
註冊新用戶

**Request:**
```json
{
  "email": "bruce@example.com",
  "password": "secure_password",
  "name": "Bruce"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "bruce@example.com",
    "name": "Bruce"
  }
}
```

#### POST /api/auth/login
用戶登入

**Request:**
```json
{
  "email": "bruce@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "bruce@example.com",
    "name": "Bruce"
  }
}
```

---

### 交易相關

#### GET /api/transactions
取得交易列表

**Query Parameters:**
- `page` (int): 頁碼，預設 1
- `limit` (int): 每頁筆數，預設 20
- `start_date` (string): 開始日期 YYYY-MM-DD
- `end_date` (string): 結束日期 YYYY-MM-DD
- `type` (string): income / expense
- `category_id` (int): 分類 ID
- `source` (string): manual / pdf / invoice / bank

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "date": "2026-01-05",
      "amount": 150.00,
      "type": "expense",
      "category": {
        "id": 4,
        "name": "餐飲",
        "icon": "🍜"
      },
      "description": "全家便利商店",
      "source": "invoice",
      "tags": ["便利商店"]
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

#### POST /api/transactions
新增交易（手動輸入）

**Request:**
```json
{
  "date": "2026-01-08",
  "amount": 500.00,
  "type": "expense",
  "category_id": 4,
  "description": "午餐",
  "tags": ["餐飲", "聚餐"]
}
```

**Response:**
```json
{
  "id": 101,
  "date": "2026-01-08",
  "amount": 500.00,
  "type": "expense",
  "category_id": 4,
  "description": "午餐",
  "source": "manual",
  "created_at": "2026-01-08T12:00:00Z"
}
```

#### PUT /api/transactions/:id
更新交易

#### DELETE /api/transactions/:id
刪除交易

---

### PDF 上傳相關

#### POST /api/upload/pdf
上傳並解析 PDF

**Request:**
- Content-Type: `multipart/form-data`
- Field: `file` (PDF 檔案)

**Response:**
```json
{
  "filename": "玉山銀行_202601.pdf",
  "file_path": "/uploads/1/pdfs/2026/01/玉山銀行_202601.pdf",
  "parsed": true,
  "bank": "玉山銀行",
  "transactions": [
    {
      "date": "2026-01-05",
      "amount": 150.00,
      "description": "全家便利商店",
      "category_suggested": "餐飲"
    },
    {
      "date": "2026-01-06",
      "amount": 2500.00,
      "description": "遠東百貨",
      "category_suggested": "購物"
    }
  ]
}
```

**Error Response (密碼錯誤):**
```json
{
  "error": "failed_to_decrypt",
  "message": "無法解密 PDF，請檢查密碼設定"
}
```

#### POST /api/transactions/import
確認匯入交易（從 PDF 解析結果）

**Request:**
```json
{
  "source_file": "/uploads/1/pdfs/2026/01/玉山銀行_202601.pdf",
  "transactions": [
    {
      "date": "2026-01-05",
      "amount": 150.00,
      "type": "expense",
      "category_id": 4,
      "description": "全家便利商店"
    }
  ]
}
```

**Response:**
```json
{
  "imported": 2,
  "skipped": 0,
  "message": "成功匯入 2 筆交易"
}
```

---

### 密碼管理相關

#### GET /api/settings/pdf-passwords
取得 PDF 密碼設定

**Response:**
```json
{
  "passwords": [
    {
      "id": 1,
      "priority": 1,
      "has_password": true,
      "filename_pattern": null
    },
    {
      "id": 2,
      "priority": 2,
      "has_password": true,
      "filename_pattern": null
    },
    {
      "id": 3,
      "priority": 3,
      "has_password": false,
      "filename_pattern": null
    },
    {
      "id": 4,
      "priority": 4,
      "has_password": false,
      "filename_pattern": null
    }
  ]
}
```

#### POST /api/settings/pdf-passwords
設定/更新 PDF 密碼

**Request:**
```json
{
  "passwords": [
    {
      "priority": 1,
      "password": "password123"
    },
    {
      "priority": 2,
      "password": "mybirthday"
    },
    {
      "priority": 3,
      "password": "idcard_last6"
    },
    {
      "priority": 4,
      "password": "another_pwd"
    }
  ]
}
```

**Response:**
```json
{
  "message": "密碼設定已更新",
  "updated": 4
}
```

---

### Gmail 整合相關

#### GET /api/gmail/auth-url
取得 Google OAuth 授權 URL

**Response:**
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=...&redirect_uri=...&scope=..."
}
```

#### POST /api/gmail/callback
處理 OAuth callback

**Request:**
```json
{
  "code": "4/0AX4XfWh..."
}
```

**Response:**
```json
{
  "success": true,
  "message": "Gmail 已成功連結"
}
```

#### POST /api/gmail/scan
掃描 Gmail 信箱

**Response:**
```json
{
  "scanned": 15,
  "found_pdfs": 3,
  "pdfs": [
    {
      "filename": "國泰世華_202601.pdf",
      "date": "2026-01-05",
      "sender": "credit@cathaybk.com.tw",
      "file_path": "/uploads/1/pdfs/2026/01/國泰世華_202601.pdf"
    }
  ],
  "auto_parsed": 2,
  "failed": 1
}
```

#### GET /api/gmail/status
查詢 Gmail 連結狀態

**Response:**
```json
{
  "connected": true,
  "email": "bruce@gmail.com",
  "last_scan": "2026-01-07T09:00:00Z"
}
```

#### DELETE /api/gmail/disconnect
取消 Gmail 連結

---

### 雲端發票相關

#### POST /api/invoice/sync
同步雲端發票

**Request:**
```json
{
  "start_date": "2026/01/01",
  "end_date": "2026/01/31"
}
```

**Response:**
```json
{
  "synced": 25,
  "new": 20,
  "duplicated": 5,
  "invoices": [
    {
      "id": 1,
      "invoice_number": "AB12345678",
      "invoice_date": "2026-01-05T14:30:00Z",
      "seller_name": "全家便利商店",
      "amount": 150.00,
      "is_duplicated": true,
      "duplicated_transaction": {
        "id": 50,
        "description": "全家便利商店",
        "amount": 150.00
      }
    }
  ]
}
```

#### GET /api/invoice/list
發票列表

**Query Parameters:**
- `page`, `limit`, `start_date`, `end_date`
- `show_duplicated` (bool): 是否顯示重複項目

#### POST /api/invoice/:id/confirm-duplicate
確認/取消重複標記

**Request:**
```json
{
  "is_duplicated": true,
  "duplicated_transaction_id": 50
}
```

---

### 圖表相關

#### GET /api/charts/monthly-summary
月度收支摘要

**Query Parameters:**
- `year` (int): 年份
- `month` (int): 月份

**Response:**
```json
{
  "year": 2026,
  "month": 1,
  "income": 50000.00,
  "expense": 35000.00,
  "balance": 15000.00,
  "categories": [
    {
      "category": "餐飲",
      "amount": 8000.00,
      "percentage": 22.86
    },
    {
      "category": "交通",
      "amount": 5000.00,
      "percentage": 14.29
    }
  ]
}
```

#### GET /api/charts/trend
趨勢圖資料

**Query Parameters:**
- `start_date`, `end_date`
- `type` (string): income / expense / both

**Response:**
```json
{
  "data": [
    {
      "date": "2026-01",
      "income": 50000.00,
      "expense": 35000.00
    },
    {
      "date": "2026-02",
      "income": 52000.00,
      "expense": 38000.00
    }
  ]
}
```

---

## PDF 處理流程

### 完整流程圖

```
┌─────────────┐
│ 用戶上傳 PDF │
└──────┬──────┘
       │
       ▼
┌────────────────────┐
│ 前端：POST /upload │
│ multipart/form-data│
└──────┬─────────────┘
       │
       ▼
┌───────────────────────────┐
│ 後端：儲存檔案              │
│ /uploads/{user}/{year}/... │
└──────┬────────────────────┘
       │
       ▼
┌──────────────────────┐
│ 取得用戶的 4 組密碼   │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│ 輪流嘗試解密 PDF      │
│ Priority 1 → 2 → 3 → 4│
└──────┬───────────────┘
       │
       ├─ 成功 ──────────────┐
       │                     ▼
       │              ┌──────────────┐
       │              │ 提取 PDF 文字 │
       │              └──────┬───────┘
       │                     │
       │                     ▼
       │              ┌─────────────────┐
       │              │ 識別銀行類型     │
       │              │ (檔名或內容比對) │
       │              └──────┬──────────┘
       │                     │
       │                     ▼
       │              ┌──────────────────┐
       │              │ 套用銀行解析器    │
       │              │ - 國泰世華 Parser │
       │              │ - 玉山 Parser    │
       │              │ - 中信 Parser    │
       │              └──────┬───────────┘
       │                     │
       │                     ▼
       │              ┌──────────────────┐
       │              │ 提取交易資料      │
       │              │ - 日期           │
       │              │ - 金額           │
       │              │ - 商家           │
       │              │ - 類別（預測）   │
       │              └──────┬───────────┘
       │                     │
       │                     ▼
       │              ┌──────────────────┐
       │              │ 回傳 JSON        │
       │              │ (交易預覽)       │
       │              └──────┬───────────┘
       │                     │
       │                     ▼
       │              ┌──────────────────┐
       │              │ 前端：顯示預覽表格│
       │              │ 用戶可編輯/確認   │
       │              └──────┬───────────┘
       │                     │
       │                     ▼
       │              ┌──────────────────┐
       │              │ 用戶點擊「匯入」  │
       │              └──────┬───────────┘
       │                     │
       │                     ▼
       │              ┌──────────────────┐
       │              │ POST /import     │
       │              │ 寫入資料庫        │
       │              └──────┬───────────┘
       │                     │
       │                     ▼
       │              ┌──────────────────┐
       │              │ 完成！           │
       │              └──────────────────┘
       │
       └─ 失敗（全部密碼都試過）
              │
              ▼
       ┌──────────────────┐
       │ 回傳錯誤訊息      │
       │ "無法解密 PDF"    │
       └──────────────────┘
```

### PDF 檔案命名建議

為了更好的解析，建議 PDF 檔名包含銀行識別關鍵字：

**範例：**
- `國泰世華_202601.pdf`
- `玉山銀行_2026_01.pdf`
- `中國信託_帳單_202601.pdf`
- `台新銀行_statement_202601.pdf`

**檔名辨識邏輯：**
```go
func IdentifyBank(filename string) string {
    filename = strings.ToLower(filename)

    if strings.Contains(filename, "國泰") || strings.Contains(filename, "cathay") {
        return "cathay"
    }
    if strings.Contains(filename, "玉山") || strings.Contains(filename, "esun") {
        return "esun"
    }
    if strings.Contains(filename, "中信") || strings.Contains(filename, "chinatrust") {
        return "chinatrust"
    }

    return "unknown"
}
```

### 密碼設定最佳實踐

**建議：**
1. 密碼 #1：最常用的銀行密碼（例如：主力信用卡）
2. 密碼 #2：次常用的密碼
3. 密碼 #3：備用密碼
4. 密碼 #4：特殊銀行密碼

**安全性：**
- 所有密碼使用 AES-256-GCM 加密儲存
- 加密金鑰從環境變數讀取 `ENCRYPTION_KEY`
- 前端傳輸使用 HTTPS
- 密碼不會記錄在 log 中

---

## Gmail 整合流程

### 完整流程圖

```
┌────────────────┐
│ 用戶點擊        │
│ 「連結 Gmail」  │
└───────┬────────┘
        │
        ▼
┌────────────────────┐
│ GET /gmail/auth-url│
│ 取得 OAuth URL     │
└───────┬────────────┘
        │
        ▼
┌─────────────────────┐
│ 前端導向 Google      │
│ 授權頁面             │
└───────┬─────────────┘
        │
        ▼
┌─────────────────────┐
│ 用戶授權             │
│ (選擇 Google 帳號)   │
└───────┬─────────────┘
        │
        ▼
┌──────────────────────┐
│ Google 重新導向       │
│ /auth/google/callback│
│ ?code=XXX            │
└───────┬──────────────┘
        │
        ▼
┌─────────────────────┐
│ 前端：POST /callback │
│ { code: "XXX" }     │
└───────┬─────────────┘
        │
        ▼
┌──────────────────────┐
│ 後端：用 code 換 token│
│ Google OAuth API     │
└───────┬──────────────┘
        │
        ▼
┌─────────────────────────┐
│ 取得：                   │
│ - access_token          │
│ - refresh_token         │
│ - expiry                │
└───────┬─────────────────┘
        │
        ▼
┌─────────────────────────┐
│ 加密並儲存 tokens        │
│ 到 gmail_tokens 表       │
└───────┬─────────────────┘
        │
        ▼
┌─────────────────────┐
│ 完成！顯示成功訊息   │
└─────────────────────┘

--- 掃描流程 ---

┌────────────────┐
│ 用戶點擊        │
│ 「掃描 Gmail」  │
└───────┬────────┘
        │
        ▼
┌─────────────────────┐
│ POST /gmail/scan    │
└───────┬─────────────┘
        │
        ▼
┌─────────────────────┐
│ 取得 refresh_token  │
└───────┬─────────────┘
        │
        ▼
┌──────────────────────┐
│ 建立 Gmail API client│
└───────┬──────────────┘
        │
        ▼
┌───────────────────────────┐
│ 搜尋郵件                   │
│ query: "has:attachment     │
│  (from:credit OR           │
│   subject:帳單)"           │
└───────┬───────────────────┘
        │
        ▼
┌─────────────────────┐
│ 遍歷搜尋結果         │
└───────┬─────────────┘
        │
        ▼
┌──────────────────────┐
│ 檢查附件              │
│ 過濾 .pdf 檔案        │
└───────┬──────────────┘
        │
        ▼
┌──────────────────────┐
│ 下載 PDF 附件         │
│ Gmail API            │
└───────┬──────────────┘
        │
        ▼
┌────────────────────────┐
│ 儲存檔案                │
│ /uploads/{user}/...    │
└───────┬────────────────┘
        │
        ▼
┌────────────────────────┐
│ 自動觸發 PDF 解析       │
│ (複用 Phase 2A 邏輯)   │
└───────┬────────────────┘
        │
        ▼
┌────────────────────────┐
│ 回傳掃描結果            │
│ - 找到幾封               │
│ - 下載幾個 PDF          │
│ - 解析成功/失敗         │
└────────────────────────┘
```

### Gmail 搜尋規則範例

**基本搜尋：**
```
has:attachment filename:pdf
```

**進階搜尋（可在設定頁面讓用戶自訂）：**
```
has:attachment filename:pdf (
  from:credit@cathaybk.com.tw OR
  from:statement@esunbank.com.tw OR
  subject:信用卡帳單 OR
  subject:電子帳單
)
after:2026/01/01
```

**GO 程式碼範例：**
```go
func (s *GmailService) SearchEmails(userID int, query string) ([]*gmail.Message, error) {
    client := s.getClient(userID)

    req := client.Users.Messages.List("me").Q(query)
    res, err := req.Do()
    if err != nil {
        return nil, err
    }

    return res.Messages, nil
}
```

---

## 部署規劃

### 開發環境

**前端：**
```bash
cd frontend
npm install
npm run dev
# 運行在 http://localhost:5173
```

**後端：**
```bash
cd backend
go mod download
go run cmd/server/main.go
# 運行在 http://localhost:8080
```

**資料庫：**
```bash
docker run --name billing-postgres \
  -e POSTGRES_PASSWORD=dev_password \
  -e POSTGRES_DB=billing_note \
  -p 5432:5432 \
  -d postgres:15
```

### 正式環境

#### 選項 A - 分離部署（建議）

**前端 → Vercel：**
```yaml
# vercel.json
{
  "buildCommand": "npm run build",
  "outputDirectory": "dist",
  "framework": "vite",
  "env": {
    "VITE_API_URL": "https://api.billing-note.com"
  }
}
```

**後端 → Railway / Fly.io：**
```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config
EXPOSE 8080
CMD ["./server"]
```

**資料庫 → Railway PostgreSQL / Supabase**

#### 選項 B - 單一 VPS 部署

**使用 Docker Compose：**
```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: billing_note
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://postgres:${DB_PASSWORD}@postgres:5432/billing_note
      JWT_SECRET: ${JWT_SECRET}
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      GOOGLE_CLIENT_ID: ${GOOGLE_CLIENT_ID}
      GOOGLE_CLIENT_SECRET: ${GOOGLE_CLIENT_SECRET}
    depends_on:
      - postgres
    volumes:
      - ./uploads:/app/uploads

  frontend:
    build: ./frontend
    ports:
      - "80:80"
    environment:
      VITE_API_URL: http://backend:8080

volumes:
  postgres_data:
```

### 環境變數

**後端 (.env)：**
```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/billing_note

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-this

# Encryption (AES-256 需要 32 bytes)
ENCRYPTION_KEY=your-32-byte-encryption-key-here

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret

# 財政部發票 API
EINVOICE_APP_ID=your-app-id

# Server
PORT=8080
GIN_MODE=release

# File Upload
MAX_UPLOAD_SIZE=10485760  # 10MB
UPLOAD_DIR=/app/uploads
```

**前端 (.env)：**
```env
VITE_API_URL=http://localhost:8080
VITE_GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
```

### 備份策略

**資料庫自動備份：**
```bash
# Cron job - 每天凌晨 3:00 備份
0 3 * * * pg_dump billing_note | gzip > /backups/billing_note_$(date +\%Y\%m\%d).sql.gz
```

**檔案備份：**
```bash
# 備份 uploads 目錄到 S3 或其他雲端儲存
aws s3 sync /app/uploads s3://billing-note-backups/uploads
```

---

## 附錄

### A. PDF 範本分析（待補充）

**待 Bruce 提供 PDF 範本後，將補充：**
- 各銀行 PDF 格式分析
- 解析規則（正則表達式）
- 測試案例

### B. 技術債務追蹤

**Phase 1：**
- [ ] 密碼加密強度驗證
- [ ] API 速率限制
- [ ] 錯誤處理標準化

**Phase 2：**
- [ ] PDF 解析器測試覆蓋率 > 80%
- [ ] 大檔案上傳優化（分片上傳）
- [ ] Gmail API token 自動更新機制

**Phase 3：**
- [ ] 去重複演算法優化
- [ ] 發票資料快取機制

### C. 參考資源

**GO 相關：**
- Gin Framework: https://gin-gonic.com/
- GORM: https://gorm.io/
- UniPDF: https://github.com/unidoc/unipdf

**React 相關：**
- Vite: https://vitejs.dev/
- TanStack Query: https://tanstack.com/query
- Recharts: https://recharts.org/

**API 文件：**
- Gmail API: https://developers.google.com/gmail/api
- 財政部電子發票: https://www.einvoice.nat.gov.tw/

---

## 版本歷史

**v1.2 - 2026-01-08**
- 新增完整測試策略章節
- 定義測試工具與套件（GO: testify/gomock、React: Vitest/React Testing Library、E2E: Playwright）
- 為每個 Phase 加入詳細測試要求
- 設定測試覆蓋率目標（後端 ≥80%、前端 ≥75%）
- 新增 CI/CD 測試流程（GitHub Actions）
- 新增測試專案結構規劃
- 要求每個 function 都必須有 unit test

**v1.1 - 2026-01-08**
- 新增 Phase 5 - 多用戶共享功能
- 更新資料模型（新增 shared_access、user_pairing_codes 表）
- 新增共享相關 API 設計
- 更新用戶角色說明（Phase 1-4 單人，Phase 5 多人互相檢視）

**v1.0 - 2026-01-08**
- 初始版本
- 定義 Phase 1-4 功能範圍
- 完整技術架構設計
- PDF 處理流程
- Gmail 整合流程

---

**文件結束**

*此文件會隨著專案進展持續更新。*
