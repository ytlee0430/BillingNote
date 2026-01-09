# Billing Note - 快速開始指南

## 第一次安裝（5-10 分鐘）

### 1. 環境檢查

確保已安裝以下工具：

```bash
# 檢查 GO 版本 (需要 1.21+)
go version

# 檢查 Node.js 版本 (需要 18+)
node --version

# 檢查 PostgreSQL 版本 (需要 15+)
psql --version

# 檢查 npm 版本
npm --version
```

### 2. 建立資料庫

```bash
# 建立資料庫
createdb billing_note

# 如果需要使用特定用戶
createdb -U postgres billing_note

# 執行 migrations
cd backend
psql -d billing_note -f migrations/001_init.sql

# 或使用特定用戶
psql -U postgres -d billing_note -f migrations/001_init.sql
```

### 3. 設置後端

```bash
cd backend

# 複製環境變數
cp .env.example .env

# 編輯 .env 檔案，設定資料庫連線
# 使用你喜歡的編輯器，例如：
nano .env
# 或
vim .env

# 安裝 GO 依賴
go mod download
go mod tidy

# 執行測試（確保一切正常）
go test ./... -v

# 啟動後端服務
go run cmd/server/main.go
```

後端將運行在 `http://localhost:8080`

### 4. 設置前端（開新 Terminal）

```bash
cd frontend

# 安裝依賴
npm install

# 執行測試（確保一切正常）
npm run test

# 啟動前端開發服務
npm run dev
```

前端將運行在 `http://localhost:5173`

### 5. 驗證安裝

1. 打開瀏覽器訪問：`http://localhost:5173`
2. 你應該看到註冊/登入頁面
3. 註冊一個新帳戶
4. 登入後應該看到 Dashboard

## 常見問題

### Q: 資料庫連線失敗

**問題**: `failed to connect to database`

**解決方案**:
```bash
# 1. 確保 PostgreSQL 服務正在運行
# macOS:
brew services start postgresql

# Linux:
sudo systemctl start postgresql

# 2. 檢查資料庫是否存在
psql -l | grep billing_note

# 3. 檢查 backend/.env 檔案的資料庫設定
```

### Q: GO 依賴安裝失敗

**解決方案**:
```bash
cd backend

# 清理並重新安裝
go clean -modcache
go mod download
go mod tidy
```

### Q: 前端依賴安裝失敗

**解決方案**:
```bash
cd frontend

# 清理並重新安裝
rm -rf node_modules package-lock.json
npm install

# 或使用 yarn
rm -rf node_modules yarn.lock
yarn install
```

### Q: Port 衝突

**後端 Port 8080 被佔用**:
編輯 `backend/.env`，修改 `PORT=8081`

**前端 Port 5173 被佔用**:
編輯 `frontend/vite.config.ts`，修改 `port: 5174`

### Q: CORS 錯誤

確保 `backend/.env` 的 `ALLOWED_ORIGINS` 包含前端 URL：
```
ALLOWED_ORIGINS=http://localhost:5173
```

## 開發流程

### 日常開發

**Terminal 1 - 後端**:
```bash
cd backend
make run
# 或
go run cmd/server/main.go
```

**Terminal 2 - 前端**:
```bash
cd frontend
npm run dev
```

### 執行測試

**後端測試**:
```bash
cd backend
make test                # 執行所有測試
make test-coverage       # 產生覆蓋率報告
open coverage.html       # 查看覆蓋率報告
```

**前端測試**:
```bash
cd frontend
npm run test            # 單元測試
npm run test:ui         # UI 模式
npm run test:coverage   # 覆蓋率報告
npm run e2e             # E2E 測試
```

### 建置生產版本

**後端**:
```bash
cd backend
make build
./bin/server
```

**前端**:
```bash
cd frontend
npm run build
npm run preview
```

## 重置資料庫

如果需要重置資料庫（會刪除所有資料）：

```bash
# 刪除資料庫
dropdb billing_note

# 重新建立
createdb billing_note

# 重新執行 migrations
cd backend
psql -d billing_note -f migrations/001_init.sql
```

## 檢查服務狀態

### 檢查後端

```bash
curl http://localhost:8080/health
# 應該返回: {"status":"ok"}
```

### 檢查前端

訪問 `http://localhost:5173`，應該看到登入頁面

## 下一步

1. 閱讀 [README.md](README.md) 了解更多功能
2. 查看 [spec.md](spec.md) 了解技術規格
3. 查看 [PROJECT_STATUS.md](PROJECT_STATUS.md) 了解開發進度
4. 開始開發新功能！

## 需要幫助？

- 查看 API 文件：`http://localhost:8080/api/*`
- 查看測試檔案了解使用範例
- 檢查 Console 和 Network 面板的錯誤訊息
