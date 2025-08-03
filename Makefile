# ============================================================================
# Marketplace Build System
# ============================================================================

.PHONY: all clean install dev build test coverage keys stripe id help \
	dev-backend dev-frontend build-backend build-frontend build-prod build-backend-prod \
	install-backend install-frontend test-verbose test-no-cache coverage-report coverage-func \
	generate-secret stripe-listen stripe-update stripe-confirm stripe-refund \
	id-generate id-generate-bulk id-decode run frontend backend lint fmt

# ============================================================================
# Variables
# ============================================================================

# Build Configuration
BINARY = bin/marketplace

# Directories
WEB_DIR = web
SRC_DIR = ./cmd/marketplace

# Security Files
PRIVATE_KEY_FILE = private.pem
PUBLIC_KEY_FILE = public.pem

# Clean Targets
CLEAN_FILES = $(BINARY) $(PRIVATE_KEY_FILE) $(PUBLIC_KEY_FILE) coverage.out *.log
CLEAN_DIRS = $(WEB_DIR)/dist $(WEB_DIR)/node_modules

# ============================================================================
# Main Targets
# ============================================================================

all: dev

help:
	@echo "Available targets:"
	@echo "  dev          - Start development environment"
	@echo "  build        - Build both backend and frontend"
	@echo "  test         - Run all tests"
	@echo "  clean        - Clean all generated files"
	@echo "  install      - Install dependencies"
	@echo "  keys         - Generate RSA keys"

# ============================================================================
# Development
# ============================================================================

dev: install
	@echo "Starting development environment..."
	@echo "Backend: http://localhost:8000"
	@echo "Frontend: http://localhost:5173"
	@make -j3 dev-backend stripe-listen dev-frontend

dev-dependencies:
	@echo "Starting database and dependencies..."
	docker compose -f deploy/local/docker-compose.yaml up -d

dev-all:
	@make dev
	@make dev-dependencies

dev-backend: build-backend
	./$(BINARY)

dev-frontend:
	cd $(WEB_DIR) && npm run dev

# ============================================================================
# Build
# ============================================================================

build: build-backend build-frontend

build-backend:
	@echo "Building backend..."
	@mkdir -p bin
	go build -o $(BINARY) $(SRC_DIR)

build-frontend:
	@echo "Building frontend..."
	cd $(WEB_DIR) && npm run build

# ============================================================================
# Dependencies
# ============================================================================

install: install-backend install-frontend

install-backend:
	@echo "Installing backend dependencies..."
	go mod download

install-frontend:
	@echo "Installing frontend dependencies..."
	cd $(WEB_DIR) && npm install

# ============================================================================
# Testing
# ============================================================================

test:
	go test ./...

test-verbose:
	go test -v ./...

test-no-cache:
	go test -count=1 ./...

coverage:
	go test -cover ./...

coverage-report:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-func:
	@if [ ! -f coverage.out ]; then \
		echo "No coverage.out file found. Run 'make coverage-report' first."; \
		exit 1; \
	fi
	go tool cover -func=coverage.out

# ============================================================================
# Code Quality
# ============================================================================

lint:
	@echo "Running linters..."
	golangci-lint run ./...
	cd $(WEB_DIR) && npm run lint

fmt:
	@echo "Formatting code..."
	go fmt ./...
	cd $(WEB_DIR) && npm run format

# ============================================================================
# Utilities
# ============================================================================

clean:
	@echo "Cleaning up..."
	rm -rf $(CLEAN_FILES) $(CLEAN_DIRS)

keys:
	@echo "Generating RSA key pair..."
	@openssl genpkey -algorithm RSA -out $(PRIVATE_KEY_FILE) -pkeyopt rsa_keygen_bits:2048
	@openssl rsa -pubout -in $(PRIVATE_KEY_FILE) -out $(PUBLIC_KEY_FILE)
	@echo "Generated: $(PRIVATE_KEY_FILE) and $(PUBLIC_KEY_FILE)"

generate-secret:
	@openssl rand -hex 32

# ============================================================================
# Stripe Tools
# ============================================================================

stripe-listen:
	go run ./cmd/stripe -cmd=listen

stripe-update:
	go run ./cmd/stripe -cmd=update -pi=$(PI)

stripe-confirm:
	go run ./cmd/stripe -cmd=confirm -pi=$(PI)

stripe-refund:
	go run ./cmd/stripe -cmd=refund -pi=$(PI)

# ============================================================================
# ID Tools
# ============================================================================

id-generate:
	go run ./cmd/id -cmd=generate-id

id-generate-bulk:
	go run ./cmd/id -cmd=generate-id -n=$(N)

id-decode:
	go run ./cmd/id -cmd=decode-id $(ID)

# ============================================================================
# Quick Run Commands
# ============================================================================

run: build-backend
	./$(BINARY)

frontend: install-frontend
	cd $(WEB_DIR) && npm run dev

backend: build-backend
	./$(BINARY)