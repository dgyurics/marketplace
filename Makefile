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
	openssl genpkey -algorithm RSA -out $(PRIVATE_KEY_FILE) -pkeyopt rsa_keygen_bits:2048
	openssl rsa -pubout -in $(PRIVATE_KEY_FILE) -out $(PUBLIC_KEY_FILE)
	@echo "Private key: $(PRIVATE_KEY_FILE)"
	@echo "Public key: $(PUBLIC_KEY_FILE)"

# Clean up generated files
clean:
	rm -f $(BINARY_NAME)
	rm -f $(PRIVATE_KEY_FILE) $(PUBLIC_KEY_FILE)

# Run tests
test:
	go test $(SRC_DIR)...

# Build the binary
build:
	go build -o $(BINARY_NAME) $(SRC_DIR)

# Clean up generated files
clean:
	rm -f $(BINARY_NAME)

# Run the binary
run: build
	DATABASE_URL=$(DATABASE_URL) ./$(BINARY_NAME)
