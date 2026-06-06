# SMAP Shared Libraries Makefile (Go only)

.PHONY: help build test clean lint format

help:
	@echo "Available targets:"
	@echo "  build     - Build Go packages"
	@echo "  test      - Run Go tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  lint      - Run golangci-lint"
	@echo "  format    - Run go fmt"

build:
	@echo "Building Go packages..."
	cd go && go build ./...

test:
	@echo "Running Go tests..."
	cd go && go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	cd go && go clean ./...

lint:
	@echo "Linting Go code..."
	cd go && golangci-lint run

format:
	@echo "Formatting Go code..."
	cd go && go fmt ./...
