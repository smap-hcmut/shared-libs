"""
TraceContext implementation for Python services.

Provides trace_id management using contextvars for thread-safe and async-safe context storage.
"""

import uuid
import re
from typing import Optional
from contextvars import ContextVar

from .interfaces import TraceContextInterface


# Context variable for trace_id storage (thread-safe, async-safe)
_trace_id_var: ContextVar[Optional[str]] = ContextVar("trace_id", default=None)

# Business context variables — used for log enrichment (not tracing)
_project_id_var: ContextVar[Optional[str]] = ContextVar("project_id", default=None)
_campaign_id_var: ContextVar[Optional[str]] = ContextVar("campaign_id", default=None)
_user_id_var: ContextVar[Optional[str]] = ContextVar("user_id", default=None)


class TraceContext(TraceContextInterface):
    """
    TraceContext implementation using contextvars.

    Provides thread-safe and async-safe trace_id management compatible with
    FastAPI, asyncio, and other Python async frameworks.

    Features:
    - UUID v4 generation and validation
    - contextvars integration for automatic context propagation
    - Cross-language compatibility with Go services
    - Graceful error handling and recovery
    """

    # UUID v4 validation regex pattern
    # Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
    # Where x is any hexadecimal digit and y is one of 8, 9, A, or B
    _UUID_V4_PATTERN = re.compile(
        r"^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
    )

    def get_trace_id(self) -> Optional[str]:
        """
        Returns current trace_id from context.

        Returns:
            Current trace_id or None if not set
        """
        return _trace_id_var.get()

    def set_trace_id(self, trace_id: str) -> None:
        """
        Sets trace_id in current context.

        Args:
            trace_id: Trace ID to set in context

        Raises:
            ValueError: If trace_id is not a valid UUID v4
        """
        if not self.validate_trace_id(trace_id):
            raise ValueError(f"Invalid trace_id format: {trace_id}")

        _trace_id_var.set(trace_id)

    def generate_trace_id(self) -> str:
        """
        Creates new UUID v4 trace_id.

        Returns:
            New UUID v4 string in lowercase format
        """
        return str(uuid.uuid4())

    def validate_trace_id(self, trace_id: str) -> bool:
        """
        Checks if trace_id is valid UUID v4.

        Args:
            trace_id: Trace ID to validate

        Returns:
            True if valid UUID v4, False otherwise
        """
        if not trace_id or not isinstance(trace_id, str):
            return False

        # Convert to lowercase for validation (consistent with Go implementation)
        trace_id_lower = trace_id.lower()

        # Check format using regex
        return bool(self._UUID_V4_PATTERN.match(trace_id_lower))

    def is_valid_trace_id(self, trace_id: str) -> bool:
        """
        Checks if a trace_id is valid and non-empty.

        Args:
            trace_id: Trace ID to check

        Returns:
            True if valid and non-empty, False otherwise
        """
        return bool(trace_id) and self.validate_trace_id(trace_id)

    def get_or_generate_trace_id(self) -> str:
        """
        Gets current trace_id or generates a new one if not set.

        Returns:
            Current trace_id or new UUID v4 if none exists
        """
        current_trace_id = self.get_trace_id()
        if current_trace_id and self.validate_trace_id(current_trace_id):
            return current_trace_id

        # Generate new trace_id and set it
        new_trace_id = self.generate_trace_id()
        self.set_trace_id(new_trace_id)
        return new_trace_id

    def clear_trace_id(self) -> None:
        """
        Clears trace_id from current context.
        """
        _trace_id_var.set(None)


# Global instance for convenience
trace_context = TraceContext()


# Convenience functions for direct access
def get_trace_id() -> Optional[str]:
    """Get current trace_id from context."""
    return trace_context.get_trace_id()


def set_trace_id(trace_id: str) -> None:
    """Set trace_id in current context."""
    trace_context.set_trace_id(trace_id)


def generate_trace_id() -> str:
    """Generate new UUID v4 trace_id."""
    return trace_context.generate_trace_id()


def validate_trace_id(trace_id: str) -> bool:
    """Validate if trace_id is valid UUID v4."""
    return trace_context.validate_trace_id(trace_id)


def get_or_generate_trace_id() -> str:
    """Get current trace_id or generate new one."""
    return trace_context.get_or_generate_trace_id()


def clear_trace_id() -> None:
    """Clear trace_id from current context."""
    trace_context.clear_trace_id()


# Convenience functions for business context (log enrichment)


def get_project_id() -> Optional[str]:
    """Get current project_id from context."""
    return _project_id_var.get()


def set_project_id(project_id: str) -> None:
    """Set project_id in current context."""
    _project_id_var.set(project_id)


def clear_project_id() -> None:
    """Clear project_id from current context."""
    _project_id_var.set(None)


def get_campaign_id() -> Optional[str]:
    """Get current campaign_id from context."""
    return _campaign_id_var.get()


def set_campaign_id(campaign_id: str) -> None:
    """Set campaign_id in current context."""
    _campaign_id_var.set(campaign_id)


def clear_campaign_id() -> None:
    """Clear campaign_id from current context."""
    _campaign_id_var.set(None)


def get_user_id() -> Optional[str]:
    """Get current user_id from context."""
    return _user_id_var.get()


def set_user_id(user_id: str) -> None:
    """Set user_id in current context."""
    _user_id_var.set(user_id)


def clear_user_id() -> None:
    """Clear user_id from current context."""
    _user_id_var.set(None)
