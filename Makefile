# Makefile for the KloneKit project

# Go parameters
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_CLEAN=$(GO_CMD) clean
GO_LINT=$(shell go env GOPATH)/bin/golangci-lint

# Binary name
BINARY_NAME=klonekit
BINARY_UNIX=$(BINARY_NAME)

.PHONY: all build test lint clean run help deps

all: build

# Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO_BUILD) -o $(BINARY_UNIX) ./cmd/klonekit

# Run all tests
test:
	@echo "Running tests..."
	$(GO_TEST) -v ./...

# Lint the codebase
# Assumes golangci-lint is installed: https://golangci-lint.run/usage/install/
lint:
	@echo "Linting codebase..."
	$(GO_LINT) run

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_UNIX)

# Install development dependencies
deps:
	@echo "Installing development dependencies..."
	@if ! command -v $(GO_LINT) >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	else \
		echo "golangci-lint is already installed"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GO_CLEAN)
	rm -f $(BINARY_UNIX)

# Display help
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  all        Build the application binary (default)"
	@echo "  build      Build the application binary"
	@echo "  test       Run all tests"
	@echo "  lint       Lint the codebase"
	@echo "  run        Build and run the application"
	@echo "  clean      Clean build artifacts"
	@echo ""
