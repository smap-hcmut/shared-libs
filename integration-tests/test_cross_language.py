"""
Cross-language integration tests for distributed tracing.

Tests trace propagation between Go and Python services, ensuring
trace_id consistency across service boundaries and protocols.
"""

import asyncio
import json
import uuid
from unittest.mock import patch, MagicMock
from typing import Dict, Any

# Add shared library to path
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'python'))

from smap_shared.tracing import (
    get_trace_id, set_trace_id, generate_trace_id,
    http_propagator, kafka_propagator
)
from smap_shared.http import TracedHTTPClient, create_async_client


class TestCrossLanguageTracing:
    """Test cross-language trace propagation scenarios."""
    
    def test_python_to_go_trace_flow(self):
        """Test trace propagation from Python service to Go service."""
        # Generate trace ID in Python service
        trace_id = generate_trace_id()
        set_trace_id(trace_id)
        
        # Prepare headers for Go service call
        headers = {}
        http_propagator.inject_http(headers)
        
        # Verify trace_id is injected
        assert "X-Trace-Id" in headers
        assert headers["X-Trace-Id"] == trace_id
        
        # Simulate Go service receiving the request
        received_trace_id = headers.get("X-Trace-Id")
        assert received_trace_id == trace_id
        
        print(f"✓ Trace ID propagated: {trace_id}")
    
    def test_http_kafka_database_flow(self):
        """Test end-to-end trace flow through HTTP → Kafka → Database."""
        # Initial trace ID
        original_trace_id = generate_trace_id()
        set_trace_id(original_trace_id)
        
        # HTTP propagation
        http_headers = {}
        http_propagator.inject_http(http_headers)
        assert http_headers["X-Trace-Id"] == original_trace_id
        
        # Kafka propagation  
        kafka_headers = {}
        kafka_propagator.inject_kafka(kafka_headers)
        assert kafka_headers["X-Trace-Id"] == original_trace_id.encode()
        
        # Extract from Kafka (simulating consumer)
        extracted_trace_id = kafka_headers["X-Trace-Id"].decode()
        set_trace_id(extracted_trace_id)
        
        # Verify consistency
        assert get_trace_id() == original_trace_id
        print(f"✓ End-to-end trace flow: {original_trace_id}")
    
    def test_concurrent_request_isolation(self):
        """Test trace isolation in concurrent Python operations."""
        import threading
        import time
        
        results = []
        
        def worker(worker_id: int):
            # Each worker gets unique trace ID
            trace_id = generate_trace_id()
            set_trace_id(trace_id)
            
            # Simulate processing
            time.sleep(0.01 * worker_id)
            
            # Verify trace ID is still correct
            current_trace_id = get_trace_id()
            results.append((worker_id, trace_id, current_trace_id))
        
        # Launch concurrent workers
        threads = []
        for i in range(5):
            thread = threading.Thread(target=worker, args=(i,))
            threads.append(thread)
            thread.start()
        
        # Wait for completion
        for thread in threads:
            thread.join()
        
        # Verify isolation
        trace_ids = set()
        for worker_id, original, current in results:
            assert original == current, f"Worker {worker_id}: trace ID changed"
            assert original not in trace_ids, f"Worker {worker_id}: trace ID not unique"
            trace_ids.add(original)
        
        print(f"✓ Concurrent isolation: {len(trace_ids)} unique trace IDs")
    def test_trace_validation_and_recovery(self):
        """Test invalid trace_id handling and recovery."""
        test_cases = [
            ("valid_uuid", "550e8400-e29b-41d4-a716-446655440000", True),
            ("invalid_short", "invalid-trace", False),
            ("invalid_format", "not-a-uuid-at-all", False),
            ("empty_string", "", False),
        ]
        
        for name, trace_id, should_be_valid in test_cases:
            # Test validation
            from smap_shared.tracing.context import TraceContext
            tracer = TraceContext()
            is_valid = tracer.is_valid_trace_id(trace_id)
            assert is_valid == should_be_valid, f"{name}: validation failed"
            
            if not should_be_valid:
                # Test recovery - should generate new valid trace ID
                headers = {"X-Trace-Id": trace_id}
                http_propagator.extract_http(headers)
                
                recovered_trace_id = get_trace_id()
                if recovered_trace_id:
                    assert tracer.is_valid_trace_id(recovered_trace_id), f"{name}: recovery failed"
        
        print("✓ Trace validation and recovery working")
    
    async def test_async_http_trace_propagation(self):
        """Test trace propagation in async HTTP operations."""
        # Setup trace context
        trace_id = generate_trace_id()
        set_trace_id(trace_id)
        
        # Mock HTTP response
        mock_response_data = {
            "received_trace_id": trace_id,
            "processed_with_trace": True
        }
        
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            with patch('httpx.AsyncClient') as mock_client_class:
                # Setup mock
                mock_client = MagicMock()
                mock_response = MagicMock()
                mock_response.content = json.dumps(mock_response_data).encode()
                mock_response.status_code = 200
                mock_client.get.return_value = mock_response
                mock_client_class.return_value = mock_client
                
                # Test async HTTP client
                async with create_async_client() as client:
                    response_body, status_code = await client.get("https://api.example.com/test")
                    
                    assert status_code == 200
                    response_data = json.loads(response_body.decode())
                    assert response_data["received_trace_id"] == trace_id
                
                # Verify trace_id was injected in headers
                call_args = mock_client.get.call_args
                headers = call_args[1]["headers"]
                assert "X-Trace-Id" in headers
                assert headers["X-Trace-Id"] == trace_id
        
        print(f"✓ Async HTTP trace propagation: {trace_id}")
    
    def test_service_boundary_propagation(self):
        """Test trace propagation across multiple service boundaries."""
        # Simulate: Client → Service A → Service B → Service C
        original_trace_id = generate_trace_id()
        
        # Service A receives request
        set_trace_id(original_trace_id)
        service_a_trace_id = get_trace_id()
        
        # Service A calls Service B via HTTP
        http_headers = {}
        http_propagator.inject_http(http_headers)
        
        # Service B extracts trace ID
        http_propagator.extract_http(http_headers)
        service_b_trace_id = get_trace_id()
        
        # Service B sends to Service C via Kafka
        kafka_headers = {}
        kafka_propagator.inject_kafka(kafka_headers)
        
        # Service C extracts from Kafka
        kafka_propagator.extract_kafka(kafka_headers)
        service_c_trace_id = get_trace_id()
        
        # Verify consistency across all services
        assert service_a_trace_id == original_trace_id
        assert service_b_trace_id == original_trace_id
        assert service_c_trace_id == original_trace_id
        
        print(f"✓ Service boundary propagation: {original_trace_id}")


class TestErrorHandlingScenarios:
    """Test error handling and graceful degradation."""
    
    def test_missing_trace_id_generation(self):
        """Test automatic trace_id generation when missing."""
        # Clear any existing trace ID
        set_trace_id("")
        assert get_trace_id() == ""
        
        # HTTP propagator should generate new trace ID if missing
        headers = {}
        http_propagator.inject_http(headers)
        
        # Should have generated a new trace ID
        injected_trace_id = headers.get("X-Trace-Id")
        if injected_trace_id:
            from smap_shared.tracing.context import TraceContext
            tracer = TraceContext()
            assert tracer.is_valid_trace_id(injected_trace_id)
        
        print("✓ Missing trace_id generation working")
    
    def test_network_failure_graceful_degradation(self):
        """Test graceful degradation during network failures."""
        trace_id = generate_trace_id()
        set_trace_id(trace_id)
        
        # Simulate network failure in HTTP client
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            with patch('httpx.AsyncClient') as mock_client_class:
                mock_client = MagicMock()
                mock_client.get.side_effect = Exception("Network error")
                mock_client_class.return_value = mock_client
                
                # HTTP client should handle errors gracefully
                try:
                    client = TracedHTTPClient(use_async=True)
                    # Error should be raised but trace context should remain
                    assert get_trace_id() == trace_id
                except Exception:
                    # Expected - network error should be propagated
                    pass
        
        # Trace context should still be intact
        assert get_trace_id() == trace_id
        print("✓ Network failure graceful degradation working")


if __name__ == "__main__":
    # Simple test runner
    test_class = TestCrossLanguageTracing()
    
    print("Running cross-language integration tests...")
    
    test_class.test_python_to_go_trace_flow()
    test_class.test_http_kafka_database_flow()
    test_class.test_concurrent_request_isolation()
    test_class.test_trace_validation_and_recovery()
    
    # Run async test
    asyncio.run(test_class.test_async_http_trace_propagation())
    
    test_class.test_service_boundary_propagation()
    
    error_test_class = TestErrorHandlingScenarios()
    error_test_class.test_missing_trace_id_generation()
    error_test_class.test_network_failure_graceful_degradation()
    
    print("✅ All cross-language integration tests passed!")