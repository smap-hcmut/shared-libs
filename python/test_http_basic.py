"""
Basic test for HTTP client implementation without external dependencies.
"""

import sys
import os

# Add the smap_shared package to Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '.'))

def test_basic_imports():
    """Test that we can import the basic tracing components."""
    try:
        from smap_shared.tracing import (
            get_trace_id, 
            set_trace_id, 
            generate_trace_id,
            validate_trace_id,
            http_propagator
        )
        print("✓ Tracing imports successful")
        return True
    except Exception as e:
        print(f"❌ Tracing import failed: {e}")
        return False

def test_trace_context():
    """Test basic trace context functionality."""
    try:
        from smap_shared.tracing import set_trace_id, get_trace_id, generate_trace_id, validate_trace_id
        
        # Test trace_id generation
        trace_id = generate_trace_id()
        assert validate_trace_id(trace_id), f"Generated trace_id is invalid: {trace_id}"
        print(f"✓ Generated valid trace_id: {trace_id}")
        
        # Test trace_id context management
        set_trace_id(trace_id)
        retrieved_trace_id = get_trace_id()
        assert retrieved_trace_id == trace_id, f"Retrieved trace_id doesn't match: {retrieved_trace_id} != {trace_id}"
        print("✓ Trace context management works")
        
        return True
    except Exception as e:
        print(f"❌ Trace context test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

def test_http_propagator():
    """Test HTTP propagator functionality."""
    try:
        from smap_shared.tracing import http_propagator, set_trace_id, generate_trace_id
        
        # Set a trace_id in context
        test_trace_id = generate_trace_id()
        set_trace_id(test_trace_id)
        
        # Test header injection
        headers = {"Content-Type": "application/json"}
        http_propagator.inject_http(headers)
        
        assert "X-Trace-Id" in headers, "X-Trace-Id header not injected"
        assert headers["X-Trace-Id"] == test_trace_id, f"Injected trace_id doesn't match: {headers['X-Trace-Id']} != {test_trace_id}"
        print(f"✓ HTTP header injection works: {headers}")
        
        # Test header extraction
        extracted_trace_id = http_propagator.extract_http(headers)
        assert extracted_trace_id == test_trace_id, f"Extracted trace_id doesn't match: {extracted_trace_id} != {test_trace_id}"
        print("✓ HTTP header extraction works")
        
        return True
    except Exception as e:
        print(f"❌ HTTP propagator test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

def test_http_client_basic():
    """Test basic HTTP client functionality without external dependencies."""
    try:
        # Mock the external dependencies
        import unittest.mock
        
        with unittest.mock.patch('smap_shared.http.client.HTTPX_AVAILABLE', True):
            with unittest.mock.patch('smap_shared.http.client.REQUESTS_AVAILABLE', True):
                from smap_shared.http.client import TracedHTTPClient
                
                # Test client creation
                client = TracedHTTPClient(use_async=True)
                assert client.use_async is True
                assert client.timeout == 30.0
                assert client.retries == 3
                print("✓ HTTP client creation works")
                
                # Test header preparation
                from smap_shared.tracing import set_trace_id, generate_trace_id
                test_trace_id = generate_trace_id()
                set_trace_id(test_trace_id)
                
                headers = client._prepare_headers({"Content-Type": "application/json"})
                assert "X-Trace-Id" in headers, "X-Trace-Id not in prepared headers"
                assert headers["X-Trace-Id"] == test_trace_id, "Trace_id not properly injected"
                assert headers["Content-Type"] == "application/json", "Custom headers not preserved"
                print("✓ HTTP client header preparation works")
                
                return True
    except Exception as e:
        print(f"❌ HTTP client basic test failed: {e}")
        import traceback
        traceback.print_exc()
        return False

def main():
    """Run all basic tests."""
    print("Running basic HTTP client tests...\n")
    
    tests = [
        test_basic_imports,
        test_trace_context,
        test_http_propagator,
        test_http_client_basic,
    ]
    
    passed = 0
    total = len(tests)
    
    for test in tests:
        if test():
            passed += 1
        print()  # Add spacing between tests
    
    print(f"Results: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 All tests passed!")
        return True
    else:
        print("❌ Some tests failed")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)