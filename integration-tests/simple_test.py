#!/usr/bin/env python3
"""
Simple integration test for core tracing functionality.
Tests basic trace_id generation, propagation, and validation.
"""

import sys
import os

# Add shared library to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'python'))

def test_basic_tracing():
    """Test basic tracing functionality."""
    print("Testing basic tracing functionality...")
    
    # Test trace ID generation
    from smap_shared.tracing.context import TraceContext
    tracer = TraceContext()
    
    trace_id = tracer.generate_trace_id()
    print(f"Generated trace_id: {trace_id}")
    assert len(trace_id) > 0, "Trace ID should not be empty"
    assert tracer.is_valid_trace_id(trace_id), "Generated trace ID should be valid"
    
    # Test trace ID validation
    valid_uuid = "550e8400-e29b-41d4-a716-446655440000"
    invalid_uuid = "invalid-trace-id"
    
    assert tracer.is_valid_trace_id(valid_uuid), "Valid UUID should pass validation"
    assert not tracer.is_valid_trace_id(invalid_uuid), "Invalid UUID should fail validation"
    
    print("✓ Basic tracing functionality working")

def test_http_propagation():
    """Test HTTP trace propagation."""
    print("Testing HTTP trace propagation...")
    
    from smap_shared.tracing.http import HTTPPropagator
    from smap_shared.tracing.context import TraceContext
    
    tracer = TraceContext()
    propagator = HTTPPropagator()
    
    # Generate trace ID
    trace_id = tracer.generate_trace_id()
    
    # Test injection
    headers = {}
    propagator.inject_http_with_trace_id(headers, trace_id)
    
    assert "X-Trace-Id" in headers, "X-Trace-Id header should be injected"
    assert headers["X-Trace-Id"] == trace_id, "Injected trace ID should match original"
    
    # Test extraction
    extracted_trace_id = propagator.extract_http_trace_id(headers)
    assert extracted_trace_id == trace_id, "Extracted trace ID should match original"
    
    print("✓ HTTP trace propagation working")

def test_kafka_propagation():
    """Test Kafka trace propagation."""
    print("Testing Kafka trace propagation...")
    
    from smap_shared.tracing.kafka import KafkaPropagator
    from smap_shared.tracing.context import TraceContext
    
    tracer = TraceContext()
    propagator = KafkaPropagator()
    
    # Generate trace ID
    trace_id = tracer.generate_trace_id()
    
    # Test injection
    headers = {}
    propagator.inject_kafka_with_trace_id(headers, trace_id)
    
    assert "X-Trace-Id" in headers, "X-Trace-Id header should be injected"
    assert headers["X-Trace-Id"] == trace_id.encode(), "Injected trace ID should match original"
    
    # Test extraction
    extracted_trace_id = propagator.extract_kafka_trace_id(headers)
    assert extracted_trace_id == trace_id, "Extracted trace ID should match original"
    
    print("✓ Kafka trace propagation working")

def test_end_to_end_flow():
    """Test end-to-end trace flow."""
    print("Testing end-to-end trace flow...")
    
    from smap_shared.tracing.context import TraceContext
    from smap_shared.tracing.http import HTTPPropagator
    from smap_shared.tracing.kafka import KafkaPropagator
    
    tracer = TraceContext()
    http_propagator = HTTPPropagator()
    kafka_propagator = KafkaPropagator()
    
    # Original trace ID
    original_trace_id = tracer.generate_trace_id()
    
    # HTTP propagation
    http_headers = {}
    http_propagator.inject_http_with_trace_id(http_headers, original_trace_id)
    http_extracted_id = http_propagator.extract_http_trace_id(http_headers)
    
    # Kafka propagation
    kafka_headers = {}
    kafka_propagator.inject_kafka_with_trace_id(kafka_headers, http_extracted_id)
    kafka_extracted_id = kafka_propagator.extract_kafka_trace_id(kafka_headers)
    
    # Verify consistency
    assert original_trace_id == http_extracted_id, "HTTP trace ID should match original"
    assert original_trace_id == kafka_extracted_id, "Kafka trace ID should match original"
    
    print(f"✓ End-to-end trace flow: {original_trace_id}")

def main():
    """Run all simple integration tests."""
    print("🧪 Running Simple Integration Tests")
    print("===================================")
    
    try:
        test_basic_tracing()
        test_http_propagation()
        test_kafka_propagation()
        test_end_to_end_flow()
        
        print("")
        print("✅ All simple integration tests passed!")
        print("🎉 Core tracing functionality is working correctly")
        return 0
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    exit(main())