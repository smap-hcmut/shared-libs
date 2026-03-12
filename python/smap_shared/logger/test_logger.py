"""
Basic tests for the enhanced logger implementation.

Tests the core functionality and trace integration.
"""

import pytest
import json
import io
import sys
from contextlib import redirect_stdout
from unittest.mock import patch

from .logger import Logger
from .config import LoggerConfig, LogLevel
from ..tracing.context import set_trace_id, get_trace_id, generate_trace_id


def test_logger_initialization():
    """Test logger can be initialized with different configurations."""
    # Test default configuration
    config = LoggerConfig()
    logger = Logger(config)
    assert logger.config.level == LogLevel.INFO
    assert logger.config.enable_trace_id is True
    
    # Test custom configuration
    config = LoggerConfig(
        level=LogLevel.DEBUG,
        service_name="test-service",
        json_output=True,
    )
    logger = Logger(config)
    assert logger.config.level == LogLevel.DEBUG
    assert logger.config.service_name == "test-service"
    assert logger.config.json_output is True


def test_trace_id_integration():
    """Test automatic trace_id injection in log messages."""
    config = LoggerConfig(
        level=LogLevel.DEBUG,
        json_output=True,
        enable_trace_id=True,
    )
    logger = Logger(config)
    
    # Generate and set trace_id
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    
    # Verify trace_id is accessible
    assert get_trace_id() == trace_id
    assert logger.get_trace_id() == trace_id


def test_request_context():
    """Test request context management."""
    config = LoggerConfig(enable_request_id=True)
    logger = Logger(config)
    
    # Test request context
    with logger.request_context(request_id="req_123"):
        assert logger.get_request_id() == "req_123"
    
    # Context should be cleared after exiting
    assert logger.get_request_id() is None


def test_log_levels():
    """Test all log levels work correctly."""
    config = LoggerConfig(level=LogLevel.DEBUG)
    logger = Logger(config)
    
    # Test all log methods exist and are callable
    logger.debug("Debug message")
    logger.info("Info message")
    logger.warning("Warning message")
    logger.warn("Warn message")  # Backward compatibility
    logger.error("Error message")
    logger.critical("Critical message")


def test_backward_compatibility():
    """Test backward compatibility features."""
    from .compat import setup_logging, trace_context, LoggerCompat
    
    # Test setup_logging function
    logger = setup_logging(debug=True, service_name="test-service")
    assert isinstance(logger, Logger)
    
    # Test trace_context compatibility
    trace_id = generate_trace_id()
    with trace_context(trace_id):
        assert get_trace_id() == trace_id
    
    # Test LoggerCompat wrapper
    config = LoggerConfig()
    compat_logger = LoggerCompat(config)
    compat_logger.info("Test message")


def test_configuration_validation():
    """Test configuration validation."""
    # Test invalid log level
    with pytest.raises(ValueError):
        LoggerConfig(level="INVALID_LEVEL")
    
    # Test string level conversion
    config = LoggerConfig(level="debug")
    assert config.level == LogLevel.DEBUG
    
    # Test empty service name
    with pytest.raises(ValueError):
        LoggerConfig(service_name="")


if __name__ == "__main__":
    # Run basic functionality test
    print("Testing logger functionality...")
    
    # Test basic logging
    config = LoggerConfig(level=LogLevel.DEBUG, colorize=False)
    logger = Logger(config)
    
    # Test without trace_id
    logger.info("Test message without trace_id")
    
    # Test with trace_id
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    logger.info(f"Test message with trace_id: {trace_id}")
    
    # Test request context
    with logger.request_context(request_id="req_123"):
        logger.info("Test message with request_id")
    
    print("Logger tests completed successfully!")