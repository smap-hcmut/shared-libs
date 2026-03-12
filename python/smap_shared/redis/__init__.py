"""
Redis package with trace context integration.

Redis client with trace_id logging and context awareness.
Enhanced version migrated from analysis-srv with automatic trace_id logging.
"""

from .client import TracedRedisClient, RedisCache
from .config import RedisConfig
from .interfaces import ICacheInterface, ICache
from .constants import (
    DEFAULT_HOST,
    DEFAULT_PORT,
    DEFAULT_DB,
    DEFAULT_MAX_CONNECTIONS,
    TRACE_LOG_FORMAT,
    OPERATION_LOG_FORMAT,
)

__all__ = [
    # Main classes
    "TracedRedisClient",
    "RedisConfig",
    
    # Interfaces
    "ICacheInterface",
    
    # Backward compatibility
    "RedisCache",
    "ICache",
    
    # Constants
    "DEFAULT_HOST",
    "DEFAULT_PORT",
    "DEFAULT_DB",
    "DEFAULT_MAX_CONNECTIONS",
    "TRACE_LOG_FORMAT",
    "OPERATION_LOG_FORMAT",
]