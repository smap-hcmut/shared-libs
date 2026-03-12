# SMAP Shared Libraries Makefile

.PHONY: help build test clean install lint format

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build Go and Python packages"
	@echo "  test      - Run all tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install Python package locally"
	@echo "  lint      - Run linters"
	@echo "  format    - Format code"

# Build targets
build: build-go build-python

build-go:
	@echo "Building Go packages..."
	cd go && go build ./...

build-python:
	@echo "Building Python package..."
	cd python && python -m build

# Test targets
test: test-go test-python

test-go:
	@echo "Running Go tests..."
	cd go && go test -v ./...

test-python:
	@echo "Running Python tests..."
	cd python && pytest -v

# Clean targets
clean:
	@echo "Cleaning build artifacts..."
	cd go && go clean ./...
	cd python && rm -rf build/ dist/ *.egg-info/

# Install targets
install:
	@echo "Installing Python package locally..."
	cd python && pip install -e .

# Lint targets
lint: lint-go lint-python

lint-go:
	@echo "Linting Go code..."
	cd go && golangci-lint run

lint-python:
	@echo "Linting Python code..."
	cd python && black --check . && isort --check-only . && mypy .

# Format targets
format: format-go format-python

format-go:
	@echo "Formatting Go code..."
	cd go && go fmt ./...

format-python:
	@echo "Formatting Python code..."
	cd python && black . && isort .