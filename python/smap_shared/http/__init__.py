"""
HTTP package with automatic trace_id propagation.

HTTP client, middleware, and utilities with X-Trace-Id header management.
"""

from .client import (
    TracedHTTPClient,
    create_async_client,
    create_sync_client,
    traced_httpx_client,
    traced_requests_session,
)

from .middleware import (
    TracingMiddleware,
    HTTPMiddleware,
    create_tracing_middleware,
    setup_tracing_middleware,
    trace_middleware,
)

from .utils import (
    HTTPUtils,
    ServiceClient,
    fetch_json,
    fetch_json_sync,
    post_json_data,
    post_json_data_sync,
)

__all__ = [
    # Client classes and functions
    "TracedHTTPClient",
    "create_async_client",
    "create_sync_client",
    "traced_httpx_client",
    "traced_requests_session",
    
    # Middleware classes and functions
    "TracingMiddleware",
    "HTTPMiddleware",
    "create_tracing_middleware",
    "setup_tracing_middleware",
    "trace_middleware",
    
    # Utility classes and functions
    "HTTPUtils",
    "ServiceClient",
    "fetch_json",
    "fetch_json_sync",
    "post_json_data",
    "post_json_data_sync",
]