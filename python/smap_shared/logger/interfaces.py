"""
Logger interfaces for Python services with trace integration.
"""

from abc import ABC, abstractmethod
from typing import Any, Optional


class LoggerInterface(ABC):
    """Interface for trace-aware logging with automatic trace_id injection."""

    @abstractmethod
    def debug(self, message: str, **kwargs: Any) -> None:
        """Log debug message with automatic trace_id injection."""
        pass

    @abstractmethod
    def info(self, message: str, **kwargs: Any) -> None:
        """Log info message with automatic trace_id injection."""
        pass

    @abstractmethod
    def warning(self, message: str, **kwargs: Any) -> None:
        """Log warning message with automatic trace_id injection."""
        pass

    @abstractmethod
    def error(self, message: str, **kwargs: Any) -> None:
        """Log error message with automatic trace_id injection."""
        pass

    @abstractmethod
    def critical(self, message: str, **kwargs: Any) -> None:
        """Log critical message with automatic trace_id injection."""
        pass

    @abstractmethod
    def exception(self, message: str, **kwargs: Any) -> None:
        """Log exception with traceback and automatic trace_id injection."""
        pass

    @abstractmethod
    def with_trace(self) -> "LoggerInterface":
        """Returns logger instance with trace_id from current context."""
        pass

    @abstractmethod
    def get_trace_id(self) -> Optional[str]:
        """Get current trace ID from context."""
        pass