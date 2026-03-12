#!/usr/bin/env python3
"""
Direct test for tracing components without going through main package.
"""

import sys
import os

# Add shared library to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'python'))

def test_direct_tracing():
    """Test tracing components directly."""
    print("Testing tracing components directly...")
    
    # Import tracing components directly
    sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'python', 'smap_shared'))
    
    from tracing.context import TraceContext
    from tracing.http import HTTPPropagator
    from tracing.kafka import KafkaPropagator
    
    # Test trace ID generation
    tracer = TraceContext()
    trace_id = tracer.generate_trace_id()
    print(f"Generated trace_id: {trace_id}")
    
    assert len(trace_id) > 0, "Trace ID should not be empty"
    assert tracer.is_valid_trace_id(trace_id), "Generated trace ID should be valid"
    
    # Test HTTP propagation
    http_propagator = HTTPPropagator()
    headers = {}
    
    # Set trace ID in context first
    tracer.set_trace_id(trace_id)
    
    # Inject headers
    http_propagator.inject_http(headers)
    
    assert "X-Trace-Id" in headers, "X-Trace-Id header should be injected"
    assert headers["X-Trace-Id"] == trace_id, "Injected trace ID should match"
    
    # Extract headers
    extracted_id = http_propagator.extract_http(headers)
    assert extracted_id == trace_id, "Extracted trace ID should match"
    
    # Test Kafka propagation
    kafka_propagator = KafkaPropagator()
    kafka_headers = {}
    
    # Inject Kafka headers
    kafka_propagator.inject_kafka(kafka_headers)
    
    assert "X-Trace-Id" in kafka_headers, "Kafka X-Trace-Id header should be injected"
    
    # Extract Kafka headers
    kafka_extracted_id = kafka_propagator.extract_kafka(kafka_headers)
    assert kafka_extracted_id == trace_id, "Kafka extracted trace ID should match"
    
    print("✓ Direct tracing test passed")
    print(f"✓ Trace ID: {trace_id}")
    print(f"✓ HTTP headers: {headers}")
    print(f"✓ Kafka headers: {kafka_headers}")

def main():
    """Run direct integration test."""
    print("🧪 Running Direct Tracing Test")
    print("==============================")
    
    try:
        test_direct_tracing()
        
        print("")
        print("✅ Direct tracing test passed!")
        print("🎉 Core tracing functionality verified")
        return 0
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    exit(main())