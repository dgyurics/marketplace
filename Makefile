# Variables
BINARY_NAME=marketplace
SRC_DIR=./cmd/marketplace
PRIVATE_KEY_FILE=private.pem
PUBLIC_KEY_FILE=public.pem

# Default target
all: build

# Clean up generated files
clean:
	rm -f $(BINARY_NAME) $(PRIVATE_KEY_FILE) $(PUBLIC_KEY_FILE)

# Run tests
test:
	go test ./...

# View test coverage
test-coverage:
	go test -cover ./...

# Generate test coverage report
test-coverage-report:
	go test ./... -coverprofile=coverage.out

# View function-level test coverage
test-coverage-func:
	go tool cover -func=coverage.out

# Build the binary
build:
	go build -o bin/$(BINARY_NAME) $(SRC_DIR)

# Run the binary
run: build
	./bin/$(BINARY_NAME)

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
