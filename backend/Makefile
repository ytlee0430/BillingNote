.PHONY: help run build test test-coverage clean migrate-up migrate-down

help:
	@echo "Available commands:"
	@echo "  make run            - Run the server"
	@echo "  make build          - Build the binary"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deps           - Install dependencies"
	@echo "  make migrate-up     - Run database migrations"

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

deps:
	go mod download
	go mod tidy

migrate-up:
	psql -h localhost -U postgres -d billing_note -f migrations/001_init.sql
