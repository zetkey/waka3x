# Waka3x Makefile
# Common development tasks

.PHONY: help build test test-coverage lint run dev clean docker-build docker-run frontend-install frontend-build frontend-dev migrate fmt vet install-tools

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=waka3x
DOCKER_IMAGE=waka3x
DOCKER_TAG=latest
CONFIG_FILE=waka3x.yml
GO_FILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./node_modules/*")
FRONTEND_DIR=frontend

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

## help: Show this help message
help:
	@echo "$(COLOR_BOLD)Waka3x Development Commands$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_GREEN)Backend:$(COLOR_RESET)"
	@echo "  make build              - Build the Go binary"
	@echo "  make run                - Run the application"
	@echo "  make dev                - Run in development mode with hot reload (requires air)"
	@echo "  make test               - Run all tests"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo "  make lint               - Run linter (requires golangci-lint)"
	@echo "  make fmt                - Format Go code"
	@echo "  make vet                - Run go vet"
	@echo ""
	@echo "$(COLOR_GREEN)Frontend:$(COLOR_RESET)"
	@echo "  make frontend-install   - Install frontend dependencies"
	@echo "  make frontend-build     - Build frontend for production"
	@echo "  make frontend-dev       - Run frontend dev server"
	@echo ""
	@echo "$(COLOR_GREEN)Docker:$(COLOR_RESET)"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-run         - Run Docker container"
	@echo ""
	@echo "$(COLOR_GREEN)Database:$(COLOR_RESET)"
	@echo "  make migrate            - Run database migrations"
	@echo ""
	@echo "$(COLOR_GREEN)Utilities:$(COLOR_RESET)"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make install-tools      - Install development tools"
	@echo "  make all                - Build everything (backend + frontend)"
	@echo ""

## build: Build the Go binary
build:
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@go build -o $(BINARY_NAME) -v
	@echo "$(COLOR_GREEN)✓ Build complete: ./$(BINARY_NAME)$(COLOR_RESET)"

## run: Run the application
run:
	@echo "$(COLOR_BLUE)Running $(BINARY_NAME)...$(COLOR_RESET)"
	@go run main.go -config $(CONFIG_FILE)

## dev: Run in development mode with hot reload (requires air)
dev:
	@echo "$(COLOR_BLUE)Starting development server with hot reload...$(COLOR_RESET)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(COLOR_YELLOW)⚠ air not found. Install with: go install github.com/cosmtrek/air@latest$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)Falling back to regular run...$(COLOR_RESET)"; \
		$(MAKE) run; \
	fi

## test: Run all tests
test:
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@go test -v ./...
	@echo "$(COLOR_GREEN)✓ Tests complete$(COLOR_RESET)"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report generated: coverage.html$(COLOR_RESET)"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(COLOR_RESET)"; \
	fi

## fmt: Format Go code
fmt:
	@echo "$(COLOR_BLUE)Formatting Go code...$(COLOR_RESET)"
	@go fmt ./...
	@echo "$(COLOR_GREEN)✓ Formatting complete$(COLOR_RESET)"

## vet: Run go vet
vet:
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@go vet ./...
	@echo "$(COLOR_GREEN)✓ Vet complete$(COLOR_RESET)"

## frontend-install: Install frontend dependencies
frontend-install:
	@echo "$(COLOR_BLUE)Installing frontend dependencies...$(COLOR_RESET)"
	@cd $(FRONTEND_DIR) && bun install
	@echo "$(COLOR_GREEN)✓ Frontend dependencies installed$(COLOR_RESET)"

## frontend-build: Build frontend for production
frontend-build:
	@echo "$(COLOR_BLUE)Building frontend...$(COLOR_RESET)"
	@cd $(FRONTEND_DIR) && bun build
	@echo "$(COLOR_GREEN)✓ Frontend build complete: $(FRONTEND_DIR)/dist$(COLOR_RESET)"

## frontend-dev: Run frontend dev server
frontend-dev:
	@echo "$(COLOR_BLUE)Starting frontend dev server...$(COLOR_RESET)"
	@cd $(FRONTEND_DIR) && bun dev

## docker-build: Build Docker image
docker-build:
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(COLOR_GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(COLOR_RESET)"

## docker-run: Run Docker container
docker-run:
	@echo "$(COLOR_BLUE)Running Docker container...$(COLOR_RESET)"
	@docker run -d \
		-p 3000:3000 \
		-v waka3x-data:/data \
		-e WAKA3X_PASSWORD_SALT=$${WAKA3X_PASSWORD_SALT:-change-me} \
		--name $(BINARY_NAME) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "$(COLOR_GREEN)✓ Container started: $(BINARY_NAME)$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Access at: http://localhost:3000$(COLOR_RESET)"

## migrate: Run database migrations
migrate:
	@echo "$(COLOR_BLUE)Running database migrations...$(COLOR_RESET)"
	@go run main.go -config $(CONFIG_FILE)
	@echo "$(COLOR_GREEN)✓ Migrations complete$(COLOR_RESET)"

## clean: Clean build artifacts
clean:
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf $(FRONTEND_DIR)/dist
	@rm -f *.db *.db-journal *.db-shm *.db-wal
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

## install-tools: Install development tools
install-tools:
	@echo "$(COLOR_BLUE)Installing development tools...$(COLOR_RESET)"
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing air (hot reload)..."
	@go install github.com/cosmtrek/air@latest
	@echo "$(COLOR_GREEN)✓ Tools installed$(COLOR_RESET)"

## all: Build everything (backend + frontend)
all: frontend-build build
	@echo "$(COLOR_GREEN)✓ Full build complete$(COLOR_RESET)"

# Development workflow shortcuts
.PHONY: start stop restart

## start: Start both backend and frontend in development mode
start:
	@echo "$(COLOR_BLUE)Starting development environment...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Run 'make dev' in one terminal and 'make frontend-dev' in another$(COLOR_RESET)"

## Quick quality checks before commit
.PHONY: check
check: fmt vet lint test
	@echo "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)"
