"""
Enhanced logger implementation with automatic trace_id injection.

Migrated from analysis-srv and enhanced with automatic trace integration
from the shared tracing library.
"""

import sys
import json
import email.utils
import os
from datetime import timezone, timedelta
from typing import Optional, Dict, Any
from contextvars import ContextVar
from contextlib import contextmanager
from pathlib import Path

try:
    from loguru import logger as _loguru_logger
except ImportError:
    raise ImportError(
        "loguru is required for the logger. Install with: pip install loguru"
    )

from .interfaces import LoggerInterface
from .config import LoggerConfig, LogLevel
from .constants import (
    TRACE_ID_KEY,
    REQUEST_ID_KEY,
    LOG_FORMAT_TIME,
    LOG_FORMAT_LEVEL,
    LOG_FORMAT_TRACE,
    LOG_FORMAT_REQUEST,
    DEFAULT_TRACE_ID_PLACEHOLDER,
    DEFAULT_REQUEST_ID_PLACEHOLDER,
)

# Import trace context from shared tracing library
try:
    from ..tracing.context import get_trace_id, set_trace_id, validate_trace_id
except ImportError:
    # Fallback if tracing is not available
    def get_trace_id() -> Optional[str]:
        return None
    
    def set_trace_id(trace_id: str) -> None:
        pass
    
    def validate_trace_id(trace_id: str) -> bool:
        return False


# Request ID context variable (separate from trace_id)
_request_id_var: ContextVar[Optional[str]] = ContextVar(REQUEST_ID_KEY, default=None)


class Logger(LoggerInterface):
    """
    Enhanced logger with automatic trace_id injection.
    
    Features:
    - Automatic trace_id injection from shared tracing context
    - Structured logging with JSON output for production
    - Colored console output for development
    - Backward compatibility with existing log calls
    - Request ID tracking (optional)
    - Relative path display for better readability
    
    Usage:
        # Initialize logger
        config = LoggerConfig(level=LogLevel.INFO, enable_trace_id=True)
        logger = Logger(config)
        
        # Use with trace context (automatic injection)
        from smap_shared.tracing import set_trace_id
        set_trace_id("550e8400-e29b-41d4-a716-446655440000")
        logger.info("Processing request")  # Automatically includes trace_id
        
        # Use with request context
        with logger.request_context(request_id="req_123"):
            logger.info("Handling request")  # Includes both trace_id and request_id
    """
    
    def __init__(self, config: LoggerConfig):
        """
        Initialize logger with configuration.
        
        Args:
            config: Logger configuration
        """
        self.config = config
        self._loguru = _loguru_logger
        
        # Cache workspace root for relative path computation
        self.workspace_root = Path.cwd()
        self._path_cache: Dict[str, str] = {}
        
        # Remove default loguru handler
        self._loguru.remove()
        
        # Add console handler based on configuration
        if self.config.enable_console:
            self._add_console_handler()
    
    def _add_console_handler(self) -> None:
        """Add console handler with appropriate formatting."""
        
        if self.config.json_output:
            # Production mode: JSON structured output
            self._add_json_handler()
        else:
            # Development mode: Colored console output
            self._add_colored_handler()
    
    def _add_json_handler(self) -> None:
        """Add JSON structured output handler for production."""
        
        # Get timezone (ICT as per original implementation)
        ict_tz = timezone(timedelta(hours=7))
        
        def json_sink(message):
            """Custom JSON sink with trace_id integration."""
            record = message.record
            dt = record["time"].astimezone(ict_tz)
            
            # Build log dictionary
            log_dict = {
                "timestamp": email.utils.format_datetime(dt),
                "level": record["level"].name.lower(),
                "caller": f"{record['file'].name}:{record['line']}",
                "message": record["message"],
                "service": self.config.service_name,
            }
            
            # Add trace_id if available and enabled
            if self.config.enable_trace_id:
                trace_id = get_trace_id()
                if trace_id:
                    log_dict["trace_id"] = trace_id
            
            # Add request_id if available and enabled
            if self.config.enable_request_id:
                request_id = _request_id_var.get()
                if request_id:
                    log_dict["request_id"] = request_id
            
            # Include extra fields from record
            for key, value in record["extra"].items():
                if key not in log_dict:
                    log_dict[key] = value
            
            print(json.dumps(log_dict), flush=True)
        
        self._loguru.add(
            json_sink,
            level=self.config.level.value,
        )
    
    def _add_colored_handler(self) -> None:
        """Add colored console handler for development."""
        
        # Capture instance variables for closure
        workspace_root = self.workspace_root
        path_cache = self._path_cache
        config = self.config
        
        def format_record(record):
            """Format record with trace_id and relative path."""
            
            # Add trace_id to extra if enabled
            if config.enable_trace_id:
                trace_id = get_trace_id()
                record["extra"][TRACE_ID_KEY] = trace_id or DEFAULT_TRACE_ID_PLACEHOLDER
            
            # Add request_id to extra if enabled
            if config.enable_request_id:
                request_id = _request_id_var.get()
                record["extra"][REQUEST_ID_KEY] = request_id or DEFAULT_REQUEST_ID_PLACEHOLDER
            
            # Compute relative path (with caching)
            abs_path = record["file"].path
            if abs_path not in path_cache:
                try:
                    file_path = Path(abs_path)
                    relative_path = file_path.relative_to(workspace_root)
                    path_cache[abs_path] = str(relative_path)
                except (ValueError, AttributeError):
                    # Fallback to filename if relative path computation fails
                    path_cache[abs_path] = record["file"].name
            
            record["extra"]["relative_path"] = path_cache[abs_path]
            return True
        
        # Build format string
        format_str = f"{LOG_FORMAT_TIME} | {LOG_FORMAT_LEVEL}"
        
        # Add trace_id if enabled
        if self.config.enable_trace_id:
            format_str += f" | {LOG_FORMAT_TRACE}"
        
        # Add request_id if enabled
        if self.config.enable_request_id:
            format_str += f" | {LOG_FORMAT_REQUEST}"
        
        # Add location and message
        format_str += " | <cyan>{extra[relative_path]}</cyan>:<cyan>{line}</cyan> - {message}"
        
        self._loguru.add(
            sys.stdout,
            colorize=self.config.colorize,
            format=format_str,
            level=self.config.level.value,
            filter=format_record,
        )
    
    @contextmanager
    def request_context(self, request_id: Optional[str] = None):
        """
        Context manager for request ID tracking.
        
        Args:
            request_id: Request ID to set in context
            
        Usage:
            with logger.request_context(request_id="req_123"):
                logger.info("Processing request")  # Includes request_id
        """
        # Save previous value
        prev_request_id = _request_id_var.get()
        
        # Set new value
        if request_id:
            _request_id_var.set(request_id)
        
        try:
            yield
        finally:
            # Restore previous value
            _request_id_var.set(prev_request_id)
    
    def set_request_id(self, request_id: str) -> None:
        """
        Set request ID for current context.
        
        Args:
            request_id: Request ID to set
        """
        _request_id_var.set(request_id)
    
    def get_request_id(self) -> Optional[str]:
        """
        Get current request ID.
        
        Returns:
            Current request ID or None
        """
        return _request_id_var.get()
    
    def get_trace_id(self) -> Optional[str]:
        """
        Get current trace ID from shared tracing context.
        
        Returns:
            Current trace ID or None
        """
        return get_trace_id()
    
    def debug(self, message: str, **kwargs: Any) -> None:
        """Log debug message with automatic trace_id injection."""
        self._loguru.opt(depth=1).debug(message, **kwargs)
    
    def info(self, message: str, **kwargs: Any) -> None:
        """Log info message with automatic trace_id injection."""
        self._loguru.opt(depth=1).info(message, **kwargs)
    
    def warning(self, message: str, **kwargs: Any) -> None:
        """Log warning message with automatic trace_id injection."""
        self._loguru.opt(depth=1).warning(message, **kwargs)
    
    def warn(self, message: str, **kwargs: Any) -> None:
        """Alias for warning() for backward compatibility."""
        self.warning(message, **kwargs)
    
    def error(self, message: str, **kwargs: Any) -> None:
        """Log error message with automatic trace_id injection."""
        self._loguru.opt(depth=1).error(message, **kwargs)
    
    def critical(self, message: str, **kwargs: Any) -> None:
        """Log critical message with automatic trace_id injection."""
        self._loguru.opt(depth=1).critical(message, **kwargs)
    
    def exception(self, message: str, **kwargs: Any) -> None:
        """Log exception with traceback and automatic trace_id injection."""
        self._loguru.opt(depth=1).exception(message, **kwargs)
    
    def with_trace(self) -> "Logger":
        """
        Returns logger instance with trace_id from current context.
        
        Note: This implementation automatically includes trace_id in all log calls
        when available, so this method returns self for compatibility.
        
        Returns:
            Self (logger automatically includes trace_id)
        """
        return self
    
    def bind(self, **kwargs: Any):
        """
        Bind additional context to logger.
        
        Args:
            **kwargs: Additional context to bind
            
        Returns:
            Bound loguru logger instance
        """
        return self._loguru.bind(**kwargs)


# Convenience functions for backward compatibility
def get_request_id() -> Optional[str]:
    """Get current request ID from context."""
    return _request_id_var.get()


def set_request_id(request_id: str) -> None:
    """Set request ID in current context."""
    _request_id_var.set(request_id)


@contextmanager
def request_context(request_id: Optional[str] = None):
    """Context manager for request ID tracking."""
    prev_request_id = _request_id_var.get()
    
    if request_id:
        _request_id_var.set(request_id)
    
    try:
        yield
    finally:
        _request_id_var.set(prev_request_id)