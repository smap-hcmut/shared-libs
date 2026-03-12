"""Interface for Redis cache operations with trace logging."""

from typing import Optional, Protocol, runtime_checkable, Dict, Any, List


@runtime_checkable
class ICacheInterface(Protocol):
    """
    Protocol for cache operations with trace logging support.
    
    Implementations are safe for concurrent use and include automatic
    trace_id injection in operation logs.
    """

    async def get(self, key: str) -> Optional[str]:
        """Get value by key with trace logging."""
        ...

    async def get_json(self, key: str) -> Optional[Any]:
        """Get value by key and deserialize from JSON with trace logging."""
        ...

    async def set(self, key: str, value: Any, ttl: Optional[int] = None) -> bool:
        """Set key-value pair with optional TTL and trace logging."""
        ...

    async def delete(self, key: str) -> bool:
        """Delete key with trace logging."""
        ...

    async def exists(self, key: str) -> bool:
        """Check if key exists with trace logging."""
        ...

    async def expire(self, key: str, ttl: int) -> bool:
        """Set expiration time for key with trace logging."""
        ...

    async def ttl(self, key: str) -> int:
        """Get remaining TTL for key with trace logging."""
        ...

    async def incr(self, key: str, amount: int = 1) -> Optional[int]:
        """Increment key by amount with trace logging."""
        ...

    async def decr(self, key: str, amount: int = 1) -> Optional[int]:
        """Decrement key by amount with trace logging."""
        ...

    async def mget(self, keys: List[str]) -> List[Optional[str]]:
        """Get multiple values by keys with trace logging."""
        ...

    async def mset(self, mapping: Dict[str, Any]) -> bool:
        """Set multiple key-value pairs with trace logging."""
        ...

    async def health_check(self) -> bool:
        """Check Redis connectivity with trace logging."""
        ...

    async def close(self) -> None:
        """Close Redis connection and cleanup resources."""
        ...

    async def get_info(self) -> Dict[str, Any]:
        """Get Redis server info with trace logging."""
        ...


# Backward compatibility alias
ICache = ICacheInterface


__all__ = [
    "ICacheInterface",
    "ICache",  # Backward compatibility
]