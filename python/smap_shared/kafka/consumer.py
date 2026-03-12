"""
Traced Kafka consumer implementation for Python services.

Enhanced Kafka consumer with automatic trace_id extraction from message headers.
Migrated from analysis-srv with enhanced tracing capabilities.
"""

import logging
from typing import Callable, Optional, Awaitable

try:
    from aiokafka import AIOKafkaConsumer  # type: ignore
except ImportError:
    raise ImportError(
        "aiokafka is required for Kafka support. Install with: pip install aiokafka"
    )

from ..tracing import TraceContext, KafkaPropagator
from .interfaces import ITracedKafkaConsumer
from .config import KafkaConsumerConfig, KafkaMessage, KafkaConsumerError


logger = logging.getLogger(__name__)


class TracedKafkaConsumer(ITracedKafkaConsumer):
    """
    Kafka consumer with automatic trace_id extraction.

    This class wraps aiokafka.AIOKafkaConsumer and automatically extracts
    trace_id from message headers for distributed tracing support.

    Features:
    - Automatic trace_id extraction from message headers
    - Trace_id context propagation to message handlers
    - Backward compatibility with existing Kafka consumer interfaces
    - Support for manual offset management
    - Graceful error handling and logging
    - Cross-language compatibility with Go services

    Attributes:
        config: Kafka consumer configuration
        consumer: Active aiokafka consumer instance
        trace_context: Trace context manager
        kafka_propagator: Kafka trace propagator
    """

    def __init__(
        self,
        config: KafkaConsumerConfig,
        trace_context: Optional[TraceContext] = None,
        kafka_propagator: Optional[KafkaPropagator] = None,
    ):
        """
        Initialize traced Kafka consumer with configuration.

        Args:
            config: Kafka consumer configuration
            trace_context: TraceContext instance (uses global instance if None)
            kafka_propagator: KafkaPropagator instance (uses global instance if None)
        """
        self.config = config
        self.consumer: Optional[AIOKafkaConsumer] = None
        self._running = False
        
        # Initialize tracing components
        self.trace_context = trace_context or TraceContext()
        self.kafka_propagator = kafka_propagator or KafkaPropagator(self.trace_context)

    async def start(self) -> None:
        """
        Start the consumer and connect to Kafka.

        Creates a consumer instance and subscribes to configured topics.

        Raises:
            KafkaConsumerError: If connection fails
        """
        try:
            # Create consumer
            self.consumer = AIOKafkaConsumer(
                *self.config.topics,
                bootstrap_servers=self.config.bootstrap_servers,
                group_id=self.config.group_id,
                auto_offset_reset=self.config.auto_offset_reset,
                enable_auto_commit=self.config.enable_auto_commit,
                max_poll_records=self.config.max_poll_records,
                session_timeout_ms=self.config.session_timeout_ms,
                client_id=self.config.client_id,
            )

            # Start consumer
            await self.consumer.start()
            self._running = True
            
            logger.info(
                f"Started traced Kafka consumer: "
                f"servers={self.config.bootstrap_servers}, "
                f"group_id={self.config.group_id}, "
                f"topics={self.config.topics}, "
                f"client_id={self.config.client_id}, "
                f"trace_extraction={self.config.enable_trace_extraction}"
            )

        except Exception as e:
            logger.error(f"Failed to start traced Kafka consumer: {e}")
            logger.exception("Kafka consumer start error details:")
            raise KafkaConsumerError(f"Failed to start consumer: {e}") from e

    async def stop(self) -> None:
        """
        Stop the consumer gracefully.

        Closes the consumer connection and cleans up resources.
        """
        try:
            logger.info("Stopping traced Kafka consumer...")

            if self.consumer:
                await self.consumer.stop()
                self._running = False
                logger.info("Traced Kafka consumer stopped successfully")

        except Exception as e:
            logger.error(f"Error stopping traced Kafka consumer: {e}")
            logger.exception("Kafka consumer stop error details:")

    async def consume(
        self, message_handler: Callable[[KafkaMessage], Awaitable[None]]
    ) -> None:
        """
        Start consuming messages from subscribed topics with trace extraction.

        This method runs indefinitely, processing messages as they arrive.
        Automatically extracts trace_id from message headers and sets it in context.

        Args:
            message_handler: Async callable to process incoming messages.
                             Should accept KafkaMessage with trace_id populated.

        Raises:
            RuntimeError: If consumer is not started
            KafkaConsumerError: If consumption fails
        """
        if not self.consumer or not self._running:
            raise RuntimeError("Consumer not started. Call start() first.")

        try:
            async for msg in self.consumer:
                try:
                    # Extract trace_id from message headers
                    trace_id = self._extract_trace_id_from_message(msg)
                    
                    # Set trace_id in context for message processing
                    if trace_id:
                        self.trace_context.set_trace_id(trace_id)
                        logger.debug(f"Set trace_id from Kafka message: {trace_id}")
                    elif self.config.auto_generate_trace_id:
                        # Generate new trace_id if missing and auto-generation is enabled
                        new_trace_id = self.trace_context.generate_trace_id()
                        self.trace_context.set_trace_id(new_trace_id)
                        logger.debug(f"Generated new trace_id for Kafka message: {new_trace_id}")
                        trace_id = new_trace_id

                    # Convert aiokafka message to our KafkaMessage model
                    kafka_msg = KafkaMessage(
                        topic=msg.topic,
                        partition=msg.partition,
                        offset=msg.offset,
                        value=msg.value,
                        key=msg.key,
                        timestamp=msg.timestamp,
                        headers=dict(msg.headers) if msg.headers else {},
                        trace_id=trace_id,  # Include extracted trace_id
                    )

                    # Process message with trace context
                    await message_handler(kafka_msg)

                    logger.debug(
                        f"Processed message: topic={msg.topic}, partition={msg.partition}, "
                        f"offset={msg.offset}, trace_id={trace_id}"
                    )

                except Exception as e:
                    logger.error(
                        f"Error processing message from topic={msg.topic}, "
                        f"partition={msg.partition}, offset={msg.offset}: {e}"
                    )
                    logger.exception("Message processing error details:")
                    # Continue processing other messages
                finally:
                    # Clear trace context after message processing
                    self.trace_context.clear_trace_id()

        except Exception as e:
            logger.error(f"Error during message consumption: {e}")
            logger.exception("Consumption error details:")
            raise KafkaConsumerError(f"Consumption failed: {e}") from e

    async def commit(self) -> None:
        """
        Manually commit offsets.

        Only needed if enable_auto_commit is False.

        Raises:
            RuntimeError: If consumer is not started
        """
        if not self.consumer or not self._running:
            raise RuntimeError("Consumer not started. Call start() first.")

        try:
            await self.consumer.commit()
            logger.debug("Offsets committed successfully")
        except Exception as e:
            logger.error(f"Failed to commit offsets: {e}")
            raise KafkaConsumerError(f"Commit failed: {e}") from e

    def is_running(self) -> bool:
        """
        Check if consumer is running.

        Returns:
            bool: True if running, False otherwise
        """
        return self._running and self.consumer is not None

    def _extract_trace_id_from_message(self, msg) -> Optional[str]:
        """
        Extract trace_id from aiokafka message headers.

        Args:
            msg: aiokafka ConsumerRecord

        Returns:
            Extracted trace_id or None if not found/invalid
        """
        if not self.config.enable_trace_extraction:
            return None

        try:
            if not msg.headers:
                logger.debug("No headers in Kafka message")
                return None

            # Convert aiokafka headers to dict for extraction
            headers_dict = {}
            for key, value in msg.headers:
                if isinstance(value, bytes):
                    headers_dict[key] = value.decode('utf-8')
                else:
                    headers_dict[key] = str(value)

            # Extract trace_id using kafka propagator
            trace_id = self.kafka_propagator.extract_kafka(headers_dict)
            
            if trace_id:
                logger.debug(f"Extracted trace_id from Kafka message: {trace_id}")
                return trace_id
            else:
                logger.debug("No valid trace_id found in Kafka message headers")
                return None

        except Exception as e:
            logger.warning(f"Failed to extract trace_id from Kafka message: {e}")
            return None


__all__ = [
    "TracedKafkaConsumer",
    "KafkaConsumerError",
]