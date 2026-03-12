"""
Factory functions for creating tracing components.

Provides convenient factory functions for creating and configuring tracing
components with sensible defaults.
"""

from typing import Optional

from .context import TraceContext
from .http import HTTPPropagator
from .kafka import KafkaPropagator
from .middleware import TracingMiddleware


class TracingFactory:
    """
    Factory class for creating tracing components.
    
    Provides a centralized way to create and configure tracing components
    with consistent settings and dependencies.
    """
    
    @staticmethod
    def create_trace_context() -> TraceContext:
        """
        Create a new TraceContext instance.
        
        Returns:
            Configured TraceContext instance
        """
        return TraceContext()
    
    @staticmethod
    def create_http_propagator(trace_context: Optional[TraceContext] = None) -> HTTPPropagator:
        """
        Create a new HTTPPropagator instance.
        
        Args:
            trace_context: TraceContext instance (creates new if None)
            
        Returns:
            Configured HTTPPropagator instance
        """
        if trace_context is None:
            trace_context = TracingFactory.create_trace_context()
        
        return HTTPPropagator(trace_context)
    
    @staticmethod
    def create_kafka_propagator(trace_context: Optional[TraceContext] = None) -> KafkaPropagator:
        """
        Create a new KafkaPropagator instance.
        
        Args:
            trace_context: TraceContext instance (creates new if None)
            
        Returns:
            Configured KafkaPropagator instance
        """
        if trace_context is None:
            trace_context = TracingFactory.create_trace_context()
        
        return KafkaPropagator(trace_context)
    
    @staticmethod
    def create_complete_tracing_suite() -> tuple[TraceContext, HTTPPropagator, KafkaPropagator]:
        """
        Create a complete tracing suite with all components.
        
        Returns:
            Tuple of (TraceContext, HTTPPropagator, KafkaPropagator)
        """
        trace_context = TracingFactory.create_trace_context()
        http_propagator = TracingFactory.create_http_propagator(trace_context)
        kafka_propagator = TracingFactory.create_kafka_propagator(trace_context)
        
        return trace_context, http_propagator, kafka_propagator
    
    @staticmethod
    def create_middleware(
        trace_context: Optional[TraceContext] = None,
        http_propagator: Optional[HTTPPropagator] = None,
        generate_if_missing: bool = True,
        log_trace_events: bool = True
    ) -> type:
        """
        Create a configured TracingMiddleware class.
        
        Args:
            trace_context: TraceContext instance (creates new if None)
            http_propagator: HTTPPropagator instance (creates new if None)
            generate_if_missing: Whether to generate new trace_id if missing
            log_trace_events: Whether to log trace_id events
            
        Returns:
            Configured TracingMiddleware class
        """
        if trace_context is None:
            trace_context = TracingFactory.create_trace_context()
        
        if http_propagator is None:
            http_propagator = TracingFactory.create_http_propagator(trace_context)
        
        class ConfiguredTracingMiddleware(TracingMiddleware):
            def __init__(self, app):
                super().__init__(
                    app,
                    trace_context_instance=trace_context,
                    http_propagator_instance=http_propagator,
                    generate_if_missing=generate_if_missing,
                    log_trace_events=log_trace_events
                )
        
        return ConfiguredTracingMiddleware


# Convenience functions
def create_tracing_suite() -> tuple[TraceContext, HTTPPropagator, KafkaPropagator]:
    """Create a complete tracing suite with all components."""
    return TracingFactory.create_complete_tracing_suite()


def create_fastapi_middleware(**kwargs) -> type:
    """Create a configured FastAPI tracing middleware."""
    return TracingFactory.create_middleware(**kwargs)