"""
Backward compatibility layer for existing service logger implementations.

Provides compatibility functions and classes to ease migration from
service-specific logger implementations to the shared library.
"""

import logging
from typing import Optional
from contextlib import contextmanager

from .logger import Logger
from .config import LoggerConfig, LogLevel
from ..tracing.context import get_trace_id, set_trace_id


class InterceptHandler(logging.Handler):
    """
    Intercept standard logging calls and redirect to loguru.
    
    Compatible with scapper-srv's InterceptHandler implementation.
    """
    
    def __init__(self, logger_instance: Logger):
        super().__init__()
        self.logger = logger_instance
    
    def emit(self, record):
        """Emit log record through loguru logger."""
        # Get corresponding level
        try:
            level = record.levelname
        except (ValueError, AttributeError):
            level = record.levelno
        
        # Find caller frame
        frame, depth = logging.currentframe(), 2
        while frame and frame.f_code.co_filename == logging.__file__:
            frame = frame.f_back
            depth += 1
        
        # Log through loguru with proper depth
        self.logger._loguru.opt(depth=depth, exception=record.exc_info).log(
            level, record.getMessage()
        )


def setup_logging(debug: bool = False, service_name: str = "python-service") -> Logger:
    """
    Setup logging with backward compatibility for existing services.
    
    Compatible with scapper-srv's setup_logging function.
    
    Args:
        debug: Enable debug mode (colored console output)
        service_name: Service name for structured logging
        
    Returns:
        Configured Logger instance
    """
    # Create configuration based on debug mode
    config = LoggerConfig(
        level=LogLevel.DEBUG if debug else LogLevel.INFO,
        colorize=debug,
        json_output=not debug,  # JSON for production, colored for debug
        service_name=service_name,
        enable_trace_id=True,
    )
    
    # Create logger instance
    logger = Logger(config)
    
    # Setup standard logging interception
    logging.basicConfig(handlers=[InterceptHandler(logger)], level=0, force=True)
    
    # Intercept common framework loggers
    for name in ["uvicorn", "uvicorn.access", "uvicorn.error", "fastapi"]:
        framework_logger = logging.getLogger(name)
        framework_logger.handlers = [InterceptHandler(logger)]
        framework_logger.propagate = False
    
    logger.info(f"Logging initialized at {config.level.value} level (JSON={config.json_output})")
    
    return logger


# Backward compatibility functions (scapper-srv style)
def get_trace_id_compat() -> Optional[str]:
    """Get trace_id - backward compatibility function."""
    return get_trace_id()


def set_trace_id_compat(trace_id: str) -> None:
    """Set trace_id - backward compatibility function."""
    set_trace_id(trace_id)


@contextmanager
def trace_context(trace_id: Optional[str] = None):
    """
    Trace context manager - backward compatibility.
    
    Compatible with scapper-srv's trace_context function.
    
    Args:
        trace_id: Trace ID to set in context
    """
    from ..tracing.context import _trace_id_var
    
    # Save previous value
    token = _trace_id_var.set(trace_id) if trace_id else None
    
    try:
        yield
    finally:
        if token:
            _trace_id_var.reset(token)


# Analysis-srv style compatibility
class LoggerCompat:
    """
    Compatibility wrapper for analysis-srv Logger interface.
    
    Provides the same interface as analysis-srv's Logger class
    while using the enhanced shared implementation underneath.
    """
    
    def __init__(self, config: LoggerConfig):
        self._logger = Logger(config)
    
    @contextmanager
    def trace_context(self, trace_id: Optional[str] = None, request_id: Optional[str] = None):
        """Context manager for trace and request IDs."""
        from ..tracing.context import _trace_id_var
        
        # Save previous values
        prev_trace_id = get_trace_id()
        prev_request_id = self._logger.get_request_id()
        
        # Set new values
        if trace_id:
            set_trace_id(trace_id)
        if request_id:
            self._logger.set_request_id(request_id)
        
        try:
            yield
        finally:
            # Restore previous values
            if prev_trace_id:
                set_trace_id(prev_trace_id)
            else:
                _trace_id_var.set(None)
            
            if prev_request_id:
                self._logger.set_request_id(prev_request_id)
            else:
                self._logger.set_request_id(None)
    
    def set_trace_id(self, trace_id: str) -> None:
        """Set trace ID in context."""
        set_trace_id(trace_id)
    
    def get_trace_id(self) -> Optional[str]:
        """Get current trace ID."""
        return get_trace_id()
    
    def set_request_id(self, request_id: str) -> None:
        """Set request ID in context."""
        self._logger.set_request_id(request_id)
    
    def get_request_id(self) -> Optional[str]:
        """Get current request ID."""
        return self._logger.get_request_id()
    
    def debug(self, message: str, **kwargs) -> None:
        """Log debug message."""
        self._logger.debug(message, **kwargs)
    
    def info(self, message: str, **kwargs) -> None:
        """Log info message."""
        self._logger.info(message, **kwargs)
    
    def warn(self, message: str, **kwargs) -> None:
        """Log warning message."""
        self._logger.warning(message, **kwargs)
    
    def error(self, message: str, **kwargs) -> None:
        """Log error message."""
        self._logger.error(message, **kwargs)
    
    def exception(self, message: str, **kwargs) -> None:
        """Log exception with traceback."""
        self._logger.exception(message, **kwargs)
    
    def bind(self, **kwargs):
        """Bind context to logger."""
        return self._logger.bind(**kwargs)