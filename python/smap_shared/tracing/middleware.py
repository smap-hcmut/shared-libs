"""
FastAPI middleware for trace_id management.

Provides ready-to-use middleware for automatic trace_id extraction and context management.
"""

import logging
from typing import Callable, Optional

try:
    from fastapi import Request, Response
    from starlette.middleware.base import BaseHTTPMiddleware
    FASTAPI_AVAILABLE = True
except ImportError:
    FASTAPI_AVAILABLE = False
    # Create dummy classes for type hints when FastAPI is not available
    class Request:
        pass
    class Response:
        pass
    class BaseHTTPMiddleware:
        def __init__(self, app):
            self.app = app

from .context import TraceContext, trace_context
from .http import HTTPPropagator, http_propagator
from .validation import sanitize_trace_id


logger = logging.getLogger(__name__)


class TracingMiddleware(BaseHTTPMiddleware):
    """
    FastAPI middleware for automatic trace_id management.
    
    Features:
    - Extracts trace_id from incoming requests
    - Generates new trace_id if none provided or invalid
    - Sets trace_id in context for request processing
    - Logs trace_id extraction and generation events
    - Graceful error handling and recovery
    
    Note: Requires FastAPI to be installed.
    """
    
    def __init__(
        self,
        app,
        trace_context_instance: Optional[TraceContext] = None,
        http_propagator_instance: Optional[HTTPPropagator] = None,
        generate_if_missing: bool = True,
        log_trace_events: bool = True
    ):
        """
        Initialize tracing middleware.
        
        Args:
            app: FastAPI application instance
            trace_context_instance: TraceContext instance (uses global if None)
            http_propagator_instance: HTTPPropagator instance (uses global if None)
            generate_if_missing: Whether to generate new trace_id if missing
            log_trace_events: Whether to log trace_id events
            
        Raises:
            ImportError: If FastAPI is not available
        """
        if not FASTAPI_AVAILABLE:
            raise ImportError("FastAPI is required for TracingMiddleware. Install with: pip install fastapi")
        
        super().__init__(app)
        self.trace_context = trace_context_instance or trace_context
        self.http_propagator = http_propagator_instance or http_propagator
        self.generate_if_missing = generate_if_missing
        self.log_trace_events = log_trace_events
    
    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        """
        Process request with trace_id management.
        
        Args:
            request: FastAPI request object
            call_next: Next middleware/handler in chain
            
        Returns:
            Response from downstream handlers
        """
        trace_id = None
        
        try:
            # Extract trace_id from request headers
            extracted_trace_id = self.http_propagator.extract_fastapi_request(request)
            
            if extracted_trace_id:
                # Sanitize and validate extracted trace_id
                trace_id = sanitize_trace_id(extracted_trace_id)
                
                if trace_id:
                    if self.log_trace_events:
                        logger.debug(f"Extracted valid trace_id from request: {trace_id}")
                else:
                    if self.log_trace_events:
                        logger.warning(f"Invalid trace_id format in request: {extracted_trace_id}")
            
            # Generate new trace_id if none found or invalid
            if not trace_id and self.generate_if_missing:
                trace_id = self.trace_context.generate_trace_id()
                if self.log_trace_events:
                    logger.debug(f"Generated new trace_id for request: {trace_id}")
            
            # Set trace_id in context if we have one
            if trace_id:
                self.trace_context.set_trace_id(trace_id)
            
            # Process request
            response = await call_next(request)
            
            return response
            
        except Exception as e:
            logger.error(f"Error in tracing middleware: {e}")
            
            # Try to generate fallback trace_id for error handling
            if not trace_id and self.generate_if_missing:
                try:
                    trace_id = self.trace_context.generate_trace_id()
                    self.trace_context.set_trace_id(trace_id)
                    logger.warning(f"Generated fallback trace_id after error: {trace_id}")
                except Exception as fallback_error:
                    logger.error(f"Failed to generate fallback trace_id: {fallback_error}")
            
            # Continue processing even if tracing fails
            response = await call_next(request)
            return response
        
        finally:
            # Clean up context (optional - contextvars handles this automatically)
            try:
                self.trace_context.clear_trace_id()
            except Exception as cleanup_error:
                logger.debug(f"Error cleaning up trace context: {cleanup_error}")


def create_tracing_middleware(
    generate_if_missing: bool = True,
    log_trace_events: bool = True,
    trace_context_instance: Optional[TraceContext] = None,
    http_propagator_instance: Optional[HTTPPropagator] = None
) -> type:
    """
    Factory function to create tracing middleware with custom configuration.
    
    Args:
        generate_if_missing: Whether to generate new trace_id if missing
        log_trace_events: Whether to log trace_id events
        trace_context_instance: Custom TraceContext instance
        http_propagator_instance: Custom HTTPPropagator instance
        
    Returns:
        Configured TracingMiddleware class
        
    Raises:
        ImportError: If FastAPI is not available
    """
    if not FASTAPI_AVAILABLE:
        raise ImportError("FastAPI is required for tracing middleware. Install with: pip install fastapi")
    
    class ConfiguredTracingMiddleware(TracingMiddleware):
        def __init__(self, app):
            super().__init__(
                app,
                trace_context_instance=trace_context_instance,
                http_propagator_instance=http_propagator_instance,
                generate_if_missing=generate_if_missing,
                log_trace_events=log_trace_events
            )
    
    return ConfiguredTracingMiddleware


# Convenience function for simple middleware setup
def setup_tracing_middleware(app, **kwargs):
    """
    Add tracing middleware to FastAPI app.
    
    Args:
        app: FastAPI application instance
        **kwargs: Arguments passed to create_tracing_middleware
        
    Raises:
        ImportError: If FastAPI is not available
    """
    if not FASTAPI_AVAILABLE:
        raise ImportError("FastAPI is required for tracing middleware. Install with: pip install fastapi")
    
    middleware_class = create_tracing_middleware(**kwargs)
    app.add_middleware(middleware_class)