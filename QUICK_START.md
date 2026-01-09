# Billing Note - Quick Start Guide

快速開始指南 - Phase 1 完整版本

---

## 前置需求

1. **PostgreSQL** - 資料庫
2. **Go 1.21+** - 後端開發
3. **Node.js 18+** - 前端開發
4. **Make** - 建置工具 (可選)

---

## 快速啟動 (5 分鐘)

### 步驟 1: 建立資料庫

```bash
# 建立主資料庫
createdb billing_note

# 建立測試資料庫 (用於測試)
createdb billing_note_test

# 執行資料庫遷移
cd /Users/Bruce/git/Billing-Note/backend
psql -d billing_note -f migrations/001_init.sql
psql -d billing_note_test -f migrations/001_init.sql
```

### 步驟 2: 設定環境變數

```bash
# 後端環境變數
cd /Users/Bruce/git/Billing-Note/backend
cat > .env << EOF
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=billing_note
DB_SSLMODE=disable
JWT_SECRET=your-secret-key-change-this
JWT_EXPIRY=24h
PORT=8080
GIN_MODE=debug
EOF
```

### 步驟 3: 安裝依賴

```bash
# 後端依賴
cd /Users/Bruce/git/Billing-Note/backend
go mod download
go mod tidy

# 前端依賴
cd /Users/Bruce/git/Billing-Note/frontend
npm install
```

### 步驟 4: 啟動服務

開啟兩個終端視窗：

**終端 1 - 後端**:
```bash
cd /Users/Bruce/git/Billing-Note/backend
make run
# 或
go run cmd/server/main.go
```

**終端 2 - 前端**:
```bash
cd /Users/Bruce/git/Billing-Note/frontend
npm run dev
```

### 步驟 5: 訪問應用

打開瀏覽器，訪問：
- 前端: http://localhost:5173
- 後端 API: http://localhost:8080

---

## 測試指南

### 後端測試

```bash
cd /Users/Bruce/git/Billing-Note/backend

# 執行所有測試
go test ./... -v

# 執行測試並生成覆蓋率報告
go test ./... -v -coverprofile=coverage.out

# 查看覆蓋率摘要
go tool cover -func=coverage.out | grep total

# 查看覆蓋率詳細報告 (在瀏覽器中)
go tool cover -html=coverage.out

# 僅執行單元測試 (跳過需要資料庫的整合測試)
go test ./... -v -short

# 執行特定測試套件
go test ./internal/handlers/... -v
go test ./internal/middleware/... -v
go test ./tests/integration/... -v
```

### 前端測試

```bash
cd /Users/Bruce/git/Billing-Note/frontend

# 執行單元測試和組件測試
npm run test

# 執行測試並生成覆蓋率報告
npm run test -- --coverage

# 以監視模式執行測試
npm run test -- --watch

# 執行特定測試文件
npm run test TransactionList.test.tsx
```

### E2E 測試

```bash
cd /Users/Bruce/git/Billing-Note/frontend

# 確保後端和前端都在運行

# 執行 E2E 測試 (無界面)
npm run e2e

# 以有界面模式執行 (可以看到瀏覽器)
npm run e2e:headed

# 除錯模式
npm run e2e:debug

# 執行特定測試文件
npx playwright test tests/e2e/manual-transaction.spec.ts
```

---

## 常用命令

### 後端

```bash
# 啟動開發服務器
make run

# 建置
make build

# 執行測試
make test

# 清理
make clean

# 查看所有可用命令
make help
```

### 前端

```bash
# 開發服務器
npm run dev

# 建置生產版本
npm run build

# 預覽生產建置
npm run preview

# 執行測試
npm run test

# 執行 linting
npm run lint

# 執行 E2E 測試
npm run e2e
```

---

## 專案結構

```
Billing-Note/
├── backend/                    # Go 後端
│   ├── cmd/server/            # 主程式入口
│   ├── internal/              # 內部套件
│   │   ├── handlers/          # HTTP 處理器
│   │   ├── middleware/        # 中介層
│   │   ├── models/            # 資料模型
│   │   ├── repository/        # 資料庫存取層
│   │   └── services/          # 業務邏輯層
│   ├── pkg/                   # 公用套件
│   ├── migrations/            # 資料庫遷移
│   └── tests/                 # 整合測試
│
├── frontend/                   # React 前端
│   ├── src/
│   │   ├── api/              # API 客戶端
│   │   ├── components/       # React 組件
│   │   ├── hooks/            # 自訂 Hooks
│   │   ├── pages/            # 頁面組件
│   │   ├── store/            # 狀態管理
│   │   ├── types/            # TypeScript 類型
│   │   └── utils/            # 工具函數
│   └── tests/                # 測試文件
│       └── e2e/              # E2E 測試
│
├── TEST_REPORT.md            # 測試報告
├── DELIVERY_SUMMARY.md       # 交付摘要
└── QUICK_START.md            # 本文件
```

---

## 功能清單

### 已實現功能 ✅

1. **使用者認證**
   - 註冊
   - 登入
   - JWT 認證
   - 受保護的路由

2. **交易管理**
   - 新增交易 (收入/支出)
   - 查看交易列表
   - 編輯交易
   - 刪除交易
   - 篩選 (類型、日期、分類)
   - 分頁

3. **統計圖表**
   - 月度統計 (收入、支出、餘額)
   - 分類統計
   - 圓餅圖 (分佈)
   - 長條圖 (比較)
   - 年月篩選
   - 收入/支出切換

4. **設定**
   - 個人資料顯示
   - 應用程式設定 (預留)
   - 關於資訊
   - 登出功能

5. **其他**
   - 響應式設計
   - 載入狀態
   - 錯誤處理
   - 表單驗證
   - 空狀態處理

### 待開發功能 (Phase 2+)

1. 收據圖片上傳
2. 循環交易
3. 預算管理
4. 資料匯出 (CSV, PDF)
5. 個人資料編輯
6. 暗黑模式
7. 多幣別支援
8. 標籤系統
9. 搜尋功能
10. 通知提醒

---

## 故障排除

### 問題 1: 資料庫連線失敗

**症狀**: `connection refused` 或 `could not connect to database`

**解決方案**:
```bash
# 檢查 PostgreSQL 是否運行
pg_isready

# 檢查資料庫是否存在
psql -l | grep billing_note

# 重新建立資料庫
dropdb billing_note
createdb billing_note
psql -d billing_note -f backend/migrations/001_init.sql
```

### 問題 2: 後端測試失敗

**症狀**: Integration tests fail

**解決方案**:
```bash
# 確保測試資料庫存在
createdb billing_note_test
psql -d billing_note_test -f backend/migrations/001_init.sql

# 或跳過整合測試
go test ./... -v -short
```

### 問題 3: 前端無法連接後端

**症狀**: API calls fail with CORS error

**解決方案**:
- 確保後端正在運行在 port 8080
- 檢查後端的 CORS 設定
- 確認 API URL 設定正確

### 問題 4: E2E 測試失敗

**症狀**: Playwright tests timeout

**解決方案**:
```bash
# 確保前後端都在運行
# Terminal 1
cd backend && make run

# Terminal 2
cd frontend && npm run dev

# Terminal 3
cd frontend && npm run e2e

# 如果還是失敗，安裝 Playwright browsers
npx playwright install
```

### 問題 5: Go 模組問題

**症狀**: `cannot find package` 錯誤

**解決方案**:
```bash
cd backend
go mod download
go mod tidy
go mod verify
```

---

## API 端點

### 認證
- `POST /api/auth/register` - 註冊
- `POST /api/auth/login` - 登入
- `GET /api/auth/me` - 取得當前使用者 (需要認證)

### 交易
- `GET /api/transactions` - 取得交易列表 (需要認證)
- `GET /api/transactions/:id` - 取得單筆交易 (需要認證)
- `POST /api/transactions` - 新增交易 (需要認證)
- `PUT /api/transactions/:id` - 更新交易 (需要認證)
- `DELETE /api/transactions/:id` - 刪除交易 (需要認證)
- `GET /api/stats/monthly` - 取得月度統計 (需要認證)
- `GET /api/stats/category` - 取得分類統計 (需要認證)

### 分類
- `GET /api/categories` - 取得所有分類
- `GET /api/categories/type/:type` - 取得指定類型的分類

---

## 環境變數參考

### 後端 (.env)

```bash
# 資料庫設定
DB_HOST=localhost          # 資料庫主機
DB_PORT=5432              # 資料庫端口
DB_USER=postgres          # 資料庫使用者
DB_PASSWORD=postgres      # 資料庫密碼
DB_NAME=billing_note      # 資料庫名稱
DB_SSLMODE=disable        # SSL 模式

# JWT 設定
JWT_SECRET=your-secret-key-change-this    # JWT 密鑰
JWT_EXPIRY=24h                            # JWT 過期時間

# 伺服器設定
PORT=8080                 # 伺服器端口
GIN_MODE=debug           # Gin 模式 (debug/release)
ALLOWED_ORIGINS=http://localhost:5173    # CORS 允許的來源

# 上傳設定
UPLOAD_DIR=./uploads      # 上傳目錄
MAX_UPLOAD_SIZE=10485760  # 最大上傳大小 (10MB)
```

### 前端 (.env)

```bash
VITE_API_URL=http://localhost:8080    # 後端 API URL
```

---

## 開發提示

1. **Hot Reload**: 前後端都支援熱重載，修改代碼會自動重新載入
2. **測試驅動開發**: 建議先寫測試再實現功能
3. **代碼格式化**:
   - 後端: `gofmt -w .`
   - 前端: `npm run lint`
4. **Git 提交**: 建議使用有意義的提交訊息
5. **分支策略**: 建議使用 feature branches

---

## 效能優化建議

1. **資料庫索引**: 已在 migrations 中建立
2. **查詢優化**: 使用 pagination 和 filters
3. **前端優化**:
   - 使用 TanStack Query 的快取
   - 懶加載組件
   - 圖片優化
4. **API 優化**:
   - 使用 HTTP/2
   - 啟用 gzip 壓縮
   - 設定適當的快取標頭

---

## 安全性注意事項

1. **密碼**: 使用 bcrypt 加密
2. **JWT**: 定期更新密鑰
3. **CORS**: 只允許可信來源
4. **SQL 注入**: 使用 GORM 參數化查詢
5. **XSS**: React 預設防護
6. **CSRF**: 使用 JWT (stateless)

---

## 支援

如有問題，請參考：
1. TEST_REPORT.md - 測試詳細資訊
2. DELIVERY_SUMMARY.md - 專案交付摘要
3. PROJECT_STATUS.md - 專案狀態
4. README.md - 專案概述

---

**最後更新**: 2026-01-09
**版本**: 1.0.0 (Phase 1)
**狀態**: ✅ 開發完成，準備測試
