#!/usr/bin/env python3
"""
Minimal integration test for core tracing functionality.
Tests only the tracing module without dependencies.
"""

import sys
import os

# Add shared library to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'python'))

def test_tracing_only():
    """Test only tracing functionality without other dependencies."""
    print("Testing core tracing functionality...")
    
    # Import only tracing components
    from smap_shared.tracing.context import TraceContext
    from smap_shared.tracing.http import HTTPPropagator
    from smap_shared.tracing.kafka import KafkaPropagator
    
    # Test trace ID generation
    tracer = TraceContext()
    trace_id = tracer.generate_trace_id()
    print(f"Generated trace_id: {trace_id}")
    
    assert len(trace_id) > 0, "Trace ID should not be empty"
    assert tracer.is_valid_trace_id(trace_id), "Generated trace ID should be valid"
    
    # Test HTTP propagation
    http_propagator = HTTPPropagator()
    headers = {}
    http_propagator.inject_http_with_trace_id(headers, trace_id)
    
    assert "X-Trace-Id" in headers, "X-Trace-Id header should be injected"
    assert headers["X-Trace-Id"] == trace_id, "Injected trace ID should match"
    
    extracted_id = http_propagator.extract_http_trace_id(headers)
    assert extracted_id == trace_id, "Extracted trace ID should match"
    
    # Test Kafka propagation
    kafka_propagator = KafkaPropagator()
    kafka_headers = {}
    kafka_propagator.inject_kafka_with_trace_id(kafka_headers, trace_id)
    
    assert "X-Trace-Id" in kafka_headers, "Kafka X-Trace-Id header should be injected"
    
    kafka_extracted_id = kafka_propagator.extract_kafka_trace_id(kafka_headers)
    assert kafka_extracted_id == trace_id, "Kafka extracted trace ID should match"
    
    print("✓ Core tracing functionality working")
    print(f"✓ Trace ID: {trace_id}")
    print(f"✓ HTTP propagation: {headers}")
    print(f"✓ Kafka propagation: {kafka_headers}")

def main():
    """Run minimal integration test."""
    print("🧪 Running Minimal Integration Test")
    print("===================================")
    
    try:
        test_tracing_only()
        
        print("")
        print("✅ Minimal integration test passed!")
        print("🎉 Core tracing functionality is working")
        return 0
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    exit(main())