# Integration Tests for Distributed Tracing

Comprehensive test suite for validating the shared library's distributed tracing functionality across Go and Python components.

## Test Categories

### 🔄 Cross-Language Integration Tests
- **Go → Python trace propagation**: Tests trace_id consistency when Go services call Python services
- **HTTP → Kafka → Database flow**: End-to-end trace propagation through different protocols
- **Service boundary validation**: Multi-hop trace propagation across service chains
- **Concurrent request isolation**: Ensures trace_id isolation in concurrent operations

### ⚡ Performance Tests
- **Latency impact**: Validates <1ms latency requirement per operation
- **Memory usage**: Tests memory consumption under high concurrency (1000+ requests)
- **Throughput**: Validates >10k operations/second capability
- **Benchmarks**: Performance benchmarks for trace generation, injection, and extraction

### 🛡️ Error Handling Tests
- **Invalid trace_id handling**: Tests recovery from malformed trace IDs
- **Network failure graceful degradation**: Ensures system stability during network issues
- **Context propagation failures**: Tests recovery when trace context is lost
- **Concurrent access safety**: Thread safety validation under high concurrency

## Running Tests

### Quick Run (All Tests)
```bash
./run_tests.sh
```

### Individual Test Categories

#### Go Tests
```bash
# Cross-language integration
go test -v -run TestGoToPythonTraceFlow

# Performance tests
go test -v -run TestHTTPLatencyImpact
go test -v -run TestMemoryUsageUnderLoad

# Error handling
go test -v -run TestInvalidTraceIDHandling
go test -v -run TestNetworkFailureGracefulDegradation

# Benchmarks
go test -bench=. -benchmem
```

#### Python Tests
```bash
# Set Python path
export PYTHONPATH="../python:$PYTHONPATH"

# Cross-language integration
python test_cross_language.py

# Performance tests
python test_performance.py

# Error handling
python test_error_handling.py
```

## Test Requirements

### Performance Requirements
- **Latency**: <1ms average latency per trace operation
- **Memory**: <10KB memory usage per request
- **Throughput**: >10,000 operations per second
- **Concurrency**: Support 1000+ concurrent requests

### Reliability Requirements
- **Error Recovery**: Graceful handling of invalid trace IDs
- **Network Resilience**: Continue operation during network failures
- **Thread Safety**: Safe concurrent access across multiple threads
- **Resource Cleanup**: No memory leaks during error scenarios

## Test Structure

```
integration-tests/
├── cross_language_test.go          # Go cross-language tests
├── performance_test.go             # Go performance tests
├── error_handling_test.go          # Go error handling tests
├── test_cross_language.py          # Python cross-language tests
├── test_performance.py             # Python performance tests
├── test_error_handling.py          # Python error handling tests
├── run_tests.sh                    # Test runner script
├── go.mod                          # Go module definition
└── README.md                       # This file
```

## Expected Results

### Successful Test Run Output
```
🧪 Running Distributed Tracing Integration Tests
================================================

📦 Running Go Integration Tests
--------------------------------
✓ Cross-language tests passed
✓ Performance tests passed  
✓ Error handling tests passed

🐍 Running Python Integration Tests
-----------------------------------
✓ Python cross-language tests passed
✓ Python performance tests passed
✓ Python error handling tests passed

📊 Test Summary
===============
Go tests passed: 3/3
Python tests passed: 3/3
Total tests passed: 6/6

🎉 All integration tests passed!
✅ Shared library is ready for service migration
```

### Performance Benchmarks
Expected benchmark results:
- Trace ID generation: >100,000 ops/sec
- HTTP trace injection: <0.1ms average latency
- Kafka trace injection: <0.1ms average latency
- Full trace flow: <1ms average latency
- Memory usage: <10KB per request

## Troubleshooting

### Common Issues

1. **Go module issues**
   ```bash
   cd integration-tests
   go mod tidy
   ```

2. **Python import errors**
   ```bash
   export PYTHONPATH="../python:$PYTHONPATH"
   ```

3. **Permission issues**
   ```bash
   chmod +x run_tests.sh
   ```

### Test Failures
If tests fail, check:
- Shared library compilation: `cd ../go && go build ./...`
- Python imports: `cd ../python && python -c "import smap_shared"`
- Dependencies: Ensure all required packages are installed

## Integration with CI/CD

This test suite is designed to be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Run Integration Tests
  run: |
    cd smap-shared-libs/integration-tests
    ./run_tests.sh
```

The test runner returns appropriate exit codes:
- `0`: All tests passed
- `1`: Some tests failed

## Next Steps

After all integration tests pass:
1. ✅ **Task 9 Complete**: Comprehensive testing suite validated
2. ⏭️ **Task 10**: Begin Go service migrations
3. ⏭️ **Task 11**: Begin Python service migrations
4. ⏭️ **Task 12**: Service migration validation checkpoint