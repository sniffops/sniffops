# SniffOps Makefile

BINARY_NAME=sniffops
VERSION?=v0.1.0
BUILD_DIR=bin
GO=go
GOFLAGS=-ldflags="-X main.version=$(VERSION)"

# 기본 타겟
.DEFAULT_GOAL := build

.PHONY: all build build-web build-backend build-all run test clean install fmt lint help

## all: 전체 빌드 (의존성 정리 + 빌드)
all: deps build

## deps: Go 의존성 정리
deps:
	@echo ">> Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## build: 바이너리 빌드 (백엔드만)
build: build-backend

## build-web: 프론트엔드 빌드 (React/Vite)
build-web:
	@echo ">> Building frontend..."
	@if [ -f web/package.json ]; then \
		cd web && npm install && npm run build; \
		echo ">> Copying frontend assets to internal/web/dist..."; \
		rm -rf ../internal/web/dist && cp -r dist ../internal/web/dist; \
	else \
		echo ">> Frontend not initialized. Using placeholder..."; \
	fi

## build-backend: Go 백엔드 빌드
build-backend:
	@echo ">> Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/sniffops

## build-all: 프론트엔드 + 백엔드 전체 빌드
build-all: build-web build-backend
	@echo "✅ SniffOps built successfully (frontend + backend)"

## run: MCP 서버 실행 (개발용)
run: build
	@echo ">> Running $(BINARY_NAME) serve..."
	$(BUILD_DIR)/$(BINARY_NAME) serve

## web: 웹 UI 서버 실행 (개발용)
web: build
	@echo ">> Running $(BINARY_NAME) web..."
	$(BUILD_DIR)/$(BINARY_NAME) web

## test: 테스트 실행
test:
	@echo ">> Running tests..."
	$(GO) test -v -race -cover ./...

## clean: 빌드 아티팩트 삭제
clean:
	@echo ">> Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf web/dist
	rm -rf web/node_modules
	$(GO) clean

## install: 바이너리를 $GOPATH/bin 에 설치
install:
	@echo ">> Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) ./cmd/sniffops

## fmt: 코드 포맷팅
fmt:
	@echo ">> Formatting code..."
	$(GO) fmt ./...

## lint: 코드 린팅 (golangci-lint 필요)
lint:
	@echo ">> Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

## help: Makefile 도움말 출력
help:
	@echo "SniffOps Makefile Commands:"
	@echo ""
	@grep -E '^##' Makefile | sed 's/^## /  /'
