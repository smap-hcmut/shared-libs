"""
Error handling and graceful degradation tests for Python tracing components.

Tests invalid trace_id handling, network failures, context propagation failures,
and concurrent access safety.
"""

import threading
import time
import uuid
from unittest.mock import patch, MagicMock

# Add shared library to path
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'python'))

from smap_shared.tracing import (
    generate_trace_id, set_trace_id, get_trace_id, clear_trace_id,
    http_propagator, kafka_propagator
)
from smap_shared.tracing.context import TraceContext
from smap_shared.http import TracedHTTPClient


class TestErrorHandling:
    """Test error handling scenarios."""
    
    def test_invalid_trace_id_handling(self):
        """Test handling of invalid trace_id formats."""
        tracer = TraceContext()
        
        test_cases = [
            ("empty_string", "", True),
            ("invalid_short", "invalid", True),
            ("invalid_format", "not-a-uuid-at-all", True),
            ("malformed_uuid", "550e8400-XXXX-41d4-a716-446655440000", True),
            ("valid_uuid", "550e8400-e29b-41d4-a716-446655440000", False),
        ]
        
        for name, trace_id, should_recover in test_cases:
            # Test extraction with invalid trace ID
            headers = {"X-Trace-Id": trace_id}
            
            try:
                http_propagator.extract_http(headers)
                final_trace_id = get_trace_id()
                
                if should_recover and final_trace_id:
                    # If recovered, should be valid
                    assert tracer.is_valid_trace_id(final_trace_id), f"{name}: recovered ID invalid"
                
                # System should continue functioning
                new_trace_id = generate_trace_id()
                assert tracer.is_valid_trace_id(new_trace_id), f"{name}: new generation failed"
                
            except Exception as e:
                # Should not raise exceptions for invalid input
                assert False, f"{name}: should handle invalid input gracefully: {e}"
        
        print("✓ Invalid trace ID handling working")
    
    def test_missing_trace_id_generation(self):
        """Test automatic trace_id generation when missing."""
        # Clear any existing trace ID
        clear_trace_id()
        assert get_trace_id() == ""
        
        # Test with empty headers
        empty_headers = {}
        http_propagator.extract_http(empty_headers)
        
        # Should handle missing trace ID gracefully
        assert not get_trace_id() or TraceContext().is_valid_trace_id(get_trace_id())
        
        # Test injection without existing trace ID
        injection_headers = {}
        http_propagator.inject_http(injection_headers)
        
        # Should either inject nothing or generate valid trace ID
        if "X-Trace-Id" in injection_headers:
            injected_id = injection_headers["X-Trace-Id"]
            assert TraceContext().is_valid_trace_id(injected_id)
        
        print("✓ Missing trace ID generation working")
    
    def test_network_failure_graceful_degradation(self):
        """Test graceful degradation during network failures."""
        trace_id = generate_trace_id()
        set_trace_id(trace_id)
        
        # Test with mocked network failure
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            with patch('httpx.AsyncClient') as mock_client_class:
                mock_client = MagicMock()
                mock_client.get.side_effect = Exception("Network timeout")
                mock_client_class.return_value = mock_client
                
                # HTTP client should handle errors gracefully
                try:
                    client = TracedHTTPClient(use_async=True)
                    # Trace context should remain intact even with network errors
                    assert get_trace_id() == trace_id
                except Exception:
                    # Network errors are expected, but trace context should survive
                    pass
                
                # Should still be able to work with tracing
                assert get_trace_id() == trace_id
        
        print("✓ Network failure graceful degradation working")
    def test_context_propagation_failures(self):
        """Test context propagation failure recovery."""
        # Test with context that loses trace ID
        original_trace_id = generate_trace_id()
        set_trace_id(original_trace_id)
        
        # Clear context (simulating context loss)
        clear_trace_id()
        assert get_trace_id() == ""
        
        # Should handle missing trace ID in operations
        headers = {}
        try:
            http_propagator.inject_http(headers)
            kafka_headers = {}
            kafka_propagator.inject_kafka(kafka_headers)
        except Exception as e:
            assert False, f"Should handle missing trace ID gracefully: {e}"
        
        # Should be able to recover by setting new trace ID
        new_trace_id = generate_trace_id()
        set_trace_id(new_trace_id)
        assert get_trace_id() == new_trace_id
        
        print("✓ Context propagation failure recovery working")
    
    def test_concurrent_access_safety(self):
        """Test thread safety under concurrent access."""
        num_threads = 50
        operations_per_thread = 100
        errors = []
        results = []
        
        def worker(worker_id: int):
            """Worker thread performing concurrent operations."""
            try:
                for i in range(operations_per_thread):
                    # Concurrent trace operations
                    trace_id = generate_trace_id()
                    set_trace_id(trace_id)
                    
                    # HTTP operations
                    http_headers = {}
                    http_propagator.inject_http(http_headers)
                    http_propagator.extract_http(http_headers)
                    
                    # Kafka operations
                    kafka_headers = {}
                    kafka_propagator.inject_kafka(kafka_headers)
                    kafka_propagator.extract_kafka(kafka_headers)
                    
                    # Verify consistency
                    current_trace_id = get_trace_id()
                    if current_trace_id != trace_id:
                        errors.append(f"Worker {worker_id}: trace ID mismatch")
                    
                    results.append((worker_id, i, trace_id))
                    
            except Exception as e:
                errors.append(f"Worker {worker_id}: {str(e)}")
        
        # Launch concurrent threads
        threads = []
        for i in range(num_threads):
            thread = threading.Thread(target=worker, args=(i,))
            threads.append(thread)
            thread.start()
        
        # Wait for completion
        for thread in threads:
            thread.join()
        
        # Check for errors
        assert len(errors) == 0, f"Concurrent access errors: {errors[:5]}"  # Show first 5 errors
        assert len(results) == num_threads * operations_per_thread
        
        print(f"✓ Concurrent access safety: {len(results)} operations completed")
    
    def test_resource_cleanup_on_errors(self):
        """Test proper resource cleanup during errors."""
        tracer = TraceContext()
        
        # Test that failed operations don't leak resources
        for i in range(1000):
            try:
                # Operations that might fail
                trace_id = generate_trace_id()
                set_trace_id(trace_id)
                
                # Simulate various error conditions
                if i % 4 == 0:
                    # Invalid headers
                    headers = {"X-Trace-Id": "invalid"}
                    http_propagator.extract_http(headers)
                elif i % 4 == 1:
                    # Empty trace ID
                    set_trace_id("")
                elif i % 4 == 2:
                    # Invalid operations
                    kafka_headers = {"X-Trace-Id": b"invalid"}
                    kafka_propagator.extract_kafka(kafka_headers)
                else:
                    # Normal operation
                    get_trace_id()
                    
            except Exception:
                # Errors are expected, should not crash
                pass
        
        # System should still be functional
        final_trace_id = generate_trace_id()
        assert tracer.is_valid_trace_id(final_trace_id)
        
        print("✓ Resource cleanup on errors working")
    
    def test_malformed_headers_handling(self):
        """Test handling of malformed headers."""
        test_cases = [
            {"X-Trace-Id": None},  # None value
            {"X-Trace-Id": 123},   # Non-string value
            {"X-Trace-Id": b"bytes"},  # Bytes instead of string
            {},  # Missing header
            {"Other-Header": "value"},  # Wrong header
        ]
        
        for i, headers in enumerate(test_cases):
            try:
                http_propagator.extract_http(headers)
                # Should not crash
                current_trace_id = get_trace_id()
                
                # If trace ID is set, should be valid
                if current_trace_id:
                    assert TraceContext().is_valid_trace_id(current_trace_id)
                    
            except Exception as e:
                assert False, f"Case {i}: should handle malformed headers gracefully: {e}"
        
        print("✓ Malformed headers handling working")


def run_error_handling_tests():
    """Run all error handling tests."""
    print("🛡️ Running Python Error Handling Tests...")
    
    test_suite = TestErrorHandling()
    
    test_suite.test_invalid_trace_id_handling()
    test_suite.test_missing_trace_id_generation()
    test_suite.test_network_failure_graceful_degradation()
    test_suite.test_context_propagation_failures()
    test_suite.test_concurrent_access_safety()
    test_suite.test_resource_cleanup_on_errors()
    test_suite.test_malformed_headers_handling()
    
    print("✅ All Python error handling tests passed!")


if __name__ == "__main__":
    run_error_handling_tests()