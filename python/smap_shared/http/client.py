"""
HTTP client implementation with automatic trace_id injection.

Provides TracedHTTPClient wrapper for httpx and requests libraries with
automatic X-Trace-Id header injection for distributed tracing.
"""

import asyncio
import logging
from typing import Any, Dict, Optional, Union, Tuple
from contextlib import asynccontextmanager

try:
    import httpx
    HTTPX_AVAILABLE = True
except ImportError:
    HTTPX_AVAILABLE = False

try:
    import requests
    REQUESTS_AVAILABLE = True
except ImportError:
    REQUESTS_AVAILABLE = False

from ..tracing import (
    get_trace_id,
    http_propagator as global_http_propagator,
    HTTPPropagator,
    TraceContext,
    trace_context as global_trace_context,
)

logger = logging.getLogger(__name__)


class TracedHTTPClient:
    """
    HTTP client with automatic trace_id injection.
    
    Wraps httpx.AsyncClient or requests.Session with automatic X-Trace-Id
    header injection for distributed tracing. Supports both async and sync
    operations.
    
    Features:
    - Automatic trace_id injection into all outbound requests
    - Support for both httpx (async) and requests (sync) backends
    - Configurable timeout, retries, and base headers
    - Graceful error handling and logging
    - Compatible with existing HTTP client patterns
    
    Usage:
        # Async usage with httpx
        async with TracedHTTPClient() as client:
            response = await client.get("https://api.example.com/data")
            
        # Sync usage with requests
        client = TracedHTTPClient(use_async=False)
        response = client.get("https://api.example.com/data")
    """
    
    def __init__(
        self,
        use_async: bool = True,
        timeout: float = 30.0,
        retries: int = 3,
        base_headers: Optional[Dict[str, str]] = None,
        trace_context: Optional[TraceContext] = None,
        http_propagator: Optional[HTTPPropagator] = None,
        **client_kwargs
    ):
        """
        Initialize TracedHTTPClient.
        
        Args:
            use_async: Whether to use async httpx client (True) or sync requests (False)
            timeout: Request timeout in seconds
            retries: Number of retry attempts for failed requests
            base_headers: Default headers to include in all requests
            trace_context: TraceContext instance (uses global if None)
            http_propagator: HTTPPropagator instance (uses global if None)
            **client_kwargs: Additional arguments passed to underlying client
        """
        self.use_async = use_async
        self.timeout = timeout
        self.retries = retries
        self.base_headers = base_headers or {}
        self.trace_context = trace_context or global_trace_context
        self.http_propagator = http_propagator or global_http_propagator
        self.client_kwargs = client_kwargs
        
        # Initialize client based on backend preference
        if self.use_async:
            if not HTTPX_AVAILABLE:
                raise ImportError("httpx is required for async HTTP client. Install with: pip install httpx")
            self._client = None  # Will be initialized in async context
        else:
            if not REQUESTS_AVAILABLE:
                raise ImportError("requests is required for sync HTTP client. Install with: pip install requests")
            self._client = self._create_sync_client()
    
    def _create_sync_client(self) -> "requests.Session":
        """Create synchronous requests session."""
        import requests
        from requests.adapters import HTTPAdapter
        from urllib3.util.retry import Retry
        
        session = requests.Session()
        
        # Configure retry strategy (use allowed_methods instead of method_whitelist for newer versions)
        try:
            retry_strategy = Retry(
                total=self.retries,
                status_forcelist=[429, 500, 502, 503, 504],
                allowed_methods=["HEAD", "GET", "OPTIONS", "POST", "PUT", "DELETE"],
                backoff_factor=1
            )
        except TypeError:
            # Fallback for older urllib3 versions
            retry_strategy = Retry(
                total=self.retries,
                status_forcelist=[429, 500, 502, 503, 504],
                method_whitelist=["HEAD", "GET", "OPTIONS", "POST", "PUT", "DELETE"],
                backoff_factor=1
            )
        
        adapter = HTTPAdapter(max_retries=retry_strategy)
        session.mount("http://", adapter)
        session.mount("https://", adapter)
        
        # Set default timeout
        session.timeout = self.timeout
        
        return session
    
    async def _create_async_client(self) -> "httpx.AsyncClient":
        """Create asynchronous httpx client."""
        import httpx
        
        # Configure retry transport
        transport = httpx.AsyncHTTPTransport(retries=self.retries)
        
        client = httpx.AsyncClient(
            timeout=self.timeout,
            transport=transport,
            **self.client_kwargs
        )
        
        return client
    
    def _prepare_headers(self, headers: Optional[Dict[str, str]] = None) -> Dict[str, str]:
        """
        Prepare headers with trace_id injection.
        
        Args:
            headers: Additional headers to include
            
        Returns:
            Headers dictionary with trace_id and base headers
        """
        # Start with base headers
        result_headers = self.base_headers.copy()
        
        # Add custom headers
        if headers:
            result_headers.update(headers)
        
        # Inject trace_id
        self.http_propagator.inject_http(result_headers)
        
        return result_headers
    
    # Async methods (httpx backend)
    async def __aenter__(self):
        """Async context manager entry."""
        if self.use_async:
            self._client = await self._create_async_client()
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        if self.use_async and self._client:
            await self._client.aclose()
    
    async def get(
        self,
        url: str,
        headers: Optional[Dict[str, str]] = None,
        params: Optional[Dict[str, Any]] = None,
        **kwargs
    ) -> Tuple[bytes, int]:
        """
        Perform async GET request with trace_id injection.
        
        Args:
            url: Request URL
            headers: Additional headers
            params: Query parameters
            **kwargs: Additional arguments passed to httpx
            
        Returns:
            Tuple of (response_body, status_code)
        """
        if not self.use_async:
            raise RuntimeError("Use sync methods for sync client")
        
        prepared_headers = self._prepare_headers(headers)
        
        try:
            response = await self._client.get(
                url,
                headers=prepared_headers,
                params=params,
                **kwargs
            )
            return response.content, response.status_code
        except Exception as e:
            logger.error(f"HTTP GET request failed: {url} - {e}")
            raise
    
    async def post(
        self,
        url: str,
        data: Optional[Any] = None,
        json: Optional[Any] = None,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[bytes, int]:
        """
        Perform async POST request with trace_id injection.
        
        Args:
            url: Request URL
            data: Request body data
            json: JSON data to send
            headers: Additional headers
            **kwargs: Additional arguments passed to httpx
            
        Returns:
            Tuple of (response_body, status_code)
        """
        if not self.use_async:
            raise RuntimeError("Use sync methods for sync client")
        
        prepared_headers = self._prepare_headers(headers)
        
        try:
            response = await self._client.post(
                url,
                data=data,
                json=json,
                headers=prepared_headers,
                **kwargs
            )
            return response.content, response.status_code
        except Exception as e:
            logger.error(f"HTTP POST request failed: {url} - {e}")
            raise
    
    async def put(
        self,
        url: str,
        data: Optional[Any] = None,
        json: Optional[Any] = None,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[bytes, int]:
        """
        Perform async PUT request with trace_id injection.
        
        Args:
            url: Request URL
            data: Request body data
            json: JSON data to send
            headers: Additional headers
            **kwargs: Additional arguments passed to httpx
            
        Returns:
            Tuple of (response_body, status_code)
        """
        if not self.use_async:
            raise RuntimeError("Use sync methods for sync client")
        
        prepared_headers = self._prepare_headers(headers)
        
        try:
            response = await self._client.put(
                url,
                data=data,
                json=json,
                headers=prepared_headers,
                **kwargs
            )
            return response.content, response.status_code
        except Exception as e:
            logger.error(f"HTTP PUT request failed: {url} - {e}")
            raise
    
    async def delete(
        self,
        url: str,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[bytes, int]:
        """
        Perform async DELETE request with trace_id injection.
        
        Args:
            url: Request URL
            headers: Additional headers
            **kwargs: Additional arguments passed to httpx
            
        Returns:
            Tuple of (response_body, status_code)
        """
        if not self.use_async:
            raise RuntimeError("Use sync methods for sync client")
        
        prepared_headers = self._prepare_headers(headers)
        
        try:
            response = await self._client.delete(
                url,
                headers=prepared_headers,
                **kwargs
            )
            return response.content, response.status_code
        except Exception as e:
            logger.error(f"HTTP DELETE request failed: {url} - {e}")
            raise
    
    # Sync methods (requests backend)
    def get_sync(
        self,
        url: str,
        headers: Optional[Dict[str, str]] = None,
        params: Optional[Dict[str, Any]] = None,
        **kwargs
    ) -> Tuple[bytes, int]:
        """
        Perform sync GET request with trace_id injection.
        
        Args:
            url: Request URL
            headers: Additional headers
            params: Query parameters
            **kwargs: Additional arguments passed to requests
            
        Returns:
            Tuple of (response_body, status_code)
        """
        if self.use_async:
            raise RuntimeError("Use async methods for async client")
        
        prepared_headers = self._prepare_headers(headers)
        
        try:
            response = self._client.get(
                url,
                headers=prepared_headers,
                params=params,
                **kwargs
            )
            return response.content, response.status_code
        except Exception as e:
            logger.error(f"HTTP GET request failed: {url} - {e}")
            raise
    
    def post_sync(
        self,
        url: str,
        data: Optional[Any] = None,
        json: Optional[Any] = None,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[bytes, int]:
        """
        Perform sync POST request with trace_id injection.
        
        Args:
            url: Request URL
            data: Request body data
            json: JSON data to send
            headers: Additional headers
            **kwargs: Additional arguments passed to requests
            
        Returns:
            Tuple of (response_body, status_code)
        """
        if self.use_async:
            raise RuntimeError("Use async methods for async client")
        
        prepared_headers = self._prepare_headers(headers)
        
        try:
            response = self._client.post(
                url,
                data=data,
                json=json,
                headers=prepared_headers,
                **kwargs
            )
            return response.content, response.status_code
        except Exception as e:
            logger.error(f"HTTP POST request failed: {url} - {e}")
            raise
    
    def put_sync(
        self,
        url: str,
        data: Optional[Any] = None,
        json: Optional[Any] = None,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[bytes, int]:
        """
        Perform sync PUT request with trace_id injection.
        
        Args:
            url: Request URL
            data: Request body data
            json: JSON data to send
            headers: Additional headers
            **kwargs: Additional arguments passed to requests
            
        Returns:
            Tuple of (response_body, status_code)
        """
        if self.use_async:
            raise RuntimeError("Use async methods for async client")
        
        prepared_headers = self._prepare_headers(headers)
        
        try:
            response = self._client.put(
                url,
                data=data,
                json=json,
                headers=prepared_headers,
                **kwargs
            )
            return response.content, response.status_code
        except Exception as e:
            logger.error(f"HTTP PUT request failed: {url} - {e}")
            raise
    
    def delete_sync(
        self,
        url: str,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[bytes, int]:
        """
        Perform sync DELETE request with trace_id injection.
        
        Args:
            url: Request URL
            headers: Additional headers
            **kwargs: Additional arguments passed to requests
            
        Returns:
            Tuple of (response_body, status_code)
        """
        if self.use_async:
            raise RuntimeError("Use async methods for async client")
        
        prepared_headers = self._prepare_headers(headers)
        
        try:
            response = self._client.delete(
                url,
                headers=prepared_headers,
                **kwargs
            )
            return response.content, response.status_code
        except Exception as e:
            logger.error(f"HTTP DELETE request failed: {url} - {e}")
            raise
    
    def close(self):
        """Close the sync client session."""
        if not self.use_async and self._client:
            self._client.close()


# Convenience functions for quick usage
def create_async_client(**kwargs) -> TracedHTTPClient:
    """Create async HTTP client with trace injection."""
    return TracedHTTPClient(use_async=True, **kwargs)


def create_sync_client(**kwargs) -> TracedHTTPClient:
    """Create sync HTTP client with trace injection."""
    return TracedHTTPClient(use_async=False, **kwargs)


@asynccontextmanager
async def traced_httpx_client(**kwargs):
    """Async context manager for httpx client with trace injection."""
    async with TracedHTTPClient(use_async=True, **kwargs) as client:
        yield client


def traced_requests_session(**kwargs) -> TracedHTTPClient:
    """Create requests session with trace injection."""
    return TracedHTTPClient(use_async=False, **kwargs)