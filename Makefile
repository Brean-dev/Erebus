# Variables
BINARY_NAME=erebus
BUILD_DIR=./tmp
MAIN_PATH=./main.go

# Default target
.DEFAULT_GOAL := help

# Build the application
.PHONY: build
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built binary to $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Running..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Build and run in one command
.PHONY: dev
dev: run

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned!"

# Run tests
.PHONY: test
test:
	@echo "Testing..."
	@go test -v ./...
.PHONY: lint
lint:
	@golangci-lint run

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Help command
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build  - Build the application"
	@echo "  make run    - Build and run the application"
	@echo "  make dev    - Same as run"
	@echo "  make clean  - Remove build artifacts"
	@echo "  make test   - Run tests"
	@echo "  make deps   - Install dependencies"
	@echo "  make help   - Show this help message"
	@echo "  make lint   - Run golangci-lint"
