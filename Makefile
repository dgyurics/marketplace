# Variables
BINARY_NAME=marketplace
SRC_DIR=./
ENV_FILE=.env

include $(ENV_FILE)

# Default target
all: run

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
