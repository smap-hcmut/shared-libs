"""Configuration for Redis cache with trace logging."""

from dataclasses import dataclass
from typing import Optional

from .constants import (
    DEFAULT_HOST,
    DEFAULT_PORT,
    DEFAULT_DB,
    DEFAULT_SSL,
    DEFAULT_ENCODING,
    DEFAULT_DECODE_RESPONSES,
    DEFAULT_MAX_CONNECTIONS,
    DEFAULT_SOCKET_TIMEOUT,
    DEFAULT_SOCKET_CONNECT_TIMEOUT,
    DEFAULT_SOCKET_KEEPALIVE,
    DEFAULT_HEALTH_CHECK_INTERVAL,
    ERROR_HOST_EMPTY,
    ERROR_INVALID_PORT,
    ERROR_INVALID_DB,
    ERROR_INVALID_MAX_CONNECTIONS,
    ERROR_INVALID_SOCKET_TIMEOUT,
    MIN_PORT,
    MAX_PORT,
    MIN_DB,
    MAX_DB,
    MIN_MAX_CONNECTIONS,
    MAX_MAX_CONNECTIONS,
    MIN_SOCKET_TIMEOUT,
    MAX_SOCKET_TIMEOUT,
)


@dataclass
class RedisConfig:
    """
    Configuration for Redis cache with trace logging support.

    Attributes:
        host: Redis host (default: localhost)
        port: Redis port (default: 6379)
        db: Redis database number (default: 0)
        password: Redis password (optional)
        username: Redis username (optional, Redis 6+)
        ssl: Enable SSL/TLS (default: False)
        encoding: String encoding (default: utf-8)
        decode_responses: Decode responses to strings (default: True)
        max_connections: Max connections in pool (default: 50)
        socket_timeout: Socket timeout in seconds (default: 5)
        socket_connect_timeout: Socket connect timeout (default: 5)
        socket_keepalive: Enable TCP keepalive (default: True)
        health_check_interval: Health check interval in seconds (default: 30)
        enable_trace_logging: Enable trace_id injection in operation logs (default: True)
    """

    host: str = DEFAULT_HOST
    port: int = DEFAULT_PORT
    db: int = DEFAULT_DB
    password: Optional[str] = None
    username: Optional[str] = None
    ssl: bool = DEFAULT_SSL
    encoding: str = DEFAULT_ENCODING
    decode_responses: bool = DEFAULT_DECODE_RESPONSES
    max_connections: int = DEFAULT_MAX_CONNECTIONS
    socket_timeout: int = DEFAULT_SOCKET_TIMEOUT
    socket_connect_timeout: int = DEFAULT_SOCKET_CONNECT_TIMEOUT
    socket_keepalive: bool = DEFAULT_SOCKET_KEEPALIVE
    health_check_interval: int = DEFAULT_HEALTH_CHECK_INTERVAL
    enable_trace_logging: bool = True

    def __post_init__(self):
        """Validate configuration."""
        if not self.host:
            raise ValueError(ERROR_HOST_EMPTY)
        if self.port < MIN_PORT or self.port > MAX_PORT:
            raise ValueError(ERROR_INVALID_PORT)
        if self.db < MIN_DB or self.db > MAX_DB:
            raise ValueError(ERROR_INVALID_DB)
        if self.max_connections < MIN_MAX_CONNECTIONS or self.max_connections > MAX_MAX_CONNECTIONS:
            raise ValueError(ERROR_INVALID_MAX_CONNECTIONS)
        if self.socket_timeout < MIN_SOCKET_TIMEOUT or self.socket_timeout > MAX_SOCKET_TIMEOUT:
            raise ValueError(ERROR_INVALID_SOCKET_TIMEOUT)

    @classmethod
    def from_url(
        cls,
        redis_url: str,
        **kwargs
    ) -> "RedisConfig":
        """
        Create configuration from Redis URL.
        
        Args:
            redis_url: Redis connection URL (redis://host:port/db)
            **kwargs: Additional configuration options
            
        Returns:
            RedisConfig instance
            
        Example:
            config = RedisConfig.from_url("redis://localhost:6379/0")
        """
        # Parse Redis URL
        # Format: redis://[username:password@]host:port/db
        import urllib.parse
        
        parsed = urllib.parse.urlparse(redis_url)
        
        host = parsed.hostname or DEFAULT_HOST
        port = parsed.port or DEFAULT_PORT
        db = int(parsed.path.lstrip('/')) if parsed.path and parsed.path != '/' else DEFAULT_DB
        
        username = parsed.username
        password = parsed.password
        
        # SSL detection
        ssl = parsed.scheme == "rediss"
        
        return cls(
            host=host,
            port=port,
            db=db,
            username=username,
            password=password,
            ssl=ssl,
            **kwargs
        )

    def get_connection_url(self) -> str:
        """
        Get Redis connection URL.
        
        Returns:
            Redis connection URL
        """
        scheme = "rediss" if self.ssl else "redis"
        
        # Build auth part
        auth_part = ""
        if self.username and self.password:
            auth_part = f"{self.username}:{self.password}@"
        elif self.password:
            auth_part = f":{self.password}@"
        
        return f"{scheme}://{auth_part}{self.host}:{self.port}/{self.db}"

    def is_ssl_enabled(self) -> bool:
        """
        Check if SSL is enabled.
        
        Returns:
            True if SSL is enabled
        """
        return self.ssl

    def has_auth(self) -> bool:
        """
        Check if authentication is configured.
        
        Returns:
            True if username or password is set
        """
        return bool(self.username or self.password)

    def get_pool_kwargs(self) -> dict:
        """
        Get connection pool keyword arguments.
        
        Returns:
            Dictionary of pool configuration
        """
        pool_kwargs = {
            "host": self.host,
            "port": self.port,
            "db": self.db,
            "password": self.password,
            "username": self.username,
            "encoding": self.encoding,
            "decode_responses": self.decode_responses,
            "max_connections": self.max_connections,
            "socket_timeout": self.socket_timeout,
            "socket_connect_timeout": self.socket_connect_timeout,
            "socket_keepalive": self.socket_keepalive,
            "health_check_interval": self.health_check_interval,
        }
        
        if self.ssl:
            pool_kwargs["ssl"] = True
            pool_kwargs["ssl_cert_reqs"] = "required"
        
        return pool_kwargs


__all__ = [
    "RedisConfig",
]