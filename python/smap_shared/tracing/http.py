"""
HTTP propagator implementation for Python services.

Handles trace_id injection and extraction for HTTP requests and responses.
"""

from typing import Optional, Dict, Union
import logging

from .interfaces import HTTPPropagatorInterface
from .context import TraceContext


# Standard header name for trace_id propagation (consistent with Go implementation)
TRACE_ID_HEADER = "X-Trace-Id"

logger = logging.getLogger(__name__)


class HTTPPropagator(HTTPPropagatorInterface):
    """
    HTTP propagator for trace_id management.
    
    Handles injection of trace_id into outbound HTTP requests and extraction
    from inbound HTTP requests. Compatible with FastAPI, httpx, requests, and
    other Python HTTP libraries.
    
    Features:
    - Automatic trace_id injection for outbound requests
    - Trace_id extraction from inbound requests
    - Graceful error handling and logging
    - Cross-language compatibility with Go services
    """
    
    def __init__(self, trace_context: Optional[TraceContext] = None):
        """
        Initialize HTTP propagator.
        
        Args:
            trace_context: TraceContext instance (uses global instance if None)
        """
        self.trace_context = trace_context or TraceContext()
    
    def inject_http(self, headers: Dict[str, str]) -> None:
        """
        Adds trace_id to outbound HTTP request headers.
        
        Args:
            headers: Dictionary of HTTP headers to modify
        """
        try:
            trace_id = self.trace_context.get_trace_id()
            if trace_id:
                headers[TRACE_ID_HEADER] = trace_id
                logger.debug(f"Injected trace_id into HTTP headers: {trace_id}")
            else:
                logger.debug("No trace_id in context, skipping HTTP injection")
        except Exception as e:
            logger.warning(f"Failed to inject trace_id into HTTP headers: {e}")
    
    def extract_http(self, headers: Dict[str, str]) -> Optional[str]:
        """
        Retrieves trace_id from inbound HTTP request headers.
        
        Args:
            headers: Dictionary of HTTP headers
            
        Returns:
            Extracted trace_id or None if not found/invalid
        """
        try:
            # Try different header name variations for robustness
            trace_id = (
                headers.get(TRACE_ID_HEADER) or
                headers.get(TRACE_ID_HEADER.lower()) or
                headers.get("x-trace-id") or
                headers.get("X-TRACE-ID")
            )
            
            if trace_id:
                # Validate extracted trace_id
                if self.trace_context.validate_trace_id(trace_id):
                    logger.debug(f"Extracted valid trace_id from HTTP headers: {trace_id}")
                    return trace_id
                else:
                    logger.warning(f"Invalid trace_id format in HTTP headers: {trace_id}")
                    return None
            else:
                logger.debug("No trace_id found in HTTP headers")
                return None
                
        except Exception as e:
            logger.warning(f"Failed to extract trace_id from HTTP headers: {e}")
            return None
    
    def inject_fastapi_request(self, request, headers: Optional[Dict[str, str]] = None) -> Dict[str, str]:
        """
        Convenience method for FastAPI request header injection.
        
        Args:
            request: FastAPI Request object
            headers: Additional headers to include
            
        Returns:
            Headers dictionary with trace_id injected
        """
        result_headers = headers.copy() if headers else {}
        self.inject_http(result_headers)
        return result_headers
    
    def extract_fastapi_request(self, request) -> Optional[str]:
        """
        Convenience method for FastAPI request header extraction.
        
        Args:
            request: FastAPI Request object
            
        Returns:
            Extracted trace_id or None
        """
        # Convert FastAPI headers to dict
        headers_dict = dict(request.headers)
        return self.extract_http(headers_dict)
    
    def inject_httpx_request(self, headers: Optional[Dict[str, str]] = None) -> Dict[str, str]:
        """
        Convenience method for httpx request header injection.
        
        Args:
            headers: Existing headers dictionary
            
        Returns:
            Headers dictionary with trace_id injected
        """
        result_headers = headers.copy() if headers else {}
        self.inject_http(result_headers)
        return result_headers
    
    def inject_requests_headers(self, headers: Optional[Dict[str, str]] = None) -> Dict[str, str]:
        """
        Convenience method for requests library header injection.
        
        Args:
            headers: Existing headers dictionary
            
        Returns:
            Headers dictionary with trace_id injected
        """
        result_headers = headers.copy() if headers else {}
        self.inject_http(result_headers)
        return result_headers


# Global instance for convenience
http_propagator = HTTPPropagator()


# Convenience functions for direct access
def inject_http_headers(headers: Dict[str, str]) -> None:
    """Inject trace_id into HTTP headers."""
    http_propagator.inject_http(headers)


def extract_http_headers(headers: Dict[str, str]) -> Optional[str]:
    """Extract trace_id from HTTP headers."""
    return http_propagator.extract_http(headers)


def get_traced_headers(base_headers: Optional[Dict[str, str]] = None) -> Dict[str, str]:
    """Get headers with trace_id injected."""
    return http_propagator.inject_httpx_request(base_headers)