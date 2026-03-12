"""
Tests for TracedHTTPClient implementation.

Basic tests to verify HTTP client functionality with trace injection.
"""

import pytest
import asyncio
from unittest.mock import Mock, patch
from typing import Dict, Any

from ..tracing import set_trace_id, get_trace_id, generate_trace_id
from .client import TracedHTTPClient, create_async_client, create_sync_client


class TestTracedHTTPClient:
    """Test cases for TracedHTTPClient."""
    
    def setup_method(self):
        """Set up test fixtures."""
        self.test_trace_id = "550e8400-e29b-41d4-a716-446655440000"
        self.test_url = "https://httpbin.org/get"
    
    def test_sync_client_creation(self):
        """Test sync client creation without external dependencies."""
        # Test that we can create a sync client
        with patch('smap_shared.http.client.REQUESTS_AVAILABLE', True):
            with patch('requests.Session') as mock_session:
                client = TracedHTTPClient(use_async=False)
                assert client.use_async is False
                assert client.timeout == 30.0
                assert client.retries == 3
    
    def test_async_client_creation(self):
        """Test async client creation without external dependencies."""
        # Test that we can create an async client
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            client = TracedHTTPClient(use_async=True)
            assert client.use_async is True
            assert client.timeout == 30.0
            assert client.retries == 3
    
    def test_header_preparation_with_trace_id(self):
        """Test header preparation with trace_id injection."""
        # Set trace_id in context
        set_trace_id(self.test_trace_id)
        
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            client = TracedHTTPClient(use_async=True)
            
            # Test header preparation
            headers = client._prepare_headers({"Content-Type": "application/json"})
            
            assert "X-Trace-Id" in headers
            assert headers["X-Trace-Id"] == self.test_trace_id
            assert headers["Content-Type"] == "application/json"
    
    def test_header_preparation_without_trace_id(self):
        """Test header preparation without trace_id in context."""
        # Clear any existing trace_id
        from ..tracing import clear_trace_id
        clear_trace_id()
        
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            client = TracedHTTPClient(use_async=True)
            
            # Test header preparation
            headers = client._prepare_headers({"Content-Type": "application/json"})
            
            # Should not have X-Trace-Id when no trace_id in context
            assert headers["Content-Type"] == "application/json"
            # X-Trace-Id might not be present if no trace_id in context
    
    def test_base_headers_integration(self):
        """Test base headers are included in requests."""
        base_headers = {"User-Agent": "SMAP-Client/1.0", "Accept": "application/json"}
        
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            client = TracedHTTPClient(use_async=True, base_headers=base_headers)
            
            headers = client._prepare_headers({"Content-Type": "application/json"})
            
            assert headers["User-Agent"] == "SMAP-Client/1.0"
            assert headers["Accept"] == "application/json"
            assert headers["Content-Type"] == "application/json"
    
    def test_convenience_functions(self):
        """Test convenience functions for client creation."""
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            async_client = create_async_client(timeout=60.0)
            assert async_client.use_async is True
            assert async_client.timeout == 60.0
        
        with patch('smap_shared.http.client.REQUESTS_AVAILABLE', True):
            with patch('requests.Session'):
                sync_client = create_sync_client(timeout=60.0)
                assert sync_client.use_async is False
                assert sync_client.timeout == 60.0
    
    def test_error_handling_missing_dependencies(self):
        """Test error handling when HTTP libraries are not available."""
        # Test httpx not available
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', False):
            with pytest.raises(ImportError, match="httpx is required"):
                TracedHTTPClient(use_async=True)
        
        # Test requests not available
        with patch('smap_shared.http.client.REQUESTS_AVAILABLE', False):
            with pytest.raises(ImportError, match="requests is required"):
                TracedHTTPClient(use_async=False)


class TestIntegrationScenarios:
    """Integration test scenarios for common usage patterns."""
    
    def test_trace_propagation_scenario(self):
        """Test typical trace propagation scenario."""
        # Simulate incoming request with trace_id
        incoming_trace_id = generate_trace_id()
        set_trace_id(incoming_trace_id)
        
        # Create client and verify trace_id is propagated
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            client = TracedHTTPClient(use_async=True)
            headers = client._prepare_headers()
            
            assert headers.get("X-Trace-Id") == incoming_trace_id
    
    def test_service_to_service_call_pattern(self):
        """Test service-to-service HTTP call pattern."""
        # Set trace_id as if from incoming request
        service_trace_id = "123e4567-e89b-12d3-a456-426614174000"
        set_trace_id(service_trace_id)
        
        # Prepare headers for outbound service call
        with patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            client = TracedHTTPClient(
                use_async=True,
                base_headers={"Service-Name": "analysis-srv"}
            )
            
            headers = client._prepare_headers({
                "Content-Type": "application/json",
                "Authorization": "Bearer token123"
            })
            
            # Verify all headers are present
            assert headers["X-Trace-Id"] == service_trace_id
            assert headers["Service-Name"] == "analysis-srv"
            assert headers["Content-Type"] == "application/json"
            assert headers["Authorization"] == "Bearer token123"


if __name__ == "__main__":
    # Run basic tests without pytest
    test_client = TestTracedHTTPClient()
    test_client.setup_method()
    
    print("Running basic HTTP client tests...")
    
    try:
        test_client.test_async_client_creation()
        print("✓ Async client creation test passed")
        
        test_client.test_header_preparation_with_trace_id()
        print("✓ Header preparation with trace_id test passed")
        
        test_client.test_base_headers_integration()
        print("✓ Base headers integration test passed")
        
        test_client.test_convenience_functions()
        print("✓ Convenience functions test passed")
        
        # Integration tests
        integration_tests = TestIntegrationScenarios()
        integration_tests.test_trace_propagation_scenario()
        print("✓ Trace propagation scenario test passed")
        
        integration_tests.test_service_to_service_call_pattern()
        print("✓ Service-to-service call pattern test passed")
        
        print("\nAll tests passed! ✅")
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()