"""
Enhanced Redis client with trace_id logging integration.

Migrated from analysis-srv and enhanced with automatic trace_id injection
in operation logs for distributed tracing support.
"""

import json
from typing import Optional, Dict, Any, List

try:
    import redis.asyncio as aioredis
    from redis.asyncio import ConnectionPool
except ImportError:
    raise ImportError(
        "redis is required for the Redis client. Install with: pip install redis"
    )

try:
    from loguru import logger
except ImportError:
    import logging
    logger = logging.getLogger(__name__)

from .interfaces import ICacheInterface
from .config import RedisConfig
from .constants import (
    TRACE_LOG_FORMAT,
    OPERATION_LOG_FORMAT,
    ERROR_CLIENT_NOT_INITIALIZED,
)

# Import trace context from shared tracing library
try:
    from ..tracing.context import get_trace_id
except ImportError:
    # Fallback if tracing is not available
    def get_trace_id() -> Optional[str]:
        return None


class TracedRedisClient:
    """
    Redis cache client with trace_id logging integration.
    
    Enhanced version of the original RedisCache from analysis-srv with:
    - Automatic trace_id injection in operation logs
    - Operation logging format: "trace_id={uuid} operation={op} key={key}"
    - Graceful handling when no trace_id exists in context
    - Backward compatibility with existing interfaces
    - Async support with redis-py
    - JSON serialization support
    
    Usage:
        # Initialize client
        config = RedisConfig(host="localhost", port=6379)
        client = TracedRedisClient(config)
        
        # Use with trace context (automatic logging)
        from smap_shared.tracing import set_trace_id
        set_trace_id("550e8400-e29b-41d4-a716-446655440000")
        
        await client.set("user:123", {"name": "John"})
        # Logs: trace_id=550e8400-e29b-41d4-a716-446655440000 operation=SET key=user:123
        
        value = await client.get("user:123")
        # Logs: trace_id=550e8400-e29b-41d4-a716-446655440000 operation=GET key=user:123
    """
    
    def __init__(self, config: RedisConfig):
        """
        Initialize Redis client with trace logging.
        
        Args:
            config: RedisConfig instance
        """
        self.config = config
        self.client = None
        self.pool = None
        self._initialize_client()
    
    def _initialize_client(self) -> None:
        """Initialize Redis client with connection pool."""
        try:
            # Create connection pool
            pool_kwargs = {
                "host": self.config.host,
                "port": self.config.port,
                "db": self.config.db,
                "password": self.config.password,
                "username": self.config.username,
                "encoding": self.config.encoding,
                "decode_responses": self.config.decode_responses,
                "max_connections": self.config.max_connections,
                "socket_timeout": self.config.socket_timeout,
                "socket_connect_timeout": self.config.socket_connect_timeout,
                "socket_keepalive": self.config.socket_keepalive,
                "health_check_interval": self.config.health_check_interval,
            }
            
            if self.config.ssl:
                pool_kwargs["ssl"] = True
                pool_kwargs["ssl_cert_reqs"] = "required"
            
            self.pool = ConnectionPool(**pool_kwargs)
            
            # Create Redis client
            self.client = aioredis.Redis(connection_pool=self.pool)
            
        except Exception as e:
            logger.error(f"Failed to initialize Redis client: {e}")
            raise
    
    def _log_operation(self, operation: str, key: str, extra_info: Optional[str] = None) -> None:
        """
        Log Redis operation with trace_id in the specified format.
        
        Args:
            operation: Redis operation name (GET, SET, DEL, etc.)
            key: Redis key
            extra_info: Additional information to log (optional)
        """
        try:
            # Get current trace_id from context
            trace_id = get_trace_id()
            
            # Build log message
            if trace_id:
                log_message = f"trace_id={trace_id} operation={operation} key={key}"
            else:
                # Log without trace_id when not available (graceful handling)
                log_message = f"operation={operation} key={key}"
            
            # Add extra info if provided
            if extra_info:
                log_message += f" {extra_info}"
            
            logger.info(log_message)
            
        except Exception as e:
            # Don't let logging errors affect Redis operations
            logger.error(f"Failed to log Redis operation with trace_id: {e}")
    
    async def get(self, key: str) -> Optional[str]:
        """
        Get value by key with trace logging.
        
        Args:
            key: Cache key
            
        Returns:
            Value as string, or None if not found
        """
        try:
            self._log_operation("GET", key)
            return await self.client.get(key)
        except Exception as e:
            logger.error(f"Redis GET error for key '{key}': {e}")
            return None
    
    async def get_json(self, key: str) -> Optional[Any]:
        """
        Get value by key and deserialize from JSON with trace logging.
        
        Args:
            key: Cache key
            
        Returns:
            Deserialized value, or None if not found
        """
        value = await self.get(key)
        if value is None:
            return None
        
        try:
            return json.loads(value)
        except json.JSONDecodeError as e:
            logger.error(f"JSON decode error for key '{key}': {e}")
            return None
    
    async def set(self, key: str, value: Any, ttl: Optional[int] = None) -> bool:
        """
        Set key-value pair with optional TTL and trace logging.
        
        Args:
            key: Cache key
            value: Value to store (will be JSON serialized if not string)
            ttl: Time-to-live in seconds (optional)
            
        Returns:
            True if successful, False otherwise
        """
        try:
            # Serialize to JSON if not string
            if not isinstance(value, str):
                value = json.dumps(value)
            
            # Log operation with TTL info
            ttl_info = f"ttl={ttl}" if ttl else "ttl=none"
            self._log_operation("SET", key, ttl_info)
            
            if ttl:
                return await self.client.setex(key, ttl, value)
            else:
                return await self.client.set(key, value)
        except Exception as e:
            logger.error(f"Redis SET error for key '{key}': {e}")
            return False
    
    async def delete(self, key: str) -> bool:
        """
        Delete key with trace logging.
        
        Args:
            key: Cache key
            
        Returns:
            True if key was deleted, False otherwise
        """
        try:
            self._log_operation("DEL", key)
            result = await self.client.delete(key)
            return result > 0
        except Exception as e:
            logger.error(f"Redis DELETE error for key '{key}': {e}")
            return False
    
    async def exists(self, key: str) -> bool:
        """
        Check if key exists with trace logging.
        
        Args:
            key: Cache key
            
        Returns:
            True if key exists, False otherwise
        """
        try:
            self._log_operation("EXISTS", key)
            result = await self.client.exists(key)
            return result > 0
        except Exception as e:
            logger.error(f"Redis EXISTS error for key '{key}': {e}")
            return False
    
    async def expire(self, key: str, ttl: int) -> bool:
        """
        Set expiration time for key with trace logging.
        
        Args:
            key: Cache key
            ttl: Time-to-live in seconds
            
        Returns:
            True if successful, False otherwise
        """
        try:
            self._log_operation("EXPIRE", key, f"ttl={ttl}")
            return await self.client.expire(key, ttl)
        except Exception as e:
            logger.error(f"Redis EXPIRE error for key '{key}': {e}")
            return False
    
    async def ttl(self, key: str) -> int:
        """
        Get remaining TTL for key with trace logging.
        
        Args:
            key: Cache key
            
        Returns:
            Remaining TTL in seconds, -1 if no expiry, -2 if key doesn't exist
        """
        try:
            self._log_operation("TTL", key)
            return await self.client.ttl(key)
        except Exception as e:
            logger.error(f"Redis TTL error for key '{key}': {e}")
            return -2
    
    async def incr(self, key: str, amount: int = 1) -> Optional[int]:
        """
        Increment key by amount with trace logging.
        
        Args:
            key: Cache key
            amount: Amount to increment (default: 1)
            
        Returns:
            New value after increment, or None on error
        """
        try:
            self._log_operation("INCR", key, f"amount={amount}")
            return await self.client.incrby(key, amount)
        except Exception as e:
            logger.error(f"Redis INCR error for key '{key}': {e}")
            return None
    
    async def decr(self, key: str, amount: int = 1) -> Optional[int]:
        """
        Decrement key by amount with trace logging.
        
        Args:
            key: Cache key
            amount: Amount to decrement (default: 1)
            
        Returns:
            New value after decrement, or None on error
        """
        try:
            self._log_operation("DECR", key, f"amount={amount}")
            return await self.client.decrby(key, amount)
        except Exception as e:
            logger.error(f"Redis DECR error for key '{key}': {e}")
            return None
    
    async def mget(self, keys: List[str]) -> List[Optional[str]]:
        """
        Get multiple values by keys with trace logging.
        
        Args:
            keys: List of cache keys
            
        Returns:
            List of values (None for missing keys)
        """
        try:
            keys_str = ",".join(keys[:5])  # Log first 5 keys to avoid long logs
            if len(keys) > 5:
                keys_str += f"... ({len(keys)} total)"
            self._log_operation("MGET", keys_str)
            return await self.client.mget(keys)
        except Exception as e:
            logger.error(f"Redis MGET error: {e}")
            return [None] * len(keys)
    
    async def mset(self, mapping: Dict[str, Any]) -> bool:
        """
        Set multiple key-value pairs with trace logging.
        
        Args:
            mapping: Dictionary of key-value pairs
            
        Returns:
            True if successful, False otherwise
        """
        try:
            keys_str = ",".join(list(mapping.keys())[:5])  # Log first 5 keys
            if len(mapping) > 5:
                keys_str += f"... ({len(mapping)} total)"
            self._log_operation("MSET", keys_str)
            
            # Serialize values to JSON if needed
            serialized = {}
            for key, value in mapping.items():
                if not isinstance(value, str):
                    serialized[key] = json.dumps(value)
                else:
                    serialized[key] = value
            
            return await self.client.mset(serialized)
        except Exception as e:
            logger.error(f"Redis MSET error: {e}")
            return False
    
    async def health_check(self) -> bool:
        """
        Check Redis connectivity with trace logging.
        
        Returns:
            True if Redis is healthy, False otherwise
        """
        try:
            self._log_operation("PING", "health_check")
            return await self.client.ping()
        except Exception as e:
            logger.error(f"Redis health check failed: {e}")
            return False
    
    async def close(self) -> None:
        """Close Redis connection and cleanup resources."""
        if self.client:
            await self.client.close()
            logger.info("Redis connection closed")
        
        if self.pool:
            await self.pool.disconnect()
            logger.info("Redis connection pool closed")
    
    async def get_info(self) -> Dict[str, Any]:
        """
        Get Redis server info with trace logging.
        
        Returns:
            Dictionary with Redis server information
        """
        try:
            self._log_operation("INFO", "server_info")
            return await self.client.info()
        except Exception as e:
            logger.error(f"Redis INFO error: {e}")
            return {}
    
    # Additional Redis operations with trace logging
    
    async def lpush(self, key: str, *values: Any) -> Optional[int]:
        """
        Push values to the left of a list with trace logging.
        
        Args:
            key: List key
            *values: Values to push
            
        Returns:
            New length of list, or None on error
        """
        try:
            self._log_operation("LPUSH", key, f"count={len(values)}")
            # Serialize values if needed
            serialized_values = []
            for value in values:
                if not isinstance(value, str):
                    serialized_values.append(json.dumps(value))
                else:
                    serialized_values.append(value)
            return await self.client.lpush(key, *serialized_values)
        except Exception as e:
            logger.error(f"Redis LPUSH error for key '{key}': {e}")
            return None
    
    async def rpop(self, key: str) -> Optional[str]:
        """
        Pop value from the right of a list with trace logging.
        
        Args:
            key: List key
            
        Returns:
            Popped value, or None if list is empty or key doesn't exist
        """
        try:
            self._log_operation("RPOP", key)
            return await self.client.rpop(key)
        except Exception as e:
            logger.error(f"Redis RPOP error for key '{key}': {e}")
            return None
    
    async def llen(self, key: str) -> int:
        """
        Get length of a list with trace logging.
        
        Args:
            key: List key
            
        Returns:
            Length of list, or 0 if key doesn't exist
        """
        try:
            self._log_operation("LLEN", key)
            return await self.client.llen(key)
        except Exception as e:
            logger.error(f"Redis LLEN error for key '{key}': {e}")
            return 0


# Backward compatibility alias
RedisCache = TracedRedisClient


__all__ = [
    "TracedRedisClient",
    "RedisCache",  # Backward compatibility
]