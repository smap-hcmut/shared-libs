#!/bin/bash

# Comprehensive test runner for distributed tracing integration tests
# Tests both Go and Python components with performance and error handling

set -e

echo "🧪 Running Distributed Tracing Integration Tests"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
GO_TESTS_PASSED=0
PYTHON_TESTS_PASSED=0
TOTAL_TESTS=0

# Function to run Go tests
run_go_tests() {
    echo -e "${BLUE}📦 Running Go Integration Tests${NC}"
    echo "--------------------------------"
    
    cd "$(dirname "$0")"
    
    # Initialize Go module if needed
    if [ ! -f "go.sum" ]; then
        echo "Initializing Go module..."
        go mod tidy
    fi
    
    # Run cross-language tests
    echo "Running cross-language integration tests..."
    if go test -v -run TestGoToPythonTraceFlow ./...; then
        echo -e "${GREEN}✓ Cross-language tests passed${NC}"
        ((GO_TESTS_PASSED++))
    else
        echo -e "${RED}✗ Cross-language tests failed${NC}"
    fi
    ((TOTAL_TESTS++))
    
    # Run performance tests
    echo "Running performance benchmarks..."
    if go test -v -run TestHTTPLatencyImpact ./...; then
        echo -e "${GREEN}✓ Performance tests passed${NC}"
        ((GO_TESTS_PASSED++))
    else
        echo -e "${RED}✗ Performance tests failed${NC}"
    fi
    ((TOTAL_TESTS++))
    
    # Run error handling tests
    echo "Running error handling tests..."
    if go test -v -run TestInvalidTraceIDHandling ./...; then
        echo -e "${GREEN}✓ Error handling tests passed${NC}"
        ((GO_TESTS_PASSED++))
    else
        echo -e "${RED}✗ Error handling tests failed${NC}"
    fi
    ((TOTAL_TESTS++))
    
    # Run benchmarks
    echo "Running performance benchmarks..."
    go test -bench=. -benchmem ./... | grep -E "(Benchmark|PASS|FAIL)"
    
    echo ""
}

# Function to run Python tests
run_python_tests() {
    echo -e "${BLUE}🐍 Running Python Integration Tests${NC}"
    echo "-----------------------------------"
    
    cd "$(dirname "$0")"
    
    # Set Python path
    export PYTHONPATH="../python:$PYTHONPATH"
    
    # Run cross-language tests
    echo "Running cross-language integration tests..."
    if python test_cross_language.py; then
        echo -e "${GREEN}✓ Python cross-language tests passed${NC}"
        ((PYTHON_TESTS_PASSED++))
    else
        echo -e "${RED}✗ Python cross-language tests failed${NC}"
    fi
    ((TOTAL_TESTS++))
    
    # Run performance tests
    echo "Running performance tests..."
    if python test_performance.py; then
        echo -e "${GREEN}✓ Python performance tests passed${NC}"
        ((PYTHON_TESTS_PASSED++))
    else
        echo -e "${RED}✗ Python performance tests failed${NC}"
    fi
    ((TOTAL_TESTS++))
    
    # Run error handling tests
    echo "Running error handling tests..."
    if python test_error_handling.py; then
        echo -e "${GREEN}✓ Python error handling tests passed${NC}"
        ((PYTHON_TESTS_PASSED++))
    else
        echo -e "${RED}✗ Python error handling tests failed${NC}"
    fi
    ((TOTAL_TESTS++))
    
    echo ""
}

# Function to display summary
show_summary() {
    echo -e "${BLUE}📊 Test Summary${NC}"
    echo "==============="
    echo "Go tests passed: $GO_TESTS_PASSED/3"
    echo "Python tests passed: $PYTHON_TESTS_PASSED/3"
    echo "Total tests passed: $((GO_TESTS_PASSED + PYTHON_TESTS_PASSED))/$TOTAL_TESTS"
    
    if [ $((GO_TESTS_PASSED + PYTHON_TESTS_PASSED)) -eq $TOTAL_TESTS ]; then
        echo -e "${GREEN}🎉 All integration tests passed!${NC}"
        echo -e "${GREEN}✅ Shared library is ready for service migration${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some tests failed${NC}"
        echo -e "${YELLOW}⚠️  Please fix failing tests before proceeding${NC}"
        exit 1
    fi
}

# Main execution
main() {
    echo "Starting comprehensive integration test suite..."
    echo "Testing: Cross-language compatibility, Performance, Error handling"
    echo ""
    
    # Check dependencies
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed${NC}"
        exit 1
    fi
    
    if ! command -v python &> /dev/null && ! command -v python3 &> /dev/null; then
        echo -e "${RED}Error: Python is not installed${NC}"
        exit 1
    fi
    
    # Run tests
    run_go_tests
    run_python_tests
    
    # Show summary
    show_summary
}

# Execute main function
main "$@"