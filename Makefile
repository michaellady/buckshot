.PHONY: build test test-short test-integration test-e2e lint coverage coverage-html coverage-pkg clean install

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"
BINARY := buckshot
COVERAGE_DIR := coverage
COVERAGE_THRESHOLD := 80

# Default target
all: lint test build

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/buckshot

# Build the mock agent for testing
build-mockagent:
	go build -o testdata/mockagent/mock-agent ./testdata/mockagent

# Run tests
test:
	go test -v ./...

# Run tests with short flag (skip integration tests)
test-short:
	go test -short -v ./...

# Run integration tests only
test-integration: build-mockagent
	go test -v -tags=integration ./...

# Run e2e tests against real agents (requires authenticated agents)
test-e2e:
	go test -v -tags=e2e ./...

# Run linter
lint:
	golangci-lint run ./...

# Generate coverage report
coverage:
	@mkdir -p $(COVERAGE_DIR)
	go test -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	go tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo ""
	@echo "To view HTML coverage report: make coverage-html"

# Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Generate per-package coverage reports
coverage-pkg:
	@mkdir -p $(COVERAGE_DIR)
	@echo "Generating per-package coverage reports..."
	@for pkg in $$(go list ./... | grep -v /testdata); do \
		name=$$(echo $$pkg | sed 's|github.com/michaellady/buckshot/||' | tr '/' '_'); \
		go test -coverprofile=$(COVERAGE_DIR)/$$name.out $$pkg 2>/dev/null || true; \
	done
	@echo ""
	@echo "Package coverage summary:"
	@for f in $(COVERAGE_DIR)/*.out; do \
		if [ -s "$$f" ]; then \
			pkg=$$(basename $$f .out | tr '_' '/'); \
			total=$$(go tool cover -func=$$f | tail -1 | awk '{print $$3}'); \
			printf "  %-50s %s\n" "$$pkg" "$$total"; \
		fi; \
	done

# Check if coverage meets threshold
coverage-check: coverage
	@echo ""
	@echo "Checking coverage threshold ($(COVERAGE_THRESHOLD)%)..."
	@total=$$(go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1 | awk '{print $$3}' | tr -d '%'); \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc) -eq 1 ]; then \
		echo "FAIL: Coverage $$total% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "PASS: Coverage $$total% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf $(COVERAGE_DIR)
	rm -f testdata/mockagent/mock-agent

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
