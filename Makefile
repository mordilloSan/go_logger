.PHONY: test fmt vet all clean help

# Default target
all: fmt vet test

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run static analysis
vet:
	@echo "Running go vet..."
	@go vet ./...

# Clean build artifacts and cache
clean:
	@echo "Cleaning..."
	@go clean -cache -testcache

# Pre-release check: format, vet, and test
pre-release: fmt vet test
	@echo "✓ All pre-release checks passed!"
	@echo "Ready to create release tag"

# Help target
help:
	@echo "Available targets:"
	@echo "  make test        - Run all tests"
	@echo "  make fmt         - Format code"
	@echo "  make vet         - Run static analysis"
	@echo "  make all         - Run fmt, vet, and test (default)"
	@echo "  make pre-release - Run all checks before release"
	@echo "  make clean       - Clean build cache"
	@echo "  make help        - Show this help message"
