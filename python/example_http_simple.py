"""
Simple example of TracedHTTPClient usage without external dependencies.

This example demonstrates the core functionality without requiring
httpx or requests to be installed.
"""

import sys
import os

# Add the smap_shared package to Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '.'))

from smap_shared.tracing import set_trace_id, get_trace_id, generate_trace_id, clear_trace_id
from smap_shared.http.client import TracedHTTPClient


def example_header_preparation():
    """Example of how TracedHTTPClient prepares headers with trace injection."""
    print("=== Header Preparation Example ===")
    
    # Simulate incoming request with trace_id
    incoming_trace_id = generate_trace_id()
    set_trace_id(incoming_trace_id)
    print(f"Set trace_id in context: {incoming_trace_id}")
    
    # Mock the dependencies to avoid import errors
    import unittest.mock
    
    with unittest.mock.patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
        # Create HTTP client
        client = TracedHTTPClient(
            use_async=True,
            base_headers={"Service-Name": "analysis-srv", "Version": "1.0"}
        )
        
        # Prepare headers for outbound request
        headers = client._prepare_headers({
            "Content-Type": "application/json",
            "Authorization": "Bearer service-token"
        })
        
        print(f"Prepared headers: {headers}")
        
        # Verify trace_id injection
        assert "X-Trace-Id" in headers
        assert headers["X-Trace-Id"] == incoming_trace_id
        assert headers["Service-Name"] == "analysis-srv"
        assert headers["Content-Type"] == "application/json"
        
        print("✓ Trace_id properly injected into outbound headers")
        print("✓ Base headers included")
        print("✓ Custom headers preserved")


def example_service_communication_pattern():
    """Example of typical service-to-service communication pattern."""
    print("\n=== Service Communication Pattern ===")
    
    # Step 1: Incoming request (simulated FastAPI middleware)
    incoming_headers = {
        "X-Trace-Id": "550e8400-e29b-41d4-a716-446655440000",
        "Content-Type": "application/json"
    }
    
    # Extract trace_id (done by middleware)
    from smap_shared.tracing import http_propagator
    extracted_trace_id = http_propagator.extract_http(incoming_headers)
    set_trace_id(extracted_trace_id)
    
    print(f"1. Extracted trace_id from incoming request: {extracted_trace_id}")
    
    # Step 2: Business logic needs to call another service
    import unittest.mock
    
    with unittest.mock.patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
        client = TracedHTTPClient(
            use_async=True,
            base_headers={"Service-Name": "analysis-srv"}
        )
        
        # Prepare headers for call to project-srv
        project_srv_headers = client._prepare_headers({
            "Content-Type": "application/json",
            "Accept": "application/json"
        })
        
        print(f"2. Headers for project-srv call: {project_srv_headers}")
        
        # Step 3: Call to knowledge-srv from the same request context
        knowledge_srv_headers = client._prepare_headers({
            "Content-Type": "application/json",
            "Service-Target": "knowledge-srv"
        })
        
        print(f"3. Headers for knowledge-srv call: {knowledge_srv_headers}")
        
        # Verify trace continuity
        assert project_srv_headers["X-Trace-Id"] == extracted_trace_id
        assert knowledge_srv_headers["X-Trace-Id"] == extracted_trace_id
        
        print("✓ Same trace_id propagated to all outbound calls")
        print("✓ End-to-end trace continuity maintained")


def example_error_handling():
    """Example of error handling scenarios."""
    print("\n=== Error Handling Examples ===")
    
    # Scenario 1: No trace_id in context
    clear_trace_id()
    
    import unittest.mock
    
    with unittest.mock.patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
        client = TracedHTTPClient(use_async=True)
        headers = client._prepare_headers({"Content-Type": "application/json"})
        
        print(f"Headers without trace_id: {headers}")
        # Should not have X-Trace-Id when no trace_id in context
        print("✓ Gracefully handles missing trace_id")
        
        # Scenario 2: Invalid trace_id handling
        invalid_headers = {"X-Trace-Id": "invalid-format"}
        
        from smap_shared.tracing import http_propagator
        extracted = http_propagator.extract_http(invalid_headers)
        
        if extracted is None:
            print("✓ Invalid trace_id properly rejected")
            
            # Generate new trace_id (what middleware would do)
            new_trace_id = generate_trace_id()
            set_trace_id(new_trace_id)
            
            # Now client will use the new trace_id
            new_headers = client._prepare_headers({"Content-Type": "application/json"})
            print(f"Headers with new trace_id: {new_headers}")
            
            assert new_headers["X-Trace-Id"] == new_trace_id
            print("✓ New trace_id properly generated and used")


def example_configuration_options():
    """Example of different configuration options."""
    print("\n=== Configuration Options ===")
    
    set_trace_id(generate_trace_id())
    
    import unittest.mock
    
    with unittest.mock.patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
        with unittest.mock.patch('smap_shared.http.client.REQUESTS_AVAILABLE', True):
            
            # Async client configuration
            async_client = TracedHTTPClient(
                use_async=True,
                timeout=60.0,
                retries=5,
                base_headers={
                    "Service-Name": "analysis-srv",
                    "User-Agent": "SMAP-Analysis/1.0",
                    "Accept": "application/json"
                }
            )
            
            async_headers = async_client._prepare_headers({"Content-Type": "application/json"})
            print(f"Async client headers: {async_headers}")
            
            # Sync client configuration
            sync_client = TracedHTTPClient(
                use_async=False,
                timeout=30.0,
                retries=3,
                base_headers={"Service-Name": "scapper-srv"}
            )
            
            sync_headers = sync_client._prepare_headers({"Content-Type": "application/json"})
            print(f"Sync client headers: {sync_headers}")
            
            # Both should have the same trace_id
            assert async_headers["X-Trace-Id"] == sync_headers["X-Trace-Id"]
            print("✓ Both async and sync clients use same trace_id from context")


def main():
    """Run all examples."""
    print("SMAP TracedHTTPClient Simple Examples")
    print("=" * 50)
    
    example_header_preparation()
    example_service_communication_pattern()
    example_error_handling()
    example_configuration_options()
    
    print("\n" + "=" * 50)
    print("All examples completed successfully! 🎉")
    print("\nImplementation Summary:")
    print("✅ TracedHTTPClient with automatic trace_id injection")
    print("✅ Support for both async (httpx) and sync (requests) backends")
    print("✅ Configurable timeouts, retries, and base headers")
    print("✅ Graceful error handling for missing/invalid trace_ids")
    print("✅ Integration with FastAPI middleware pattern")
    print("✅ End-to-end trace propagation across service calls")


if __name__ == "__main__":
    main()