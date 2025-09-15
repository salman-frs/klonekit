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

.PHONY: all build test lint clean run help deps security vuln-check coverage

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
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	else \
		echo "gosec is already installed"; \
	fi
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	else \
		echo "govulncheck is already installed"; \
	fi

# Run security analysis
security:
	@echo "Running security analysis..."
	gosec -conf .gosec.json ./...

# Run vulnerability check
vuln-check:
	@echo "Running vulnerability check..."
	govulncheck ./...

# Generate test coverage report
coverage:
	@echo "Running tests with coverage..."
	$(GO_TEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Coverage summary:"
	go tool cover -func=coverage.out | grep total

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
	@echo "  security   Run security analysis with gosec"
	@echo "  vuln-check Run vulnerability check with govulncheck"
	@echo "  coverage   Generate test coverage report"
	@echo "  run        Build and run the application"
	@echo "  deps       Install development dependencies"
	@echo "  clean      Clean build artifacts"
	@echo ""
