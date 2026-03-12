"""
Usage examples for the enhanced logger with trace integration.

Shows how to migrate from existing service-specific logger implementations
to the shared library with automatic trace_id injection.
"""

from typing import Optional
from smap_shared.logger import Logger, LoggerConfig, LogLevel
from smap_shared.logger.compat import setup_logging, trace_context, LoggerCompat
from smap_shared.tracing.context import set_trace_id, generate_trace_id


def example_basic_usage():
    """Basic logger usage with trace integration."""
    print("=== Basic Usage Example ===")
    
    # Initialize logger
    config = LoggerConfig(
        level=LogLevel.INFO,
        enable_trace_id=True,
        colorize=True,
        service_name="example-service"
    )
    logger = Logger(config)
    
    # Log without trace_id
    logger.info("Starting service")
    
    # Set trace_id and log
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    logger.info("Processing request with trace_id")
    logger.debug("Debug information")
    logger.warning("Warning message")
    logger.error("Error occurred")
    
    print()


def example_production_json():
    """Production-ready JSON logging example."""
    print("=== Production JSON Logging ===")
    
    config = LoggerConfig(
        level=LogLevel.INFO,
        json_output=True,
        colorize=False,
        service_name="production-service",
        enable_trace_id=True
    )
    logger = Logger(config)
    
    # Simulate request processing
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    
    logger.info("Request received", extra={"user_id": "user123", "endpoint": "/api/data"})
    logger.info("Database query executed", extra={"query_time": "45ms", "rows": 150})
    logger.info("Response sent", extra={"status_code": 200, "response_time": "120ms"})
    
    print()


def example_request_tracking():
    """Request ID tracking example."""
    print("=== Request Tracking Example ===")
    
    config = LoggerConfig(
        level=LogLevel.INFO,
        enable_trace_id=True,
        enable_request_id=True,
        service_name="request-service"
    )
    logger = Logger(config)
    
    # Set trace_id for distributed tracing
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    
    # Use request context for request-specific tracking
    with logger.request_context(request_id="req_456"):
        logger.info("Processing user request")
        logger.info("Validating input data")
        logger.info("Request completed successfully")
    
    print()


def example_scapper_srv_migration():
    """Migration example for scapper-srv style logging."""
    print("=== Scapper-srv Migration Example ===")
    
    # Old way (scapper-srv style)
    # from app.logger import setup_logging, trace_context
    
    # New way (using compatibility layer)
    logger = setup_logging(debug=True, service_name="scapper-srv")
    
    # Use trace context (same API as before)
    trace_id = generate_trace_id()
    with trace_context(trace_id):
        logger.info("Scraping job started")
        logger.info("Processing data")
        logger.info("Scraping job completed")
    
    print()


def example_analysis_srv_migration():
    """Migration example for analysis-srv style logging."""
    print("=== Analysis-srv Migration Example ===")
    
    # Old way (analysis-srv style)
    # from pkg.logger import Logger, LoggerConfig
    
    # New way (using compatibility wrapper)
    config = LoggerConfig(
        level=LogLevel.DEBUG,
        enable_trace_id=True,
        service_name="analysis-srv"
    )
    compat_logger = LoggerCompat(config)
    
    # Use same API as before
    trace_id = generate_trace_id()
    with compat_logger.trace_context(trace_id=trace_id, request_id="req_789"):
        compat_logger.info("Analysis job started")
        compat_logger.debug("Loading model")
        compat_logger.info("Analysis completed")
    
    print()


def example_error_handling():
    """Error handling and exception logging example."""
    print("=== Error Handling Example ===")
    
    config = LoggerConfig(level=LogLevel.DEBUG, service_name="error-service")
    logger = Logger(config)
    
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    
    try:
        # Simulate an error
        raise ValueError("Something went wrong")
    except Exception as e:
        logger.exception("An error occurred during processing")
        logger.error(f"Error details: {str(e)}")
    
    print()


def example_structured_logging():
    """Structured logging with additional context."""
    print("=== Structured Logging Example ===")
    
    config = LoggerConfig(
        level=LogLevel.INFO,
        json_output=True,
        service_name="structured-service"
    )
    logger = Logger(config)
    
    trace_id = generate_trace_id()
    set_trace_id(trace_id)
    
    # Log with structured data
    logger.info(
        "User action performed",
        extra={
            "user_id": "user123",
            "action": "create_project",
            "project_id": "proj456",
            "duration_ms": 250,
            "success": True
        }
    )
    
    # Use bind for persistent context
    bound_logger = logger.bind(component="auth", module="jwt")
    bound_logger.info("JWT token validated")
    bound_logger.info("User permissions checked")
    
    print()


if __name__ == "__main__":
    """Run all examples."""
    print("Logger Examples - Enhanced Python Logger with Trace Integration")
    print("=" * 70)
    
    example_basic_usage()
    example_production_json()
    example_request_tracking()
    example_scapper_srv_migration()
    example_analysis_srv_migration()
    example_error_handling()
    example_structured_logging()
    
    print("All examples completed successfully!")
    print("\nMigration Notes:")
    print("1. Replace service-specific logger imports with smap_shared.logger")
    print("2. Use LoggerConfig for configuration instead of hardcoded values")
    print("3. Trace_id is automatically injected when available in context")
    print("4. Use compatibility layer for gradual migration")
    print("5. JSON output is recommended for production environments")