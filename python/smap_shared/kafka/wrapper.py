"""
Backward compatibility wrappers for existing Kafka implementations.

Provides drop-in replacements for existing analysis-srv Kafka classes
with enhanced tracing capabilities.
"""

import json
import logging
from typing import Optional, Dict, List, Callable, Awaitable

from .producer import TracedKafkaProducer
from .consumer import TracedKafkaConsumer
from .config import (
    KafkaProducerConfig,
    KafkaConsumerConfig,
    KafkaMessage,
    KafkaProducerError,
    KafkaConsumerError,
)


logger = logging.getLogger(__name__)


class KafkaProducer(TracedKafkaProducer):
    """
    Backward compatibility wrapper for analysis-srv KafkaProducer.
    
    Provides the same interface as the original analysis-srv implementation
    but with enhanced tracing capabilities.
    """
    
    def __init__(self, config: KafkaProducerConfig):
        """
        Initialize Kafka producer with configuration.
        
        Args:
            config: Kafka producer configuration
        """
        super().__init__(config)
    
    async def send_json(
        self,
        topic: str,
        value: Dict[str, object],
        key: Optional[str] = None,
        partition: Optional[int] = None,
        headers: Optional[Dict[str, bytes]] = None,
    ) -> None:
        """
        Send a JSON message to Kafka topic.
        
        Backward compatibility method that matches analysis-srv interface.
        
        Args:
            topic: Topic name
            value: Message value (dict)
            key: Optional message key (string)
            partition: Optional partition number
            headers: Optional message headers
        """
        await super().send_json(topic, value, key, partition, headers)


class KafkaConsumer(TracedKafkaConsumer):
    """
    Backward compatibility wrapper for analysis-srv KafkaConsumer.
    
    Provides the same interface as the original analysis-srv implementation
    but with enhanced tracing capabilities.
    """
    
    def __init__(self, config: KafkaConsumerConfig):
        """
        Initialize Kafka consumer with configuration.
        
        Args:
            config: Kafka consumer configuration
        """
        super().__init__(config)


# Legacy interface compatibility
class IKafkaProducer:
    """Legacy interface for backward compatibility."""
    
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
        """Send a message to Kafka topic."""
        ...

    async def send_batch(self, messages: List[Dict[str, bytes]]) -> None:
        """Send multiple messages in batch."""
        ...

    def is_running(self) -> bool:
        """Check if producer is running."""
        ...


class IKafkaConsumer:
    """Legacy interface for backward compatibility."""

    async def start(self) -> None:
        """Start the consumer and connect to Kafka."""
        ...

    async def stop(self) -> None:
        """Stop the consumer gracefully."""
        ...

    async def consume(
        self, message_handler: Callable[[KafkaMessage], Awaitable[None]]
    ) -> None:
        """Start consuming messages from subscribed topics."""
        ...

    async def commit(self) -> None:
        """Manually commit offsets."""
        ...

    def is_running(self) -> bool:
        """Check if consumer is running."""
        ...


__all__ = [
    "KafkaProducer",
    "KafkaConsumer", 
    "IKafkaProducer",
    "IKafkaConsumer",
    "KafkaProducerError",
    "KafkaConsumerError",
]