# Variables
BINARY_NAME=marketplace
SRC_DIR=./
ENV_FILE=.env
PRIVATE_KEY_FILE=private.pem
PUBLIC_KEY_FILE=public.pem

include $(ENV_FILE)

# Default target
all: run

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

# Clean up generated files
clean:
	rm -f $(BINARY_NAME) $(PRIVATE_KEY_FILE) $(PUBLIC_KEY_FILE)

# Run tests
test:
	go test ./...

# Build the binary
build:
	go build -o $(BINARY_NAME) $(SRC_DIR)

# Run the binary
run: build
	DATABASE_URL=$(DATABASE_URL) HMAC_SECRET=$(HMAC_SECRET) ./$(BINARY_NAME)
