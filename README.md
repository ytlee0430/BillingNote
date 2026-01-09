# Billing Note - 記帳系統

個人使用的 Web 記帳系統，支援自動化匯入信用卡帳單、雲端發票等資料來源。

## 專案架構

```
Billing-Note/
├── backend/          # GO + Gin + PostgreSQL 後端
├── frontend/         # React + TypeScript + Vite 前端
└── spec.md          # 完整技術規格文件
```

## 技術棧

### 後端
- **語言**: GO 1.21+
- **框架**: Gin
- **資料庫**: PostgreSQL 15+
- **ORM**: GORM
- **認證**: JWT (golang-jwt/jwt)
- **測試**: testify, go-sqlmock

### 前端
- **框架**: React 18 + TypeScript 5
- **建置工具**: Vite 5
- **路由**: React Router v6
- **狀態管理**: Zustand
- **資料獲取**: TanStack Query (React Query)
- **樣式**: Tailwind CSS 3
- **圖表**: Recharts
- **測試**: Vitest, React Testing Library, Playwright

## Phase 1 功能 (MVP)

- ✅ 用戶註冊/登入（JWT）
- ✅ 手動輸入交易記錄
- ✅ 交易列表檢視（分頁、篩選）
- ✅ 基礎圖表（月度收支、類別分布）
- ✅ 基礎設定頁面

## 快速開始

### 環境需求

- GO 1.21+
- Node.js 18+
- PostgreSQL 15+
- npm 或 yarn

### 資料庫設置

1. 安裝並啟動 PostgreSQL
2. 建立資料庫：

```bash
createdb billing_note
```

3. 執行 migrations：

```bash
cd backend
psql -d billing_note -f migrations/001_init.sql
```

### 後端設置

1. 複製環境變數範例：

```bash
cd backend
cp .env.example .env
```

2. 編輯 `.env` 檔案，設定資料庫連線資訊

3. 安裝依賴：

```bash
go mod download
go mod tidy
```

4. 啟動後端：

```bash
make run
# 或
go run cmd/server/main.go
```

後端將運行在 `http://localhost:8080`

### 前端設置

1. 安裝依賴：

```bash
cd frontend
npm install
```

2. 啟動開發伺服器：

```bash
npm run dev
```

前端將運行在 `http://localhost:5173`

## 測試

### 後端測試

```bash
cd backend

# 執行所有測試
make test

# 執行測試並產生覆蓋率報告
make test-coverage

# 查看覆蓋率報告
open coverage.html
```

**目標覆蓋率**: ≥ 80%

### 前端測試

```bash
cd frontend

# 單元測試
npm run test

# UI 模式
npm run test:ui

# 覆蓋率報告
npm run test:coverage

# E2E 測試
npm run e2e

# E2E Debug 模式
npm run e2e:debug
```

**目標覆蓋率**: ≥ 75%

## API 文件

### 認證

**POST /api/auth/register**
註冊新用戶

**POST /api/auth/login**
用戶登入

**GET /api/auth/me** (需要 JWT)
取得當前用戶資訊

### 交易記錄

**GET /api/transactions** (需要 JWT)
取得交易列表（支援分頁和篩選）

**POST /api/transactions** (需要 JWT)
建立新交易

**GET /api/transactions/:id** (需要 JWT)
取得單一交易

**PUT /api/transactions/:id** (需要 JWT)
更新交易

**DELETE /api/transactions/:id** (需要 JWT)
刪除交易

### 統計資料

**GET /api/stats/monthly** (需要 JWT)
取得月度統計

**GET /api/stats/category** (需要 JWT)
取得分類統計

### 分類

**GET /api/categories** (需要 JWT)
取得所有分類

**GET /api/categories/type/:type** (需要 JWT)
依類型取得分類（income/expense）

## 開發指令

### 後端

```bash
make run            # 啟動伺服器
make build          # 建置執行檔
make test           # 執行測試
make test-coverage  # 測試覆蓋率
make clean          # 清理建置檔案
make deps           # 安裝依賴
```

### 前端

```bash
npm run dev         # 開發模式
npm run build       # 建置生產版本
npm run preview     # 預覽生產版本
npm run test        # 單元測試
npm run e2e         # E2E 測試
npm run lint        # Lint 檢查
```

## 專案結構

### 後端結構

```
backend/
├── cmd/server/          # 程式入口
├── internal/
│   ├── handlers/        # HTTP 處理器
│   ├── services/        # 業務邏輯
│   ├── models/          # 資料模型
│   ├── repository/      # 資料庫操作
│   └── middleware/      # 中介層
├── pkg/                 # 可重用套件
│   ├── config/         # 配置管理
│   ├── database/       # 資料庫連線
│   └── utils/          # 工具函數
├── migrations/         # 資料庫遷移
└── tests/              # 測試檔案
```

### 前端結構

```
frontend/
├── src/
│   ├── components/     # React 組件
│   │   ├── common/    # 通用組件
│   │   └── ...
│   ├── pages/         # 頁面組件
│   ├── hooks/         # Custom Hooks
│   ├── api/           # API 客戶端
│   ├── types/         # TypeScript 型別
│   ├── utils/         # 工具函數
│   ├── store/         # Zustand Store
│   └── tests/         # 測試檔案
└── tests/e2e/         # Playwright E2E 測試
```

## 待辦事項

- [ ] 完成所有 Phase 1 測試（覆蓋率達標）
- [ ] 新增交易記錄管理頁面
- [ ] 新增圖表視覺化頁面
- [ ] 新增設定頁面
- [ ] Phase 2: PDF 信用卡帳單自動匯入
- [ ] Phase 3: 雲端發票整合
- [ ] Phase 4: Gmail 自動掃描
- [ ] Phase 5: 多用戶支援

## 授權

MIT License
