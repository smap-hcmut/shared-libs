"""
Interfaces for Kafka operations with trace support.

Defines the contracts for Kafka producer and consumer with automatic trace_id propagation.
"""

from typing import Callable, Dict, List, Optional, Awaitable, Protocol, runtime_checkable

from .config import KafkaMessage


@runtime_checkable
class ITracedKafkaConsumer(Protocol):
    """Protocol for traced Kafka consumer operations."""

    async def start(self) -> None:
        """Start the consumer and connect to Kafka."""
        ...

    async def stop(self) -> None:
        """Stop the consumer gracefully."""
        ...

    async def consume(
        self, message_handler: Callable[[KafkaMessage], Awaitable[None]]
    ) -> None:
        """Start consuming messages from subscribed topics with trace extraction."""
        ...

    async def commit(self) -> None:
        """Manually commit offsets."""
        ...

    def is_running(self) -> bool:
        """Check if consumer is running."""
        ...


@runtime_checkable
class ITracedKafkaProducer(Protocol):
    """Protocol for traced Kafka producer operations."""

    async def start(self) -> None:
        """Start the producer and connect to Kafka."""
        ...

    async def stop(self) -> None:
        """Stop the producer gracefully."""
        ...

    async def send(
        self,
        topic: str,
        value: bytes,
        key: Optional[bytes] = None,
        partition: Optional[int] = None,
        headers: Optional[Dict[str, bytes]] = None,
    ) -> None:
        """Send a message to Kafka topic with trace_id injection."""
        ...

    async def send_json(
        self,
        topic: str,
        value: Dict[str, object],
        key: Optional[str] = None,
        partition: Optional[int] = None,
        headers: Optional[Dict[str, bytes]] = None,
    ) -> None:
        """Send a JSON message to Kafka topic with trace_id injection."""
        ...

    async def send_batch(self, messages: List[Dict[str, bytes]]) -> None:
        """Send multiple messages in batch with trace_id injection."""
        ...

    def is_running(self) -> bool:
        """Check if producer is running."""
        ...


__all__ = ["ITracedKafkaConsumer", "ITracedKafkaProducer"]