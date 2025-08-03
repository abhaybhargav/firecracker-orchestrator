# Firecracker Orchestrator Makefile

# Variables
BINARY_NAME=orchestrator
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./cmd/orchestrator
BUILD_FLAGS=-ldflags="-s -w"
CGO_ENABLED=1

# Default target
.DEFAULT_GOAL := build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build completed: $(BINARY_PATH)"

# Build for Linux (useful when developing on macOS)
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p bin
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_PATH)-linux $(MAIN_PATH)
	@echo "Linux build completed: $(BINARY_PATH)-linux"

# Build for Linux without CGO (pure Go - easier deployment but slower database)
build-linux-static:
	@echo "Building $(BINARY_NAME) for Linux (static/pure Go)..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_PATH)-linux-static $(MAIN_PATH)
	@echo "Linux static build completed: $(BINARY_PATH)-linux-static"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "Clean completed"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run the application in development mode
dev: build
	@echo "Starting development server..."
	@LOG_LEVEL=debug ./$(BINARY_PATH)

# Run the application
run: build
	@echo "Starting orchestrator..."
	@./$(BINARY_PATH)

# Install dependencies and setup environment
setup:
	@echo "Setting up environment..."
	@./scripts/setup.sh

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Generate and serve documentation
docs:
	@echo "Generating documentation..."
	@godoc -http=:6060

# Create release build
release: clean
	@echo "Creating release build..."
	@mkdir -p bin
	@CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BINARY_PATH)-linux-amd64 $(MAIN_PATH)
	@echo "Release builds completed"

# Docker build (for container development)
docker-build:
	@echo "Building Docker image..."
	@docker build -t firecracker-orchestrator .

# Show help
help:
	@echo "Available targets:"
	@echo "  build             - Build the application"
	@echo "  build-linux       - Build for Linux (with CGO)"
	@echo "  build-linux-static - Build for Linux (pure Go, no CGO)"
	@echo "  clean             - Clean build artifacts"
	@echo "  deps              - Download and update dependencies"
	@echo "  test              - Run tests"
	@echo "  dev               - Run in development mode with debug logging"
	@echo "  run               - Run the application"
	@echo "  setup             - Setup environment (requires sudo)"
	@echo "  fmt               - Format code"
	@echo "  lint              - Lint code"
	@echo "  docs              - Generate and serve documentation"
	@echo "  release           - Create release builds"
	@echo "  docker-build      - Build Docker image"
	@echo "  help              - Show this help"

.PHONY: build build-linux build-linux-static clean deps test dev run setup fmt lint docs release docker-build help