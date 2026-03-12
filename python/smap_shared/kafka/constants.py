"""
Kafka constants and default values.

Migrated from analysis-srv with enhancements for tracing support.
"""

# Kafka Consumer Defaults
DEFAULT_AUTO_OFFSET_RESET = "earliest"
DEFAULT_ENABLE_AUTO_COMMIT = True
DEFAULT_MAX_POLL_RECORDS = 500
DEFAULT_SESSION_TIMEOUT_MS = 10000

# Kafka Producer Defaults
DEFAULT_ACKS = "all"
DEFAULT_COMPRESSION_TYPE = None
DEFAULT_MAX_BATCH_SIZE = 16384
DEFAULT_LINGER_MS = 0
DEFAULT_ENABLE_IDEMPOTENCE = True

# Tracing Defaults
DEFAULT_ENABLE_TRACE_INJECTION = True
DEFAULT_ENABLE_TRACE_EXTRACTION = True
DEFAULT_AUTO_GENERATE_TRACE_ID = True

# Valid Values
VALID_AUTO_OFFSET_RESET = ["earliest", "latest"]
VALID_ACKS = ["all", "1", "0", 1, 0, -1]
VALID_COMPRESSION_TYPES = ["gzip", "snappy", "lz4", "zstd"]

# Error Messages
ERROR_BOOTSTRAP_SERVERS_EMPTY = "bootstrap_servers cannot be empty"
ERROR_TOPICS_EMPTY = "topics cannot be empty"
ERROR_GROUP_ID_EMPTY = "group_id cannot be empty"
ERROR_INVALID_AUTO_OFFSET_RESET = (
    "auto_offset_reset must be 'earliest' or 'latest', got {value}"
)
ERROR_INVALID_MAX_POLL_RECORDS = "max_poll_records must be positive, got {value}"
ERROR_INVALID_SESSION_TIMEOUT = "session_timeout_ms must be positive, got {value}"
ERROR_INVALID_ACKS = "acks must be 'all', 1, or 0, got {value}"
ERROR_INVALID_COMPRESSION_TYPE = (
    "compression_type must be 'gzip', 'snappy', 'lz4', 'zstd', or None, got {value}"
)
ERROR_INVALID_MAX_BATCH_SIZE = "max_batch_size must be positive, got {value}"
ERROR_INVALID_LINGER_MS = "linger_ms must be non-negative, got {value}"
ERROR_AIOKAFKA_NOT_INSTALLED = (
    "aiokafka is required for Kafka support. Install with: pip install aiokafka"
)

# Trace Header Constants
TRACE_ID_HEADER = "X-Trace-Id"
TRACE_ID_HEADER_LOWER = "x-trace-id"

__all__ = [
    # Consumer defaults
    "DEFAULT_AUTO_OFFSET_RESET",
    "DEFAULT_ENABLE_AUTO_COMMIT", 
    "DEFAULT_MAX_POLL_RECORDS",
    "DEFAULT_SESSION_TIMEOUT_MS",
    
    # Producer defaults
    "DEFAULT_ACKS",
    "DEFAULT_COMPRESSION_TYPE",
    "DEFAULT_MAX_BATCH_SIZE",
    "DEFAULT_LINGER_MS",
    "DEFAULT_ENABLE_IDEMPOTENCE",
    
    # Tracing defaults
    "DEFAULT_ENABLE_TRACE_INJECTION",
    "DEFAULT_ENABLE_TRACE_EXTRACTION",
    "DEFAULT_AUTO_GENERATE_TRACE_ID",
    
    # Valid values
    "VALID_AUTO_OFFSET_RESET",
    "VALID_ACKS",
    "VALID_COMPRESSION_TYPES",
    
    # Error messages
    "ERROR_BOOTSTRAP_SERVERS_EMPTY",
    "ERROR_TOPICS_EMPTY",
    "ERROR_GROUP_ID_EMPTY",
    "ERROR_INVALID_AUTO_OFFSET_RESET",
    "ERROR_INVALID_MAX_POLL_RECORDS",
    "ERROR_INVALID_SESSION_TIMEOUT",
    "ERROR_INVALID_ACKS",
    "ERROR_INVALID_COMPRESSION_TYPE",
    "ERROR_INVALID_MAX_BATCH_SIZE",
    "ERROR_INVALID_LINGER_MS",
    "ERROR_AIOKAFKA_NOT_INSTALLED",
    
    # Trace headers
    "TRACE_ID_HEADER",
    "TRACE_ID_HEADER_LOWER",
]