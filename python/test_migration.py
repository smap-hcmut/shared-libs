#!/usr/bin/env python3
"""
Test script to verify Python logging package migration is complete.

This script tests:
1. Basic logger functionality
2. Automatic trace_id injection from contextvars
3. Structured logging with trace_id field
4. Backward compatibility with existing log calls
5. Requirements 7.1, 7.2, 7.3, 7.4, 7.5 compliance
"""

import json
import sys
from io import StringIO
from contextlib import redirect_stdout, redirect_stderr

# Test imports
from smap_shared.logger import Logger, LoggerConfig, LogLevel
from smap_shared.logger.compat import setup_logging, trace_context, LoggerCompat
from smap_shared.tracing.context import set_trace_id, get_trace_id, generate_trace_id


def test_basic_functionality():
    """Test basic logger functionality."""
    print("=== Testing Basic Functionality ===")
    
    config = LoggerConfig(
        level=LogLevel.INFO,
        enable_trace_id=True,
        colorize=False,
        service_name="test-service"
    )
    logger = Logger(config)
    
    # Test without trace_id
    logger.info("Message without trace_id")
    
    # Test with trace_id
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    logger.info("Message with trace_id")
    
    assert get_trace_id() == trace_id
    print("✓ Basic functionality works")


def test_automatic_trace_injection():
    """Test automatic trace_id injection from contextvars."""
    print("=== Testing Automatic Trace Injection ===")
    
    config = LoggerConfig(
        level=LogLevel.INFO,
        json_output=True,
        enable_trace_id=True,
        service_name="trace-test"
    )
    logger = Logger(config)
    
    # Capture JSON output
    output = StringIO()
    
    # Set trace_id in context
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    
    # Redirect stdout to capture JSON
    with redirect_stdout(output):
        logger.info("Test message with automatic trace injection")
    
    # Parse JSON output
    json_output = output.getvalue().strip()
    log_data = json.loads(json_output)
    
    # Verify trace_id is included
    assert "trace_id" in log_data
    assert log_data["trace_id"] == trace_id
    assert log_data["message"] == "Test message with automatic trace injection"
    assert log_data["service"] == "trace-test"
    
    print("✓ Automatic trace_id injection works")


def test_structured_logging():
    """Test structured logging with trace_id field."""
    print("=== Testing Structured Logging ===")
    
    config = LoggerConfig(
        level=LogLevel.INFO,
        json_output=True,
        enable_trace_id=True,
        service_name="structured-test"
    )
    logger = Logger(config)
    
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    
    # Capture JSON output
    output = StringIO()
    
    with redirect_stdout(output):
        logger.info(
            "Structured log message",
            extra={
                "user_id": "user123",
                "action": "test_action",
                "duration": 100
            }
        )
    
    # Parse and verify JSON structure
    json_output = output.getvalue().strip()
    log_data = json.loads(json_output)
    
    # Verify required fields
    assert log_data["trace_id"] == trace_id
    assert log_data["message"] == "Structured log message"
    assert log_data["service"] == "structured-test"
    assert "timestamp" in log_data
    assert "level" in log_data
    assert "caller" in log_data
    
    # Verify extra fields
    assert "extra" in log_data
    assert log_data["extra"]["user_id"] == "user123"
    assert log_data["extra"]["action"] == "test_action"
    assert log_data["extra"]["duration"] == 100
    
    print("✓ Structured logging with trace_id works")


def test_backward_compatibility():
    """Test backward compatibility with existing log calls."""
    print("=== Testing Backward Compatibility ===")
    
    # Test scapper-srv style compatibility
    logger = setup_logging(debug=True, service_name="compat-test")
    trace_id = generate_trace_id()
    
    with trace_context(trace_id):
        assert get_trace_id() == trace_id
        logger.info("Scapper-srv style logging")
    
    # Test analysis-srv style compatibility
    config = LoggerConfig(level=LogLevel.INFO, enable_trace_id=True)
    compat_logger = LoggerCompat(config)
    
    with compat_logger.trace_context(trace_id=trace_id, request_id="req_123"):
        compat_logger.info("Analysis-srv style logging")
        assert compat_logger.get_trace_id() == trace_id
        assert compat_logger.get_request_id() == "req_123"
    
    print("✓ Backward compatibility works")


def test_requirements_compliance():
    """Test compliance with requirements 7.1-7.5."""
    print("=== Testing Requirements Compliance ===")
    
    # Requirement 7.1: Include trace_id when exists in context
    config = LoggerConfig(json_output=True, enable_trace_id=True, service_name="req-test")
    logger = Logger(config)
    
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    
    output = StringIO()
    with redirect_stdout(output):
        logger.info("Test message")
    
    log_data = json.loads(output.getvalue().strip())
    assert log_data["trace_id"] == trace_id
    print("✓ Requirement 7.1: trace_id included when exists")
    
    # Requirement 7.2: Structured field format
    assert "trace_id" in log_data
    assert isinstance(log_data["trace_id"], str)
    print("✓ Requirement 7.2: Structured field format")
    
    # Requirement 7.3: Omit trace_id when not in context
    from smap_shared.tracing.context import clear_trace_id
    clear_trace_id()  # Clear trace_id properly
    
    output = StringIO()
    with redirect_stdout(output):
        logger.info("Test without trace_id")
    
    log_data = json.loads(output.getvalue().strip())
    # Should not have trace_id field when not in context
    assert "trace_id" not in log_data or log_data.get("trace_id") == ""
    print("✓ Requirement 7.3: Omit trace_id when not in context")
    
    # Requirement 7.4: Consistent field naming
    set_trace_id(trace_id)
    output = StringIO()
    with redirect_stdout(output):
        logger.info("Field naming test")
    
    log_data = json.loads(output.getvalue().strip())
    assert "trace_id" in log_data  # Consistent field name
    print("✓ Requirement 7.4: Consistent field naming")
    
    # Requirement 7.5: Existing structure unchanged
    required_fields = ["timestamp", "level", "caller", "message", "service"]
    for field in required_fields:
        assert field in log_data
    print("✓ Requirement 7.5: Existing structure unchanged")


def main():
    """Run all migration tests."""
    print("Python Logger Migration Test")
    print("=" * 50)
    
    try:
        test_basic_functionality()
        test_automatic_trace_injection()
        test_structured_logging()
        test_backward_compatibility()
        test_requirements_compliance()
        
        print("\n" + "=" * 50)
        print("✅ ALL TESTS PASSED - Migration Complete!")
        print("\nKey Features Verified:")
        print("- ✓ Automatic trace_id injection from contextvars")
        print("- ✓ Structured logging with trace_id field")
        print("- ✓ Backward compatibility for existing log calls")
        print("- ✓ Requirements 7.1, 7.2, 7.3, 7.4, 7.5 compliance")
        print("- ✓ Enhanced shared library integration")
        
        return True
        
    except Exception as e:
        print(f"\n❌ TEST FAILED: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)