"""
Core tracing interfaces for Python services.

Defines the contracts for trace_id management and propagation.
"""

from abc import ABC, abstractmethod
from typing import Optional, Dict


class TraceContextInterface(ABC):
    """Interface for trace context management."""

    @abstractmethod
    def get_trace_id(self) -> Optional[str]:
        """Returns current trace_id from context."""
        pass

    @abstractmethod
    def set_trace_id(self, trace_id: str) -> None:
        """Sets trace_id in current context."""
        pass

    @abstractmethod
    def generate_trace_id(self) -> str:
        """Creates new UUID v4 trace_id."""
        pass

    @abstractmethod
    def validate_trace_id(self, trace_id: str) -> bool:
        """Checks if trace_id is valid UUID v4."""
        pass


class HTTPPropagatorInterface(ABC):
    """Interface for HTTP trace_id propagation."""

    @abstractmethod
    def inject_http(self, headers: Dict[str, str]) -> None:
        """Adds trace_id to outbound HTTP request headers."""
        pass

    @abstractmethod
    def extract_http(self, headers: Dict[str, str]) -> Optional[str]:
        """Retrieves trace_id from inbound HTTP request headers."""
        pass


class KafkaPropagatorInterface(ABC):
    """Interface for Kafka trace_id propagation."""

    @abstractmethod
    def inject_kafka(self, headers: Dict[str, str]) -> None:
        """Adds trace_id to Kafka message headers."""
        pass

    @abstractmethod
    def extract_kafka(self, headers: Dict[str, str]) -> Optional[str]:
        """Retrieves trace_id from Kafka message headers."""
        pass