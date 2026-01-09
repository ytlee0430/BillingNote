# 建置階段
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 複製 go mod 檔案
COPY go.mod ./
RUN go mod download

# 複製所有源碼
COPY . .

# 移除可能損壞的 go.sum 並重新整理依賴
RUN rm -f go.sum && go mod tidy

# 建置應用程式
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# 運行階段
FROM alpine:latest

WORKDIR /app

# 安裝必要的工具
RUN apk --no-cache add ca-certificates tzdata

# 從建置階段複製編譯好的二進制檔案
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

# 建立 uploads 目錄
RUN mkdir -p /app/uploads

# 暴露端口
EXPOSE 8080

# 運行應用程式
CMD ["./server"]
