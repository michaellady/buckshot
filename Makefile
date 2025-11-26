.PHONY: build test lint coverage clean install

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"
BINARY := buckshot

# Default target
all: lint test build

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/buckshot

# Run tests
test:
	go test -v ./...

# Run tests with short flag (skip integration tests)
test-short:
	go test -short -v ./...

# Run linter
lint:
	golangci-lint run ./...

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML coverage report: go tool cover -html=coverage.out"

# Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	rm -f $(BINARY) coverage.out coverage.html

# Install the binary to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/buckshot

# Run the binary
run: build
	./$(BINARY)

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Check for vulnerabilities
vuln:
	govulncheck ./...
