"""
Traced Kafka producer implementation for Python services.

Enhanced Kafka producer with automatic trace_id injection into message headers.
Migrated from analysis-srv with enhanced tracing capabilities.
"""

import json
import logging
from typing import Optional, Dict, List

try:
    from aiokafka import AIOKafkaProducer  # type: ignore
except ImportError:
    raise ImportError(
        "aiokafka is required for Kafka support. Install with: pip install aiokafka"
    )

from ..tracing import TraceContext, KafkaPropagator
from .interfaces import ITracedKafkaProducer
from .config import KafkaProducerConfig, KafkaProducerError


logger = logging.getLogger(__name__)


class TracedKafkaProducer(ITracedKafkaProducer):
    """
    Kafka producer with automatic trace_id injection.

    This class wraps aiokafka.AIOKafkaProducer and automatically injects
    trace_id into message headers for distributed tracing support.

    Features:
    - Automatic trace_id injection into message headers
    - Backward compatibility with existing Kafka producer interfaces
    - Support for JSON serialization with trace headers
    - Batch message sending with trace propagation
    - Graceful error handling and logging
    - Cross-language compatibility with Go services

    Attributes:
        config: Kafka producer configuration
        producer: Active aiokafka producer instance
        trace_context: Trace context manager
        kafka_propagator: Kafka trace propagator
    """

    def __init__(
        self,
        config: KafkaProducerConfig,
        trace_context: Optional[TraceContext] = None,
        kafka_propagator: Optional[KafkaPropagator] = None,
    ):
        """
        Initialize traced Kafka producer with configuration.

        Args:
            config: Kafka producer configuration
            trace_context: TraceContext instance (uses global instance if None)
            kafka_propagator: KafkaPropagator instance (uses global instance if None)
        """
        self.config = config
        self.producer: Optional[AIOKafkaProducer] = None
        self._running = False
        
        # Initialize tracing components
        self.trace_context = trace_context or TraceContext()
        self.kafka_propagator = kafka_propagator or KafkaPropagator(self.trace_context)

    async def start(self) -> None:
        """
        Start the producer and connect to Kafka.

        Creates a producer instance and establishes connection.

        Raises:
            KafkaProducerError: If connection fails
        """
        try:
            # Create producer
            self.producer = AIOKafkaProducer(
                bootstrap_servers=self.config.bootstrap_servers,
                acks=self.config.acks,
                compression_type=self.config.compression_type,
                max_batch_size=self.config.max_batch_size,
                linger_ms=self.config.linger_ms,
                client_id=self.config.client_id,
                enable_idempotence=self.config.enable_idempotence,
            )

            # Start producer
            await self.producer.start()
            self._running = True
            
            logger.info(
                f"Started traced Kafka producer: "
                f"servers={self.config.bootstrap_servers}, "
                f"client_id={self.config.client_id}, "
                f"trace_injection={self.config.enable_trace_injection}"
            )

        except Exception as e:
            logger.error(f"Failed to start traced Kafka producer: {e}")
            logger.exception("Kafka producer start error details:")
            raise KafkaProducerError(f"Failed to start producer: {e}") from e

    async def stop(self) -> None:
        """
        Stop the producer gracefully.

        Flushes pending messages and closes the producer connection.
        """
        try:
            logger.info("Stopping traced Kafka producer...")

            if self.producer:
                # Flush pending messages
                await self.producer.flush()
                await self.producer.stop()
                self._running = False
                logger.info("Traced Kafka producer stopped successfully")

        except Exception as e:
            logger.error(f"Error stopping traced Kafka producer: {e}")
            logger.exception("Kafka producer stop error details:")

    async def send(
        self,
        topic: str,
        value: bytes,
        key: Optional[bytes] = None,
        partition: Optional[int] = None,
        headers: Optional[Dict[str, bytes]] = None,
    ) -> None:
        """
        Send a message to Kafka topic with automatic trace_id injection.

        Args:
            topic: Topic name
            value: Message value (bytes)
            key: Optional message key (bytes)
            partition: Optional partition number
            headers: Optional message headers

        Raises:
            RuntimeError: If producer is not started
            KafkaProducerError: If sending fails
        """
        if not self.producer or not self._running:
            raise RuntimeError("Producer not started. Call start() first.")

        try:
            # Prepare headers with trace_id injection
            final_headers = self._prepare_headers_with_trace(headers)

            # Send message
            await self.producer.send(
                topic=topic,
                value=value,
                key=key,
                partition=partition,
                headers=final_headers,
            )

            # Log successful send (with trace_id if available)
            trace_id = self.trace_context.get_trace_id()
            logger.debug(
                f"Sent message to topic={topic}, partition={partition}, "
                f"key={key[:20] if key else None}, size={len(value)} bytes, "
                f"trace_id={trace_id}"
            )

        except Exception as e:
            logger.error(f"Failed to send message to topic={topic}: {e}")
            raise KafkaProducerError(f"Send failed: {e}") from e

    async def send_json(
        self,
        topic: str,
        value: Dict[str, object],
        key: Optional[str] = None,
        partition: Optional[int] = None,
        headers: Optional[Dict[str, bytes]] = None,
    ) -> None:
        """
        Send a JSON message to Kafka topic with automatic trace_id injection.

        Convenience method that serializes dict to JSON bytes.
        Enhanced from analysis-srv implementation.

        Args:
            topic: Topic name
            value: Message value (dict)
            key: Optional message key (string)
            partition: Optional partition number
            headers: Optional message headers

        Raises:
            RuntimeError: If producer is not started
            KafkaProducerError: If sending fails
        """
        # Serialize value to JSON bytes
        value_bytes = json.dumps(value, ensure_ascii=False).encode("utf-8")

        # Serialize key to bytes if provided
        key_bytes = key.encode("utf-8") if key else None

        await self.send(
            topic=topic,
            value=value_bytes,
            key=key_bytes,
            partition=partition,
            headers=headers,
        )

    async def send_batch(self, messages: List[Dict[str, bytes]]) -> None:
        """
        Send multiple messages in batch with automatic trace_id injection.

        Args:
            messages: List of message dicts with 'topic', 'value', 'key', etc.

        Raises:
            RuntimeError: If producer is not started
            KafkaProducerError: If sending fails
        """
        if not self.producer or not self._running:
            raise RuntimeError("Producer not started. Call start() first.")

        try:
            for msg in messages:
                topic = msg.get("topic")
                value = msg.get("value")
                key = msg.get("key")
                partition = msg.get("partition")
                headers = msg.get("headers")

                if not topic or value is None:
                    logger.warning(f"Skipping invalid message: {msg}")
                    continue

                await self.send(
                    topic=topic,
                    value=value,
                    key=key,
                    partition=partition,
                    headers=headers,
                )

            logger.info(f"Sent batch of {len(messages)} messages with trace propagation")

        except Exception as e:
            logger.error(f"Failed to send batch: {e}")
            raise KafkaProducerError(f"Batch send failed: {e}") from e

    def is_running(self) -> bool:
        """
        Check if producer is running.

        Returns:
            bool: True if running, False otherwise
        """
        return self._running and self.producer is not None

    def _prepare_headers_with_trace(
        self, headers: Optional[Dict[str, bytes]] = None
    ) -> Optional[List[tuple]]:
        """
        Prepare message headers with trace_id injection.

        Args:
            headers: Optional existing headers

        Returns:
            Headers list in aiokafka format with trace_id injected
        """
        if not self.config.enable_trace_injection:
            # Convert headers to aiokafka format without trace injection
            if headers:
                return [(k, v) for k, v in headers.items()]
            return None

        try:
            # Get current trace_id
            trace_id = self.trace_context.get_trace_id()
            
            # Generate trace_id if missing and auto-generation is enabled
            if not trace_id and self.config.auto_generate_trace_id:
                trace_id = self.trace_context.generate_trace_id()
                self.trace_context.set_trace_id(trace_id)
                logger.debug(f"Generated new trace_id for Kafka message: {trace_id}")

            # Prepare headers list
            result_headers = []
            
            # Add existing headers
            if headers:
                result_headers.extend([(k, v) for k, v in headers.items()])
            
            # Add trace_id header if available
            if trace_id:
                result_headers.append(("X-Trace-Id", trace_id.encode("utf-8")))
                logger.debug(f"Injected trace_id into Kafka message headers: {trace_id}")

            return result_headers if result_headers else None

        except Exception as e:
            logger.warning(f"Failed to prepare headers with trace_id: {e}")
            # Return original headers without trace injection
            if headers:
                return [(k, v) for k, v in headers.items()]
            return None


__all__ = [
    "TracedKafkaProducer",
    "KafkaProducerError",
]