"""
Example usage of TracedHTTPClient for SMAP services.

This example demonstrates how to use the TracedHTTPClient in typical
service-to-service communication scenarios.
"""

import asyncio
import sys
import os

# Add the smap_shared package to Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '.'))

from smap_shared.tracing import set_trace_id, get_trace_id, generate_trace_id
from smap_shared.http import TracedHTTPClient, create_async_client


async def example_async_service_call():
    """Example of async service-to-service HTTP call with trace propagation."""
    print("=== Async Service Call Example ===")
    
    # Simulate incoming request with trace_id (e.g., from FastAPI middleware)
    incoming_trace_id = generate_trace_id()
    set_trace_id(incoming_trace_id)
    print(f"Incoming request trace_id: {incoming_trace_id}")
    
    # Create HTTP client with trace injection
    async with create_async_client(
        timeout=30.0,
        base_headers={"Service-Name": "analysis-srv", "Version": "1.0"}
    ) as client:
        
        # Simulate call to project-srv
        print("\nMaking HTTP call to project-srv...")
        
        # The client will automatically inject the trace_id
        headers = client._prepare_headers({
            "Content-Type": "application/json",
            "Authorization": "Bearer service-token"
        })
        
        print(f"Outbound headers: {headers}")
        
        # Verify trace_id propagation
        assert headers["X-Trace-Id"] == incoming_trace_id
        print("✓ Trace_id properly propagated to outbound request")
        
        # In a real scenario, you would make the actual HTTP call:
        # response_body, status_code = await client.get(
        #     "http://project-srv:8080/api/projects",
        #     headers={"Authorization": "Bearer token"}
        # )


def example_sync_service_call():
    """Example of sync service-to-service HTTP call with trace propagation."""
    print("\n=== Sync Service Call Example ===")
    
    # Simulate incoming request with trace_id
    incoming_trace_id = generate_trace_id()
    set_trace_id(incoming_trace_id)
    print(f"Incoming request trace_id: {incoming_trace_id}")
    
    # Create sync HTTP client
    client = TracedHTTPClient(
        use_async=False,
        timeout=30.0,
        base_headers={"Service-Name": "scapper-srv"}
    )
    
    try:
        # Prepare headers for outbound call
        headers = client._prepare_headers({
            "Content-Type": "application/json"
        })
        
        print(f"Outbound headers: {headers}")
        
        # Verify trace_id propagation
        assert headers["X-Trace-Id"] == incoming_trace_id
        print("✓ Trace_id properly propagated to sync request")
        
        # In a real scenario:
        # response_body, status_code = client.get_sync(
        #     "http://identity-srv:8080/api/validate",
        #     headers={"Authorization": "Bearer token"}
        # )
        
    finally:
        client.close()


def example_middleware_integration():
    """Example of how the HTTP client integrates with FastAPI middleware."""
    print("\n=== Middleware Integration Example ===")
    
    # This simulates what happens in FastAPI middleware:
    
    # 1. Extract trace_id from incoming request headers
    incoming_headers = {
        "X-Trace-Id": "550e8400-e29b-41d4-a716-446655440000",
        "Content-Type": "application/json",
        "User-Agent": "SMAP-Client/1.0"
    }
    
    from smap_shared.tracing import http_propagator
    extracted_trace_id = http_propagator.extract_http(incoming_headers)
    print(f"Extracted trace_id from incoming request: {extracted_trace_id}")
    
    # 2. Set trace_id in context (done by middleware)
    set_trace_id(extracted_trace_id)
    
    # 3. Business logic makes outbound HTTP calls
    client = TracedHTTPClient(use_async=True)
    headers = client._prepare_headers({"Accept": "application/json"})
    
    print(f"Outbound request headers: {headers}")
    
    # 4. Verify end-to-end trace propagation
    assert headers["X-Trace-Id"] == extracted_trace_id
    print("✓ End-to-end trace propagation works!")


def example_error_scenarios():
    """Example of error handling and fallback scenarios."""
    print("\n=== Error Handling Examples ===")
    
    # Scenario 1: No trace_id in context
    from smap_shared.tracing import clear_trace_id
    clear_trace_id()
    
    client = TracedHTTPClient(use_async=True)
    headers = client._prepare_headers({"Content-Type": "application/json"})
    
    print(f"Headers without trace_id in context: {headers}")
    print("✓ Client handles missing trace_id gracefully")
    
    # Scenario 2: Invalid trace_id in incoming request
    invalid_headers = {"X-Trace-Id": "invalid-trace-id"}
    
    from smap_shared.tracing import http_propagator
    extracted = http_propagator.extract_http(invalid_headers)
    
    if extracted is None:
        print("✓ Invalid trace_id properly rejected")
        # Middleware would generate new trace_id here
        new_trace_id = generate_trace_id()
        set_trace_id(new_trace_id)
        print(f"Generated new trace_id: {new_trace_id}")


async def main():
    """Run all examples."""
    print("SMAP TracedHTTPClient Usage Examples")
    print("=" * 50)
    
    await example_async_service_call()
    example_sync_service_call()
    example_middleware_integration()
    example_error_scenarios()
    
    print("\n" + "=" * 50)
    print("All examples completed successfully! 🎉")
    print("\nKey benefits:")
    print("- Automatic trace_id injection in all HTTP requests")
    print("- Compatible with both async (httpx) and sync (requests) backends")
    print("- Seamless integration with FastAPI middleware")
    print("- Graceful error handling and fallback scenarios")
    print("- Cross-service trace propagation for end-to-end tracking")


if __name__ == "__main__":
    asyncio.run(main())