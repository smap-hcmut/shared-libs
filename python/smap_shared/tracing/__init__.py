"""
Python tracing library for distributed trace_id management.

Provides comprehensive trace_id management and propagation for Python services
in the SMAP platform. Features include:

- TraceContext: Thread-safe and async-safe trace_id management using contextvars
- HTTPPropagator: HTTP header injection and extraction for trace_id propagation
- KafkaPropagator: Kafka message header management for trace_id propagation
- FastAPI middleware: Ready-to-use middleware for automatic trace_id handling
- Validation utilities: UUID v4 validation consistent with Go services
- Cross-language compatibility: Works seamlessly with Go services

Usage:
    # Basic usage
    from smap_shared.tracing import get_trace_id, set_trace_id, generate_trace_id
    
    # Set trace_id in context
    set_trace_id("550e8400-e29b-41d4-a716-446655440000")
    
    # Get current trace_id
    current_id = get_trace_id()
    
    # HTTP propagation
    from smap_shared.tracing import get_traced_headers
    headers = get_traced_headers({"Content-Type": "application/json"})
    
    # FastAPI middleware
    from smap_shared.tracing import setup_tracing_middleware
    setup_tracing_middleware(app)
"""

# Core interfaces
from .interfaces import (
    TraceContextInterface,
    HTTPPropagatorInterface,
    KafkaPropagatorInterface,
)

# Core implementations
from .context import (
    TraceContext,
    trace_context,
    get_trace_id,
    set_trace_id,
    generate_trace_id,
    validate_trace_id,
    get_or_generate_trace_id,
    clear_trace_id,
)

from .http import (
    HTTPPropagator,
    http_propagator,
    TRACE_ID_HEADER,
    inject_http_headers,
    extract_http_headers,
    get_traced_headers,
)

from .kafka import (
    KafkaPropagator,
    kafka_propagator,
    inject_kafka_headers,
    extract_kafka_headers,
    get_traced_kafka_headers,
    inject_aiokafka_message_headers,
    extract_aiokafka_message_headers,
)

# Validation utilities
from .validation import (
    validate_uuid_v4,
    is_valid_trace_id,
    normalize_trace_id,
    sanitize_trace_id,
    UUID_V4_PATTERN,
)

# FastAPI middleware
from .middleware import (
    TracingMiddleware,
    create_tracing_middleware,
    setup_tracing_middleware,
)

# Factory functions
from .factory import (
    TracingFactory,
    create_tracing_suite,
    create_fastapi_middleware,
)

# Error classes
from .errors import (
    TracingError,
    InvalidTraceIDError,
    TraceContextError,
    PropagationError,
    HTTPPropagationError,
    KafkaPropagationError,
    MiddlewareError,
)

# Version info
__version__ = "1.0.0"

# Public API
__all__ = [
    # Interfaces
    "TraceContextInterface",
    "HTTPPropagatorInterface", 
    "KafkaPropagatorInterface",
    
    # Core classes
    "TraceContext",
    "HTTPPropagator",
    "KafkaPropagator",
    
    # Global instances
    "trace_context",
    "http_propagator",
    "kafka_propagator",
    
    # Context management functions
    "get_trace_id",
    "set_trace_id",
    "generate_trace_id",
    "validate_trace_id",
    "get_or_generate_trace_id",
    "clear_trace_id",
    
    # HTTP propagation functions
    "inject_http_headers",
    "extract_http_headers",
    "get_traced_headers",
    
    # Kafka propagation functions
    "inject_kafka_headers",
    "extract_kafka_headers",
    "get_traced_kafka_headers",
    "inject_aiokafka_message_headers",
    "extract_aiokafka_message_headers",
    
    # Validation functions
    "validate_uuid_v4",
    "is_valid_trace_id",
    "normalize_trace_id",
    "sanitize_trace_id",
    
    # Middleware
    "TracingMiddleware",
    "create_tracing_middleware",
    "setup_tracing_middleware",
    
    # Factory functions
    "TracingFactory",
    "create_tracing_suite",
    "create_fastapi_middleware",
    
    # Error classes
    "TracingError",
    "InvalidTraceIDError",
    "TraceContextError",
    "PropagationError",
    "HTTPPropagationError",
    "KafkaPropagationError",
    "MiddlewareError",
    
    # Constants
    "TRACE_ID_HEADER",
    "UUID_V4_PATTERN",
    "__version__",
]