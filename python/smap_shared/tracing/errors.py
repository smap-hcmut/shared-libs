"""
Error classes for tracing operations.

Defines custom exceptions for tracing-related errors with consistent error handling.
"""


class TracingError(Exception):
    """Base exception for all tracing-related errors."""
    
    def __init__(self, message: str, trace_id: str = None):
        """
        Initialize tracing error.
        
        Args:
            message: Error message
            trace_id: Associated trace_id (if any)
        """
        super().__init__(message)
        self.trace_id = trace_id


class InvalidTraceIDError(TracingError):
    """Raised when an invalid trace_id format is encountered."""
    
    def __init__(self, trace_id: str, message: str = None):
        """
        Initialize invalid trace_id error.
        
        Args:
            trace_id: The invalid trace_id
            message: Custom error message
        """
        if message is None:
            message = f"Invalid trace_id format: {trace_id}"
        
        super().__init__(message, trace_id)
        self.invalid_trace_id = trace_id


class TraceContextError(TracingError):
    """Raised when trace context operations fail."""
    
    def __init__(self, message: str, operation: str = None, trace_id: str = None):
        """
        Initialize trace context error.
        
        Args:
            message: Error message
            operation: The operation that failed
            trace_id: Associated trace_id (if any)
        """
        super().__init__(message, trace_id)
        self.operation = operation


class PropagationError(TracingError):
    """Raised when trace_id propagation fails."""
    
    def __init__(self, message: str, propagation_type: str = None, trace_id: str = None):
        """
        Initialize propagation error.
        
        Args:
            message: Error message
            propagation_type: Type of propagation (HTTP, Kafka, etc.)
            trace_id: Associated trace_id (if any)
        """
        super().__init__(message, trace_id)
        self.propagation_type = propagation_type


class HTTPPropagationError(PropagationError):
    """Raised when HTTP trace_id propagation fails."""
    
    def __init__(self, message: str, trace_id: str = None):
        """
        Initialize HTTP propagation error.
        
        Args:
            message: Error message
            trace_id: Associated trace_id (if any)
        """
        super().__init__(message, "HTTP", trace_id)


class KafkaPropagationError(PropagationError):
    """Raised when Kafka trace_id propagation fails."""
    
    def __init__(self, message: str, trace_id: str = None):
        """
        Initialize Kafka propagation error.
        
        Args:
            message: Error message
            trace_id: Associated trace_id (if any)
        """
        super().__init__(message, "Kafka", trace_id)


class MiddlewareError(TracingError):
    """Raised when tracing middleware encounters errors."""
    
    def __init__(self, message: str, middleware_type: str = None, trace_id: str = None):
        """
        Initialize middleware error.
        
        Args:
            message: Error message
            middleware_type: Type of middleware (FastAPI, etc.)
            trace_id: Associated trace_id (if any)
        """
        super().__init__(message, trace_id)
        self.middleware_type = middleware_type