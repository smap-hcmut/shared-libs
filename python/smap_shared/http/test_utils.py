"""
Test cases for HTTP utilities with trace integration.
"""

import json
import pytest
from unittest.mock import patch, MagicMock
from typing import Any, Dict

from .utils import HTTPUtils, ServiceClient, fetch_json_sync, post_json_data_sync


class TestHTTPUtils:
    """Test cases for HTTPUtils class."""
    
    def test_build_url(self):
        """Test URL building functionality."""
        # Test with base URL
        utils = HTTPUtils(base_url="https://api.example.com")
        assert utils.build_url("users") == "https://api.example.com/users"
        assert utils.build_url("/users") == "https://api.example.com/users"
        
        # Test with full URL (should return as-is)
        full_url = "https://other.example.com/data"
        assert utils.build_url(full_url) == full_url
        
        # Test without base URL
        utils_no_base = HTTPUtils()
        assert utils_no_base.build_url("users") == "users"
        assert utils_no_base.build_url("https://example.com/users") == "https://example.com/users"
    
    @patch('smap_shared.http.utils.create_sync_client')
    def test_get_json_sync(self, mock_create_client):
        """Test synchronous JSON GET request."""
        # Mock client and response
        mock_client = MagicMock()
        mock_client.get_sync.return_value = (b'{"result": "success"}', 200)
        mock_create_client.return_value = mock_client
        
        utils = HTTPUtils(base_url="https://api.example.com")
        result, status = utils.get_json_sync("/data")
        
        assert result == {"result": "success"}
        assert status == 200
        mock_client.get_sync.assert_called_once()
        mock_client.close.assert_called_once()
    
    @patch('smap_shared.http.utils.create_sync_client')
    def test_post_json_sync(self, mock_create_client):
        """Test synchronous JSON POST request."""
        # Mock client and response
        mock_client = MagicMock()
        mock_client.post_sync.return_value = (b'{"id": 123}', 201)
        mock_create_client.return_value = mock_client
        
        utils = HTTPUtils(base_url="https://api.example.com")
        test_data = {"name": "test"}
        result, status = utils.post_json_sync("/create", test_data)
        
        assert result == {"id": 123}
        assert status == 201
        mock_client.post_sync.assert_called_once()
        mock_client.close.assert_called_once()


class TestServiceClient:
    """Test cases for ServiceClient class."""
    
    def test_prepare_headers(self):
        """Test header preparation with authentication."""
        client = ServiceClient(
            service_name="test-service",
            base_url="https://api.example.com",
            auth_token="test-token"
        )
        
        headers = client._prepare_headers()
        assert headers['Authorization'] == 'Bearer test-token'
        assert headers['User-Agent'] == 'SMAP-Service-Client/test-service'
        
        # Test with additional headers
        custom_headers = {"Custom-Header": "value"}
        headers = client._prepare_headers(custom_headers)
        assert headers['Authorization'] == 'Bearer test-token'
        assert headers['Custom-Header'] == 'value'
    
    def test_prepare_headers_no_auth(self):
        """Test header preparation without authentication."""
        client = ServiceClient(
            service_name="test-service",
            base_url="https://api.example.com"
        )
        
        headers = client._prepare_headers()
        assert 'Authorization' not in headers
        assert headers['User-Agent'] == 'SMAP-Service-Client/test-service'


class TestConvenienceFunctions:
    """Test cases for convenience functions."""
    
    @patch('smap_shared.http.utils.HTTPUtils.get_json_sync')
    def test_fetch_json_sync(self, mock_get_json):
        """Test fetch_json_sync convenience function."""
        mock_get_json.return_value = ({"data": "test"}, 200)
        
        result, status = fetch_json_sync("https://api.example.com/data")
        
        assert result == {"data": "test"}
        assert status == 200
        mock_get_json.assert_called_once()
    
    @patch('smap_shared.http.utils.HTTPUtils.post_json_sync')
    def test_post_json_data_sync(self, mock_post_json):
        """Test post_json_data_sync convenience function."""
        mock_post_json.return_value = ({"id": 456}, 201)
        
        test_data = {"name": "test"}
        result, status = post_json_data_sync("https://api.example.com/create", test_data)
        
        assert result == {"id": 456}
        assert status == 201
        mock_post_json.assert_called_once()


def test_module_imports():
    """Test that all utilities can be imported correctly."""
    from .utils import (
        HTTPUtils,
        ServiceClient,
        fetch_json,
        fetch_json_sync,
        post_json_data,
        post_json_data_sync,
    )
    
    # Verify classes and functions exist
    assert HTTPUtils is not None
    assert ServiceClient is not None
    assert callable(fetch_json)
    assert callable(fetch_json_sync)
    assert callable(post_json_data)
    assert callable(post_json_data_sync)


if __name__ == "__main__":
    # Simple test runner for basic validation
    test_module_imports()
    
    # Test HTTPUtils
    utils_test = TestHTTPUtils()
    utils_test.test_build_url()
    
    # Test ServiceClient
    service_test = TestServiceClient()
    service_test.test_prepare_headers()
    service_test.test_prepare_headers_no_auth()
    
    print("✓ All HTTP utilities tests passed")