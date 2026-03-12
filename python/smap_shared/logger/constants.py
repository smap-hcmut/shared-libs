"""
Logger constants and format strings.
"""

# Context variable keys
TRACE_ID_KEY = "trace_id"
REQUEST_ID_KEY = "request_id"

# Time format
LOG_TIME_FORMAT = "YYYY-MM-DD HH:mm:ss"

# Format strings for colored console output
LOG_FORMAT_TIME = "<green>{time:YYYY-MM-DD HH:mm:ss}</green>"
LOG_FORMAT_LEVEL = "<level>{level: <8}</level>"
LOG_FORMAT_TRACE = "<cyan>{extra[trace_id]: <36}</cyan>"
LOG_FORMAT_REQUEST = "<yellow>{extra[request_id]: <16}</yellow>"
LOG_FORMAT_LOCATION = "<cyan>{extra[relative_path]}</cyan>:<cyan>{line}</cyan>"
LOG_FORMAT_MESSAGE = "<level>{message}</level>"

# Default values
DEFAULT_TRACE_ID_PLACEHOLDER = "-"
DEFAULT_REQUEST_ID_PLACEHOLDER = "-"

# Header names for HTTP propagation
HTTP_TRACE_ID_HEADER = "X-Trace-Id"
HTTP_REQUEST_ID_HEADER = "X-Request-Id"

# Kafka header names
KAFKA_TRACE_ID_HEADER = "X-Trace-Id"
KAFKA_REQUEST_ID_HEADER = "X-Request-Id"