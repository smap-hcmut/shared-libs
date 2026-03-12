"""
HTTP utility functions with trace integration.

Provides common HTTP utility functions that are frequently used across services
with automatic trace_id propagation and error handling.
"""

import json
import logging
from typing import Any, Dict, Optional, Union, Tuple
from urllib.parse import urljoin, urlparse

from .client import TracedHTTPClient, create_async_client, create_sync_client
from ..tracing import get_trace_id, set_trace_id

logger = logging.getLogger(__name__)


class HTTPUtils:
    """
    Collection of HTTP utility functions with trace integration.
    
    Provides commonly used HTTP operations with automatic trace_id propagation,
    JSON handling, and error management. Compatible with existing service patterns.
    """
    
    def __init__(self, base_url: Optional[str] = None, timeout: float = 30.0):
        """
        Initialize HTTP utilities.
        
        Args:
            base_url: Base URL for relative requests
            timeout: Default timeout for requests
        """
        self.base_url = base_url
        self.timeout = timeout
    
    def build_url(self, path: str) -> str:
        """
        Build full URL from base URL and path.
        
        Args:
            path: URL path or full URL
            
        Returns:
            Complete URL
        """
        if not path:
            return self.base_url or ""
        
        # If path is already a full URL, return as-is
        if urlparse(path).scheme:
            return path
        
        # Join with base URL if available
        if self.base_url:
            return urljoin(self.base_url.rstrip('/') + '/', path.lstrip('/'))
        
        return path
    
    async def get_json(
        self,
        url: str,
        headers: Optional[Dict[str, str]] = None,
        params: Optional[Dict[str, Any]] = None,
        **kwargs
    ) -> Tuple[Any, int]:
        """
        Perform GET request and parse JSON response.
        
        Args:
            url: Request URL (can be relative if base_url is set)
            headers: Additional headers
            params: Query parameters
            **kwargs: Additional arguments passed to client
            
        Returns:
            Tuple of (parsed_json, status_code)
            
        Raises:
            json.JSONDecodeError: If response is not valid JSON
            Exception: For HTTP errors
        """
        full_url = self.build_url(url)
        
        async with create_async_client(timeout=self.timeout) as client:
            response_body, status_code = await client.get(
                full_url, headers=headers, params=params, **kwargs
            )
            
            try:
                parsed_json = json.loads(response_body.decode('utf-8'))
                return parsed_json, status_code
            except json.JSONDecodeError as e:
                logger.error(f"Failed to parse JSON response from {full_url}: {e}")
                raise
    
    async def post_json(
        self,
        url: str,
        data: Any,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[Any, int]:
        """
        Perform POST request with JSON data and parse JSON response.
        
        Args:
            url: Request URL (can be relative if base_url is set)
            data: Data to send as JSON
            headers: Additional headers
            **kwargs: Additional arguments passed to client
            
        Returns:
            Tuple of (parsed_json, status_code)
            
        Raises:
            json.JSONDecodeError: If response is not valid JSON
            Exception: For HTTP errors
        """
        full_url = self.build_url(url)
        
        # Ensure Content-Type is set for JSON
        request_headers = headers or {}
        if 'Content-Type' not in request_headers:
            request_headers['Content-Type'] = 'application/json'
        
        async with create_async_client(timeout=self.timeout) as client:
            response_body, status_code = await client.post(
                full_url, json=data, headers=request_headers, **kwargs
            )
            
            try:
                parsed_json = json.loads(response_body.decode('utf-8'))
                return parsed_json, status_code
            except json.JSONDecodeError as e:
                logger.error(f"Failed to parse JSON response from {full_url}: {e}")
                raise
    
    def get_json_sync(
        self,
        url: str,
        headers: Optional[Dict[str, str]] = None,
        params: Optional[Dict[str, Any]] = None,
        **kwargs
    ) -> Tuple[Any, int]:
        """
        Perform sync GET request and parse JSON response.
        
        Args:
            url: Request URL (can be relative if base_url is set)
            headers: Additional headers
            params: Query parameters
            **kwargs: Additional arguments passed to client
            
        Returns:
            Tuple of (parsed_json, status_code)
            
        Raises:
            json.JSONDecodeError: If response is not valid JSON
            Exception: For HTTP errors
        """
        full_url = self.build_url(url)
        
        client = create_sync_client(timeout=self.timeout)
        try:
            response_body, status_code = client.get_sync(
                full_url, headers=headers, params=params, **kwargs
            )
            
            try:
                parsed_json = json.loads(response_body.decode('utf-8'))
                return parsed_json, status_code
            except json.JSONDecodeError as e:
                logger.error(f"Failed to parse JSON response from {full_url}: {e}")
                raise
        finally:
            client.close()
    
    def post_json_sync(
        self,
        url: str,
        data: Any,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[Any, int]:
        """
        Perform sync POST request with JSON data and parse JSON response.
        
        Args:
            url: Request URL (can be relative if base_url is set)
            data: Data to send as JSON
            headers: Additional headers
            **kwargs: Additional arguments passed to client
            
        Returns:
            Tuple of (parsed_json, status_code)
            
        Raises:
            json.JSONDecodeError: If response is not valid JSON
            Exception: For HTTP errors
        """
        full_url = self.build_url(url)
        
        # Ensure Content-Type is set for JSON
        request_headers = headers or {}
        if 'Content-Type' not in request_headers:
            request_headers['Content-Type'] = 'application/json'
        
        client = create_sync_client(timeout=self.timeout)
        try:
            response_body, status_code = client.post_sync(
                full_url, json=data, headers=request_headers, **kwargs
            )
            
            try:
                parsed_json = json.loads(response_body.decode('utf-8'))
                return parsed_json, status_code
            except json.JSONDecodeError as e:
                logger.error(f"Failed to parse JSON response from {full_url}: {e}")
                raise
        finally:
            client.close()


# Convenience functions for common patterns
async def fetch_json(
    url: str,
    headers: Optional[Dict[str, str]] = None,
    params: Optional[Dict[str, Any]] = None,
    timeout: float = 30.0,
    **kwargs
) -> Tuple[Any, int]:
    """
    Convenience function to fetch JSON data with trace integration.
    
    Args:
        url: Request URL
        headers: Additional headers
        params: Query parameters
        timeout: Request timeout
        **kwargs: Additional arguments
        
    Returns:
        Tuple of (parsed_json, status_code)
    """
    utils = HTTPUtils(timeout=timeout)
    return await utils.get_json(url, headers=headers, params=params, **kwargs)


def fetch_json_sync(
    url: str,
    headers: Optional[Dict[str, str]] = None,
    params: Optional[Dict[str, Any]] = None,
    timeout: float = 30.0,
    **kwargs
) -> Tuple[Any, int]:
    """
    Convenience function to fetch JSON data synchronously with trace integration.
    
    Args:
        url: Request URL
        headers: Additional headers
        params: Query parameters
        timeout: Request timeout
        **kwargs: Additional arguments
        
    Returns:
        Tuple of (parsed_json, status_code)
    """
    utils = HTTPUtils(timeout=timeout)
    return utils.get_json_sync(url, headers=headers, params=params, **kwargs)


async def post_json_data(
    url: str,
    data: Any,
    headers: Optional[Dict[str, str]] = None,
    timeout: float = 30.0,
    **kwargs
) -> Tuple[Any, int]:
    """
    Convenience function to post JSON data with trace integration.
    
    Args:
        url: Request URL
        data: Data to send as JSON
        headers: Additional headers
        timeout: Request timeout
        **kwargs: Additional arguments
        
    Returns:
        Tuple of (parsed_json, status_code)
    """
    utils = HTTPUtils(timeout=timeout)
    return await utils.post_json(url, data, headers=headers, **kwargs)


def post_json_data_sync(
    url: str,
    data: Any,
    headers: Optional[Dict[str, str]] = None,
    timeout: float = 30.0,
    **kwargs
) -> Tuple[Any, int]:
    """
    Convenience function to post JSON data synchronously with trace integration.
    
    Args:
        url: Request URL
        data: Data to send as JSON
        headers: Additional headers
        timeout: Request timeout
        **kwargs: Additional arguments
        
    Returns:
        Tuple of (parsed_json, status_code)
    """
    utils = HTTPUtils(timeout=timeout)
    return utils.post_json_sync(url, data, headers=headers, **kwargs)


# Service-to-service communication helpers
class ServiceClient:
    """
    HTTP client for service-to-service communication with trace propagation.
    
    Provides a higher-level interface for calling other SMAP services with
    automatic trace_id propagation, authentication, and error handling.
    """
    
    def __init__(
        self,
        service_name: str,
        base_url: str,
        auth_token: Optional[str] = None,
        timeout: float = 30.0
    ):
        """
        Initialize service client.
        
        Args:
            service_name: Name of the target service (for logging)
            base_url: Base URL of the target service
            auth_token: Authentication token (JWT)
            timeout: Request timeout
        """
        self.service_name = service_name
        self.base_url = base_url.rstrip('/')
        self.auth_token = auth_token
        self.timeout = timeout
        self.utils = HTTPUtils(base_url=base_url, timeout=timeout)
    
    def _prepare_headers(self, headers: Optional[Dict[str, str]] = None) -> Dict[str, str]:
        """Prepare headers with authentication and service identification."""
        result_headers = headers or {}
        
        # Add authentication if available
        if self.auth_token:
            result_headers['Authorization'] = f'Bearer {self.auth_token}'
        
        # Add service identification
        result_headers['User-Agent'] = f'SMAP-Service-Client/{self.service_name}'
        
        return result_headers
    
    async def call_api(
        self,
        endpoint: str,
        method: str = 'GET',
        data: Optional[Any] = None,
        params: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
        **kwargs
    ) -> Tuple[Any, int]:
        """
        Call service API endpoint with trace propagation.
        
        Args:
            endpoint: API endpoint path
            method: HTTP method (GET, POST, PUT, DELETE)
            data: Request data (will be JSON-encoded for POST/PUT)
            params: Query parameters
            headers: Additional headers
            **kwargs: Additional arguments
            
        Returns:
            Tuple of (response_data, status_code)
        """
        prepared_headers = self._prepare_headers(headers)
        
        try:
            if method.upper() == 'GET':
                return await self.utils.get_json(
                    endpoint, headers=prepared_headers, params=params, **kwargs
                )
            elif method.upper() in ['POST', 'PUT']:
                if method.upper() == 'POST':
                    return await self.utils.post_json(
                        endpoint, data, headers=prepared_headers, **kwargs
                    )
                else:  # PUT
                    async with create_async_client(timeout=self.timeout) as client:
                        response_body, status_code = await client.put(
                            self.utils.build_url(endpoint),
                            json=data,
                            headers=prepared_headers,
                            **kwargs
                        )
                        return json.loads(response_body.decode('utf-8')), status_code
            else:
                raise ValueError(f"Unsupported HTTP method: {method}")
                
        except Exception as e:
            logger.error(f"Service call failed: {self.service_name} {method} {endpoint} - {e}")
            raise


# Export commonly used functions
__all__ = [
    'HTTPUtils',
    'ServiceClient',
    'fetch_json',
    'fetch_json_sync',
    'post_json_data',
    'post_json_data_sync',
]