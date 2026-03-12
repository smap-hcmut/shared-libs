"""
Logger package with automatic trace_id injection.

Enhanced logging that automatically includes trace_id from context when available.
Migrated from analysis-srv and enhanced with shared tracing integration.

Usage:
    from smap_shared.logger import Logger, LoggerConfig, LogLevel
    from smap_shared.tracing import set_trace_id
    
    # Initialize logger
    config = LoggerConfig(level=LogLevel.INFO, enable_trace_id=True)
    logger = Logger(config)
    
    # Set trace context
    set_trace_id("550e8400-e29b-41d4-a716-446655440000")
    
    # Log with automatic trace_id injection
    logger.info("Processing request")  # Includes trace_id automatically
"""

from .logger import Logger
from .config import LoggerConfig, LogLevel
from .interfaces import LoggerInterface

__all__ = [
    "Logger",
    "LoggerConfig", 
    "LogLevel",
    "LoggerInterface",
]