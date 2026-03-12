"""
FastAPI middleware for automatic trace_id extraction and generation.

Provides middleware for FastAPI applications to automatically handle
X-Trace-Id header extraction from incoming requests and generation
of new trace_ids when missing.
"""

import logging
from typing import Callable, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    from fastapi import Request, Response

try:
    from fastapi import Request, Response
    from fastapi.middleware.base import BaseHTTPMiddleware
    FASTAPI_AVAILABLE = True
except ImportError:
    FASTAPI_AVAILABLE = False
    # Create dummy classes for type hints when FastAPI is not available
    class BaseHTTPMiddleware:
        def __init__(self, app):
            self.app = app
    
    class Request:
        pass
    
    class Response:
        pass

from ..tracing import (
    get_trace_id,
    set_trace_id,
    generate_trace_id,
    validate_trace_id,
    http_propagator,
    HTTPPropagator,
    TraceContext,
    trace_context,
)

logger = logging.getLogger(__name__)


class TracingMiddleware(BaseHTTPMiddleware):
    """
    FastAPI middleware for automatic trace_id management.
    
    Extracts trace_id from X-Trace-Id header in incoming requests or generates
    a new UUID v4 trace_id if missing or invalid. The trace_id is stored in
    the request context and available throughout the request lifecycle.
    
    Features:
    - Automatic trace_id extraction from X-Trace-Id header
    - UUID v4 validation and generation for invalid/missing trace_ids
    - Context storage using contextvars (thread-safe, async-safe)
    - Graceful error handling and logging
    - Compatible with FastAPI async request handling
    
    Usage:
        from fastapi import FastAPI
        from smap_shared.http import TracingMiddleware
        
        app = FastAPI()
        app.add_middleware(TracingMiddleware)
    
    Note:
        Requires FastAPI to be installed. Will raise ImportError if not available.
    """
    
    def __init__(
        self,
        app,
        trace_context: Optional[TraceContext] = None,
        http_propagator: Optional[HTTPPropagator] = None,
        log_trace_extraction: bool = True,
    ):
        """
        Initialize tracing middleware.
        
        Args:
            app: FastAPI application instance
            trace_context: TraceContext instance (uses global if None)
            http_propagator: HTTPPropagator instance (uses global if None)
            log_trace_extraction: Whether to log trace_id extraction events
            
        Raises:
            ImportError: If FastAPI is not available
        """
        if not FASTAPI_AVAILABLE:
            raise ImportError("FastAPI is required for TracingMiddleware. Install with: pip install fastapi")
        
        super().__init__(app)
        self.trace_context = trace_context or trace_context
        self.http_propagator = http_propagator or http_propagator
        self.log_trace_extraction = log_trace_extraction
    
    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        """
        Process incoming request with trace_id management.
        
        Args:
            request: FastAPI Request object
            call_next: Next middleware/handler in chain
            
        Returns:
            Response from downstream handlers
        """
        try:
            # Extract trace_id from request headers
            trace_id = self.http_propagator.extract_fastapi_request(request)
            
            # Validate extracted trace_id
            if trace_id and self.trace_context.validate_trace_id(trace_id):
                # Valid trace_id found
                if self.log_trace_extraction:
                    logger.debug(f"Extracted valid trace_id from request: {trace_id}")
            else:
                # Generate new trace_id for invalid/missing trace_id
                if trace_id and self.log_trace_extraction:
                    logger.warning(f"Invalid trace_id format in request: {trace_id}, generating new one")
                
                trace_id = self.trace_context.generate_trace_id()
                
                if self.log_trace_extraction:
                    logger.debug(f"Generated new trace_id for request: {trace_id}")
            
            # Set trace_id in context for request processing
            self.trace_context.set_trace_id(trace_id)
            
            # Process request with trace_id in context
            response = await call_next(request)
            
            return response
            
        except Exception as e:
            # Log error but don't block request processing
            logger.error(f"Error in tracing middleware: {e}")
            
            # Try to generate fallback trace_id
            try:
                fallback_trace_id = self.trace_context.generate_trace_id()
                self.trace_context.set_trace_id(fallback_trace_id)
                logger.warning(f"Using fallback trace_id due to middleware error: {fallback_trace_id}")
            except Exception as fallback_error:
                logger.error(f"Failed to generate fallback trace_id: {fallback_error}")
            
            # Continue with request processing
            response = await call_next(request)
            return response


class HTTPMiddleware:
    """
    Alternative HTTP middleware implementation for non-FastAPI frameworks.
    
    Provides trace_id management for ASGI applications that don't use FastAPI's
    middleware system. Can be used with Starlette, Django, or other ASGI frameworks.
    """
    
    def __init__(
        self,
        trace_context: Optional[TraceContext] = None,
        http_propagator: Optional[HTTPPropagator] = None,
        log_trace_extraction: bool = True,
    ):
        """
        Initialize HTTP middleware.
        
        Args:
            trace_context: TraceContext instance (uses global if None)
            http_propagator: HTTPPropagator instance (uses global if None)
            log_trace_extraction: Whether to log trace_id extraction events
        """
        self.trace_context = trace_context or trace_context
        self.http_propagator = http_propagator or http_propagator
        self.log_trace_extraction = log_trace_extraction
    
    async def __call__(self, scope, receive, send):
        """
        ASGI middleware callable.
        
        Args:
            scope: ASGI scope dictionary
            receive: ASGI receive callable
            send: ASGI send callable
        """
        if scope["type"] != "http":
            # Not an HTTP request, pass through
            await self.app(scope, receive, send)
            return
        
        try:
            # Extract headers from ASGI scope
            headers = dict(scope.get("headers", []))
            headers_str = {k.decode(): v.decode() for k, v in headers.items()}
            
            # Extract trace_id from headers
            trace_id = self.http_propagator.extract_http(headers_str)
            
            # Validate and generate trace_id if needed
            if trace_id and self.trace_context.validate_trace_id(trace_id):
                if self.log_trace_extraction:
                    logger.debug(f"Extracted valid trace_id from ASGI request: {trace_id}")
            else:
                if trace_id and self.log_trace_extraction:
                    logger.warning(f"Invalid trace_id format in ASGI request: {trace_id}, generating new one")
                
                trace_id = self.trace_context.generate_trace_id()
                
                if self.log_trace_extraction:
                    logger.debug(f"Generated new trace_id for ASGI request: {trace_id}")
            
            # Set trace_id in context
            self.trace_context.set_trace_id(trace_id)
            
        except Exception as e:
            logger.error(f"Error in HTTP middleware: {e}")
            # Generate fallback trace_id
            try:
                fallback_trace_id = self.trace_context.generate_trace_id()
                self.trace_context.set_trace_id(fallback_trace_id)
                logger.warning(f"Using fallback trace_id due to middleware error: {fallback_trace_id}")
            except Exception as fallback_error:
                logger.error(f"Failed to generate fallback trace_id: {fallback_error}")
        
        # Continue with request processing
        await self.app(scope, receive, send)


# Convenience functions for middleware setup
def create_tracing_middleware(
    trace_context: Optional[TraceContext] = None,
    http_propagator: Optional[HTTPPropagator] = None,
    log_trace_extraction: bool = True,
) -> TracingMiddleware:
    """
    Create FastAPI tracing middleware with custom configuration.
    
    Args:
        trace_context: TraceContext instance (uses global if None)
        http_propagator: HTTPPropagator instance (uses global if None)
        log_trace_extraction: Whether to log trace_id extraction events
        
    Returns:
        Configured TracingMiddleware instance
    """
    return TracingMiddleware(
        app=None,  # Will be set by FastAPI
        trace_context=trace_context,
        http_propagator=http_propagator,
        log_trace_extraction=log_trace_extraction,
    )


def setup_tracing_middleware(
    app,
    trace_context: Optional[TraceContext] = None,
    http_propagator: Optional[HTTPPropagator] = None,
    log_trace_extraction: bool = True,
):
    """
    Add tracing middleware to FastAPI application.
    
    Args:
        app: FastAPI application instance
        trace_context: TraceContext instance (uses global if None)
        http_propagator: HTTPPropagator instance (uses global if None)
        log_trace_extraction: Whether to log trace_id extraction events
        
    Raises:
        ImportError: If FastAPI is not available
    """
    if not FASTAPI_AVAILABLE:
        raise ImportError("FastAPI is required for setup_tracing_middleware. Install with: pip install fastapi")
    
    app.add_middleware(
        TracingMiddleware,
        trace_context=trace_context,
        http_propagator=http_propagator,
        log_trace_extraction=log_trace_extraction,
    )


# Legacy function name for backward compatibility
def trace_middleware(request: Request, call_next: Callable) -> Response:
    """
    Legacy function-based middleware for trace_id management.
    
    Note: This is deprecated. Use TracingMiddleware class instead.
    
    Args:
        request: FastAPI Request object
        call_next: Next middleware/handler in chain
        
    Returns:
        Response from downstream handlers
    """
    import warnings
    warnings.warn(
        "trace_middleware function is deprecated. Use TracingMiddleware class instead.",
        DeprecationWarning,
        stacklevel=2
    )
    
    # Create temporary middleware instance
    middleware = TracingMiddleware(app=None)
    return middleware.dispatch(request, call_next)