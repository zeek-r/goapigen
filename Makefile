.PHONY: build clean test test-cov lint vet fmt install run-example all help

# Variables
BINARY_NAME=goapigen
BUILD_DIR=bin
COVERAGE_FILE=coverage.out
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")
EXAMPLE_SPEC=examples/petstore/openapi.yaml
EXAMPLE_OUT=examples/petstore/gen
EXAMPLE_PKG=petstore

# Default goal when 'make' is executed without arguments
.DEFAULT_GOAL := build

# Make all - builds everything
all: clean test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/goapigen

# Clean up the build directory and test coverage files
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE)
	@rm -f $(BINARY_NAME)

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-cov:
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE)

# Run linting
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Format all Go files
fmt:
	@echo "Formatting Go files..."
	@gofmt -w $(GO_FILES)

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd/goapigen

# Run the example
run-example: build
	@echo "Running example on $(EXAMPLE_SPEC)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -spec $(EXAMPLE_SPEC) -output $(EXAMPLE_OUT) -package $(EXAMPLE_PKG) -types -mongo

# Help command
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  clean        - Clean up build artifacts"
	@echo "  test         - Run all tests"
	@echo "  test-cov     - Run tests with coverage"
	@echo "  lint         - Run linter"
	@echo "  vet          - Run go vet"
	@echo "  fmt          - Format Go files"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  run-example  - Run the example"
	@echo "  all          - Clean, test and build"
	@echo "  help         - Print this help message"