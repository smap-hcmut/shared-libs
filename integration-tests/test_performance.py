"""
Performance and load testing for Python shared library components.

Tests latency impact, memory usage, and throughput requirements
to ensure <1ms latency impact per operation.
"""

import asyncio
import time
import threading
import psutil
import os
from typing import List
from statistics import mean, stdev

# Add shared library to path
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'python'))

from smap_shared.tracing import (
    generate_trace_id, set_trace_id, get_trace_id,
    http_propagator, kafka_propagator
)
from smap_shared.logger import Logger, LoggerConfig, LogLevel


class PerformanceTestSuite:
    """Performance test suite for Python tracing components."""
    
    def test_trace_id_generation_performance(self):
        """Test trace ID generation performance."""
        num_iterations = 10000
        
        start_time = time.perf_counter()
        
        for _ in range(num_iterations):
            trace_id = generate_trace_id()
            assert len(trace_id) > 0
        
        end_time = time.perf_counter()
        duration = end_time - start_time
        ops_per_second = num_iterations / duration
        
        # Should generate >100k trace IDs per second
        assert ops_per_second > 100000, f"Generation rate {ops_per_second:.0f} ops/s too slow"
        
        print(f"✓ Trace ID generation: {ops_per_second:.0f} ops/second")
        return ops_per_second
    
    def test_http_trace_injection_performance(self):
        """Test HTTP trace injection performance."""
        num_iterations = 10000
        trace_id = generate_trace_id()
        set_trace_id(trace_id)
        
        latencies = []
        
        for _ in range(num_iterations):
            start_time = time.perf_counter()
            
            headers = {}
            http_propagator.inject_http(headers)
            
            end_time = time.perf_counter()
            latencies.append((end_time - start_time) * 1000)  # Convert to ms
        
        avg_latency = mean(latencies)
        max_latency = max(latencies)
        
        # Should be <0.1ms per injection
        assert avg_latency < 0.1, f"Average latency {avg_latency:.3f}ms too high"
        assert max_latency < 1.0, f"Max latency {max_latency:.3f}ms too high"
        
        print(f"✓ HTTP injection: avg={avg_latency:.3f}ms, max={max_latency:.3f}ms")
        return avg_latency
    
    def test_kafka_trace_injection_performance(self):
        """Test Kafka trace injection performance."""
        num_iterations = 10000
        trace_id = generate_trace_id()
        set_trace_id(trace_id)
        
        latencies = []
        
        for _ in range(num_iterations):
            start_time = time.perf_counter()
            
            headers = {}
            kafka_propagator.inject_kafka(headers)
            
            end_time = time.perf_counter()
            latencies.append((end_time - start_time) * 1000)  # Convert to ms
        
        avg_latency = mean(latencies)
        max_latency = max(latencies)
        
        # Should be <0.1ms per injection
        assert avg_latency < 0.1, f"Average latency {avg_latency:.3f}ms too high"
        assert max_latency < 1.0, f"Max latency {max_latency:.3f}ms too high"
        
        print(f"✓ Kafka injection: avg={avg_latency:.3f}ms, max={max_latency:.3f}ms")
        return avg_latency
    
    def test_full_trace_flow_latency(self):
        """Test full trace flow latency impact (<1ms requirement)."""
        num_requests = 1000
        latencies = []
        
        for _ in range(num_requests):
            start_time = time.perf_counter()
            
            # Full trace flow simulation
            trace_id = generate_trace_id()
            set_trace_id(trace_id)
            
            # HTTP injection
            http_headers = {}
            http_propagator.inject_http(http_headers)
            
            # HTTP extraction (simulating receiving service)
            http_propagator.extract_http(http_headers)
            extracted_trace_id = get_trace_id()
            
            # Kafka injection
            kafka_headers = {}
            kafka_propagator.inject_kafka(kafka_headers)
            
            # Kafka extraction
            kafka_propagator.extract_kafka(kafka_headers)
            
            end_time = time.perf_counter()
            latencies.append((end_time - start_time) * 1000)  # Convert to ms
            
            # Verify correctness
            assert extracted_trace_id == trace_id
        
        avg_latency = mean(latencies)
        max_latency = max(latencies)
        p95_latency = sorted(latencies)[int(0.95 * len(latencies))]
        
        # Performance requirements: <1ms per operation
        assert avg_latency < 1.0, f"Average latency {avg_latency:.3f}ms should be <1ms"
        assert p95_latency < 2.0, f"P95 latency {p95_latency:.3f}ms should be <2ms"
        
        print(f"✓ Full trace flow: avg={avg_latency:.3f}ms, p95={p95_latency:.3f}ms, max={max_latency:.3f}ms")
        return avg_latency
    def test_memory_usage_under_load(self):
        """Test memory usage under high concurrency (1000+ requests)."""
        process = psutil.Process(os.getpid())
        memory_before = process.memory_info().rss / 1024 / 1024  # MB
        
        num_threads = 100
        requests_per_thread = 100
        results = []
        
        def worker(worker_id: int):
            """Worker thread simulating request processing."""
            thread_results = []
            
            for i in range(requests_per_thread):
                # Simulate request with tracing
                trace_id = generate_trace_id()
                set_trace_id(trace_id)
                
                # Simulate processing
                headers = {}
                http_propagator.inject_http(headers)
                kafka_headers = {}
                kafka_propagator.inject_kafka(kafka_headers)
                
                # Verify trace ID consistency
                current_trace_id = get_trace_id()
                thread_results.append(current_trace_id == trace_id)
                
                # Small delay to simulate work
                time.sleep(0.001)
            
            results.extend(thread_results)
        
        # Launch concurrent threads
        threads = []
        for i in range(num_threads):
            thread = threading.Thread(target=worker, args=(i,))
            threads.append(thread)
            thread.start()
        
        # Wait for completion
        for thread in threads:
            thread.join()
        
        memory_after = process.memory_info().rss / 1024 / 1024  # MB
        memory_used = memory_after - memory_before
        total_requests = num_threads * requests_per_thread
        memory_per_request = (memory_used * 1024) / total_requests  # KB per request
        
        # Verify all requests processed correctly
        assert all(results), "Some requests failed trace ID consistency check"
        
        # Memory usage should be reasonable (<10KB per request)
        assert memory_per_request < 10.0, f"Memory per request {memory_per_request:.2f}KB too high"
        
        print(f"✓ Memory usage: {memory_used:.2f}MB total, {memory_per_request:.2f}KB per request")
        return memory_per_request
    
    def test_logging_performance_with_tracing(self):
        """Test logging performance impact with trace_id injection."""
        config = LoggerConfig(
            level=LogLevel.INFO,
            enable_trace_id=True,
            json_output=False
        )
        logger = Logger(config)
        
        num_logs = 1000
        latencies = []
        
        for i in range(num_logs):
            trace_id = generate_trace_id()
            set_trace_id(trace_id)
            
            start_time = time.perf_counter()
            
            # Log with trace context
            logger.info(f"Processing request {i}", extra={"user_id": i, "action": "test"})
            
            end_time = time.perf_counter()
            latencies.append((end_time - start_time) * 1000)  # Convert to ms
        
        avg_latency = mean(latencies)
        max_latency = max(latencies)
        
        # Logging with trace_id should add minimal overhead (<1ms)
        assert avg_latency < 1.0, f"Average logging latency {avg_latency:.3f}ms too high"
        
        print(f"✓ Logging performance: avg={avg_latency:.3f}ms, max={max_latency:.3f}ms")
        return avg_latency
    
    async def test_async_operations_performance(self):
        """Test performance of async operations with tracing."""
        num_operations = 1000
        latencies = []
        
        async def async_operation(operation_id: int):
            """Simulate async operation with tracing."""
            start_time = time.perf_counter()
            
            trace_id = generate_trace_id()
            set_trace_id(trace_id)
            
            # Simulate async work
            await asyncio.sleep(0.001)
            
            # Verify trace context
            current_trace_id = get_trace_id()
            assert current_trace_id == trace_id
            
            end_time = time.perf_counter()
            return (end_time - start_time) * 1000  # Convert to ms
        
        # Run async operations concurrently
        tasks = [async_operation(i) for i in range(num_operations)]
        latencies = await asyncio.gather(*tasks)
        
        avg_latency = mean(latencies)
        max_latency = max(latencies)
        
        # Async operations should maintain good performance
        assert avg_latency < 5.0, f"Average async latency {avg_latency:.3f}ms too high"
        
        print(f"✓ Async operations: avg={avg_latency:.3f}ms, max={max_latency:.3f}ms")
        return avg_latency


class LoadTestSuite:
    """Load testing for high-throughput scenarios."""
    
    def test_high_throughput_trace_processing(self):
        """Test high-throughput trace processing (>10k ops/second)."""
        duration_seconds = 5
        operations_count = 0
        
        start_time = time.time()
        end_time = start_time + duration_seconds
        
        while time.time() < end_time:
            # High-speed trace operations
            trace_id = generate_trace_id()
            set_trace_id(trace_id)
            
            headers = {}
            http_propagator.inject_http(headers)
            http_propagator.extract_http(headers)
            
            kafka_headers = {}
            kafka_propagator.inject_kafka(kafka_headers)
            kafka_propagator.extract_kafka(kafka_headers)
            
            operations_count += 1
        
        actual_duration = time.time() - start_time
        throughput = operations_count / actual_duration
        
        # Should achieve >10k operations per second
        assert throughput > 10000, f"Throughput {throughput:.0f} ops/s should be >10k ops/s"
        
        print(f"✓ High throughput: {throughput:.0f} operations/second")
        return throughput


def run_performance_tests():
    """Run all performance tests."""
    print("🚀 Running Python Performance Tests...")
    
    perf_suite = PerformanceTestSuite()
    load_suite = LoadTestSuite()
    
    # Performance tests
    perf_suite.test_trace_id_generation_performance()
    perf_suite.test_http_trace_injection_performance()
    perf_suite.test_kafka_trace_injection_performance()
    perf_suite.test_full_trace_flow_latency()
    perf_suite.test_memory_usage_under_load()
    perf_suite.test_logging_performance_with_tracing()
    
    # Async performance test
    asyncio.run(perf_suite.test_async_operations_performance())
    
    # Load tests
    load_suite.test_high_throughput_trace_processing()
    
    print("✅ All Python performance tests passed!")


if __name__ == "__main__":
    run_performance_tests()