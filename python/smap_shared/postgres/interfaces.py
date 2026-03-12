"""Interface for PostgreSQL database operations with trace logging."""

from typing import AsyncGenerator, Protocol, runtime_checkable, Optional, Dict, Any

from sqlalchemy.ext.asyncio import AsyncSession


@runtime_checkable
class IDatabaseInterface(Protocol):
    """
    Protocol for database operations with trace logging support.
    
    Implementations are safe for concurrent use and include automatic
    trace_id injection in query logs.
    """

    async def get_session(self) -> AsyncGenerator[AsyncSession, None]:
        """
        Get database session.

        Yields:
            AsyncSession instance
        """
        ...

    async def health_check(self) -> bool:
        """
        Check database connectivity.

        Returns:
            True if database is healthy, False otherwise
        """
        ...

    async def close(self) -> None:
        """Close database connections and cleanup resources."""
        ...

    async def execute_raw(self, query: str, params: Optional[Dict[str, Any]] = None) -> Any:
        """
        Execute raw SQL query with trace logging.
        
        Args:
            query: SQL query string
            params: Query parameters (optional)
            
        Returns:
            Query result
        """
        ...

    async def get_pool_status(self) -> Dict[str, Any]:
        """
        Get connection pool status.
        
        Returns:
            Dictionary with pool statistics
        """
        ...


# Backward compatibility alias
IDatabase = IDatabaseInterface


__all__ = [
    "IDatabaseInterface",
    "IDatabase",  # Backward compatibility
]