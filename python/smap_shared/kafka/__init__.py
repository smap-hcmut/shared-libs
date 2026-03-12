"""
Kafka package with trace_id propagation via message headers.

Enhanced Kafka producer/consumer with automatic trace header management.
Migrated from analysis-srv with enhanced tracing capabilities.
"""

from .producer import TracedKafkaProducer
from .consumer import TracedKafkaConsumer
from .wrapper import KafkaProducer, KafkaConsumer, IKafkaProducer, IKafkaConsumer
from .config import (
    KafkaConsumerConfig,
    KafkaProducerConfig,
    KafkaMessage,
    KafkaError,
    KafkaConsumerError,
    KafkaProducerError,
)
from .interfaces import ITracedKafkaConsumer, ITracedKafkaProducer
from .constants import (
    DEFAULT_AUTO_OFFSET_RESET,
    DEFAULT_ENABLE_AUTO_COMMIT,
    DEFAULT_MAX_POLL_RECORDS,
    DEFAULT_SESSION_TIMEOUT_MS,
    DEFAULT_ACKS,
    DEFAULT_COMPRESSION_TYPE,
    DEFAULT_MAX_BATCH_SIZE,
    DEFAULT_LINGER_MS,
    DEFAULT_ENABLE_IDEMPOTENCE,
    DEFAULT_ENABLE_TRACE_INJECTION,
    DEFAULT_ENABLE_TRACE_EXTRACTION,
    DEFAULT_AUTO_GENERATE_TRACE_ID,
    TRACE_ID_HEADER,
)

__all__ = [
    # Enhanced traced implementations
    "TracedKafkaProducer",
    "TracedKafkaConsumer",
    
    # Backward compatibility wrappers
    "KafkaProducer",
    "KafkaConsumer",
    "IKafkaProducer",
    "IKafkaConsumer",
    
    # Configuration classes
    "KafkaConsumerConfig",
    "KafkaProducerConfig",
    "KafkaMessage",
    
    # Error classes
    "KafkaError",
    "KafkaConsumerError",
    "KafkaProducerError",
    
    # Interfaces
    "ITracedKafkaConsumer",
    "ITracedKafkaProducer",
    
    # Constants
    "DEFAULT_AUTO_OFFSET_RESET",
    "DEFAULT_ENABLE_AUTO_COMMIT",
    "DEFAULT_MAX_POLL_RECORDS",
    "DEFAULT_SESSION_TIMEOUT_MS",
    "DEFAULT_ACKS",
    "DEFAULT_COMPRESSION_TYPE",
    "DEFAULT_MAX_BATCH_SIZE",
    "DEFAULT_LINGER_MS",
    "DEFAULT_ENABLE_IDEMPOTENCE",
    "DEFAULT_ENABLE_TRACE_INJECTION",
    "DEFAULT_ENABLE_TRACE_EXTRACTION",
    "DEFAULT_AUTO_GENERATE_TRACE_ID",
    "TRACE_ID_HEADER",
]