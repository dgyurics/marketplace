# Variables
BINARY_NAME=marketplace
SRC_DIR=./
ENV_FILE=.env
PRIVATE_KEY_FILE=private.pem
PUBLIC_KEY_FILE=public.pem

# Include environment variables from .env
ifneq (,$(wildcard $(ENV_FILE)))
	include $(ENV_FILE)
  export
endif

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
	env $(shell cat $(ENV_FILE) | xargs) ./$(BINARY_NAME)

# stripe listen is a command provided by the Stripe CLI that listens for
# events (webhooks) from Stripe in real-time. It allows you to test and debug
# webhook integrations locally without needing to deploy your server to the internet
stripe-listen:
	@echo "Reading Stripe Secret Key from .env..."
	@if ! grep -q "^STRIPE_SECRET_KEY=" $(ENV_FILE); then \
		echo "Error: STRIPE_SECRET_KEY not found in .env file"; \
		exit 1; \
	fi
	@SK=$$(grep "^STRIPE_SECRET_KEY=" $(ENV_FILE) | cut -d '=' -f2); \
	if [ -z "$$SK" ]; then \
		echo "Error: STRIPE_SECRET_KEY is empty"; \
		exit 1; \
	fi; \
	echo "Running Stripe listen with API Key $$SK..."; \
	stripe listen --api-key $$SK --forward-to http://localhost:8000/orders/events

# make update-payment-intent PI=pi_xxxx
update-payment-intent:
	@echo "Updating payment intent $(PI)..."
	@if [ -z "$(PI)" ]; then \
		echo "Error: PI (Payment Intent) is not set"; \
		exit 1; \
	fi
	stripe payment_intents update $(PI) --payment-method pm_card_visa


# make confirm-payment-intent PI=pi_xxxx
confirm-payment-intent:
	@echo "Confirming payment intent $(PI)..."
	@if [ -z "$(PI)" ]; then \
		echo "Error: PI (Payment Intent) is not set"; \
		exit 1; \
	fi
	stripe payment_intents confirm $(PI)

# make refund-payment-intent PI=pi_xxxx
refund-payment-intent:
	@echo "Refunding payment intent $(PI)..."
	@if [ -z "$(PI)" ]; then \
		echo "Error: PI (Payment Intent) is not set"; \
		exit 1; \
	fi
	stripe refunds create --payment_intent $(PI)
