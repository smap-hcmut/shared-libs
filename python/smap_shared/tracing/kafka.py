"""
Kafka propagator implementation for Python services.

Handles trace_id injection and extraction for Kafka message headers.
"""

from typing import Optional, Dict, Any, List, Tuple
import logging

from .interfaces import KafkaPropagatorInterface
from .context import TraceContext
from .http import TRACE_ID_HEADER  # Use same header name as HTTP


logger = logging.getLogger(__name__)


class KafkaPropagator(KafkaPropagatorInterface):
    """
    Kafka propagator for trace_id management.
    
    Handles injection of trace_id into Kafka message headers and extraction
    from consumed messages. Compatible with aiokafka, kafka-python, and other
    Python Kafka libraries.
    
    Features:
    - Automatic trace_id injection for produced messages
    - Trace_id extraction from consumed messages
    - Support for different Kafka header formats
    - Graceful error handling and logging
    - Cross-language compatibility with Go services
    """
    
    def __init__(self, trace_context: Optional[TraceContext] = None):
        """
        Initialize Kafka propagator.
        
        Args:
            trace_context: TraceContext instance (uses global instance if None)
        """
        self.trace_context = trace_context or TraceContext()
    
    def inject_kafka(self, headers: Dict[str, str]) -> None:
        """
        Adds trace_id to Kafka message headers.
        
        Args:
            headers: Dictionary of Kafka message headers to modify
        """
        try:
            trace_id = self.trace_context.get_trace_id()
            if trace_id:
                headers[TRACE_ID_HEADER] = trace_id
                logger.debug(f"Injected trace_id into Kafka headers: {trace_id}")
            else:
                logger.debug("No trace_id in context, skipping Kafka injection")
        except Exception as e:
            logger.warning(f"Failed to inject trace_id into Kafka headers: {e}")
    
    def extract_kafka(self, headers: Dict[str, str]) -> Optional[str]:
        """
        Retrieves trace_id from Kafka message headers.
        
        Args:
            headers: Dictionary of Kafka message headers
            
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
                    logger.debug(f"Extracted valid trace_id from Kafka headers: {trace_id}")
                    return trace_id
                else:
                    logger.warning(f"Invalid trace_id format in Kafka headers: {trace_id}")
                    return None
            else:
                logger.debug("No trace_id found in Kafka headers")
                return None
                
        except Exception as e:
            logger.warning(f"Failed to extract trace_id from Kafka headers: {e}")
            return None
    
    def inject_aiokafka_headers(self, headers: Optional[List[Tuple[str, bytes]]] = None) -> List[Tuple[str, bytes]]:
        """
        Convenience method for aiokafka header injection.
        
        Args:
            headers: Existing aiokafka headers list
            
        Returns:
            Headers list with trace_id injected
        """
        result_headers = list(headers) if headers else []
        
        try:
            trace_id = self.trace_context.get_trace_id()
            if trace_id:
                # Add trace_id header as (key, value) tuple with bytes value
                result_headers.append((TRACE_ID_HEADER, trace_id.encode('utf-8')))
                logger.debug(f"Injected trace_id into aiokafka headers: {trace_id}")
        except Exception as e:
            logger.warning(f"Failed to inject trace_id into aiokafka headers: {e}")
        
        return result_headers
    
    def extract_aiokafka_headers(self, headers: Optional[List[Tuple[str, bytes]]]) -> Optional[str]:
        """
        Convenience method for aiokafka header extraction.
        
        Args:
            headers: aiokafka headers list
            
        Returns:
            Extracted trace_id or None
        """
        if not headers:
            return None
        
        try:
            # Convert aiokafka headers to dict
            headers_dict = {}
            for key, value in headers:
                if isinstance(value, bytes):
                    headers_dict[key] = value.decode('utf-8')
                else:
                    headers_dict[key] = str(value)
            
            return self.extract_kafka(headers_dict)
        except Exception as e:
            logger.warning(f"Failed to extract trace_id from aiokafka headers: {e}")
            return None
    
    def inject_kafka_python_headers(self, headers: Optional[Dict[str, bytes]] = None) -> Dict[str, bytes]:
        """
        Convenience method for kafka-python header injection.
        
        Args:
            headers: Existing kafka-python headers dict
            
        Returns:
            Headers dict with trace_id injected
        """
        result_headers = dict(headers) if headers else {}
        
        try:
            trace_id = self.trace_context.get_trace_id()
            if trace_id:
                # Add trace_id header as bytes value
                result_headers[TRACE_ID_HEADER] = trace_id.encode('utf-8')
                logger.debug(f"Injected trace_id into kafka-python headers: {trace_id}")
        except Exception as e:
            logger.warning(f"Failed to inject trace_id into kafka-python headers: {e}")
        
        return result_headers
    
    def extract_kafka_python_headers(self, headers: Optional[Dict[str, bytes]]) -> Optional[str]:
        """
        Convenience method for kafka-python header extraction.
        
        Args:
            headers: kafka-python headers dict
            
        Returns:
            Extracted trace_id or None
        """
        if not headers:
            return None
        
        try:
            # Convert kafka-python headers to string dict
            headers_dict = {}
            for key, value in headers.items():
                if isinstance(value, bytes):
                    headers_dict[key] = value.decode('utf-8')
                else:
                    headers_dict[key] = str(value)
            
            return self.extract_kafka(headers_dict)
        except Exception as e:
            logger.warning(f"Failed to extract trace_id from kafka-python headers: {e}")
            return None
    
    def inject_message_headers(self, message_dict: Dict[str, Any]) -> Dict[str, Any]:
        """
        Inject trace_id into a message dictionary with headers.
        
        Args:
            message_dict: Message dictionary that may contain 'headers' key
            
        Returns:
            Message dictionary with trace_id injected into headers
        """
        result_message = message_dict.copy()
        
        # Ensure headers exist
        if 'headers' not in result_message:
            result_message['headers'] = {}
        
        # Inject trace_id into headers
        self.inject_kafka(result_message['headers'])
        
        return result_message
    
    def extract_message_headers(self, message_dict: Dict[str, Any]) -> Optional[str]:
        """
        Extract trace_id from a message dictionary with headers.
        
        Args:
            message_dict: Message dictionary that may contain 'headers' key
            
        Returns:
            Extracted trace_id or None
        """
        headers = message_dict.get('headers', {})
        if not headers:
            return None
        
        return self.extract_kafka(headers)


# Global instance for convenience
kafka_propagator = KafkaPropagator()


# Convenience functions for direct access
def inject_kafka_headers(headers: Dict[str, str]) -> None:
    """Inject trace_id into Kafka headers."""
    kafka_propagator.inject_kafka(headers)


def extract_kafka_headers(headers: Dict[str, str]) -> Optional[str]:
    """Extract trace_id from Kafka headers."""
    return kafka_propagator.extract_kafka(headers)


def get_traced_kafka_headers(base_headers: Optional[Dict[str, str]] = None) -> Dict[str, str]:
    """Get Kafka headers with trace_id injected."""
    result_headers = base_headers.copy() if base_headers else {}
    kafka_propagator.inject_kafka(result_headers)
    return result_headers


def inject_aiokafka_message_headers(headers: Optional[List[Tuple[str, bytes]]] = None) -> List[Tuple[str, bytes]]:
    """Get aiokafka headers with trace_id injected."""
    return kafka_propagator.inject_aiokafka_headers(headers)


def extract_aiokafka_message_headers(headers: Optional[List[Tuple[str, bytes]]]) -> Optional[str]:
    """Extract trace_id from aiokafka headers."""
    return kafka_propagator.extract_aiokafka_headers(headers)