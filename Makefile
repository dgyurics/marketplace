.PHONY: all clean test test-no-cache test-coverage test-coverage-report test-coverage-func build run stripe-listen update-payment-intent confirm-payment-intent refund-payment-intent generate-id generate-ids decode-id generate-keys

# Variables
BIN_DIR=bin
BINARY_NAME=marketplace
BINARY=$(BIN_DIR)/$(BINARY_NAME)
SRC_DIR=./cmd/marketplace
PRIVATE_KEY_FILE=private.pem
PUBLIC_KEY_FILE=public.pem

# Build variables - can be overridden
GOOS ?= linux
GOARCH ?= amd64

# Default target
all: run

# Clean up generated files
clean:
	rm -f $(BINARY) $(PRIVATE_KEY_FILE) $(PUBLIC_KEY_FILE)

# Run tests
test:
	go test ./...

test-no-cache:
	go test -count=1 ./...

# View test coverage
test-coverage:
	go test -cover ./...

# Generate test coverage report
test-coverage-report:
	go test -coverprofile=coverage.out ./...

# View function-level test coverage
test-coverage-func:
	go tool cover -func=coverage.out

# Build the binary (development)
build:
	go build -o $(BINARY) $(SRC_DIR)

# Build the binary for production
build-prod:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINARY) $(SRC_DIR)

# Run the binary
run: build
	./$(BINARY)

# Stripe listen
stripe-listen:
	go run ./cmd/stripe -cmd=listen

# Payment Intent Commands
update-payment-intent:
	go run ./cmd/stripe -cmd=update -pi=$(PI)

confirm-payment-intent:
	go run ./cmd/stripe -cmd=confirm -pi=$(PI)

refund-payment-intent:
	go run ./cmd/stripe -cmd=refund -pi=$(PI)

# Generate unique IDs
generate-id:
	go run ./cmd/id -cmd=generate-id

generate-ids:
	go run ./cmd/id -cmd=generate-id -n=$(n)

decode-id:
	go run ./cmd/id -cmd=decode-id $(id)

# Generate RSA keys (private and public PEM files)
generate-keys:
	@echo "Generating RSA private and public keys..."
	@if ! openssl genpkey -algorithm RSA -out $(PRIVATE_KEY_FILE) -pkeyopt rsa_keygen_bits:2048; then \
		echo "Failed to generate private key"; \
		exit 1; \
	fi
	@if ! openssl rsa -pubout -in $(PRIVATE_KEY_FILE) -out $(PUBLIC_KEY_FILE); then \
		echo "Failed to generate public key"; \
		exit 1; \
	fi
	@echo "Private key: $(PRIVATE_KEY_FILE)"
	@echo "Public key: $(PUBLIC_KEY_FILE)"

# Generate a random 32-byte hex string
generate-rand:
	openssl rand -hex 32