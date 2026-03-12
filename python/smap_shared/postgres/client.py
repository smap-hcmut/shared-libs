"""
Enhanced PostgreSQL client with trace_id logging integration.

Migrated from analysis-srv and enhanced with automatic trace_id injection
in query logs for distributed tracing support.
"""

import json
from contextlib import asynccontextmanager
from typing import AsyncGenerator, Optional, Dict, Any
from pathlib import Path

from sqlalchemy import text, event
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.pool import NullPool
from sqlalchemy.engine import Engine

try:
    from loguru import logger
except ImportError:
    import logging
    logger = logging.getLogger(__name__)

from .interfaces import IDatabaseInterface
from .config import PostgresConfig
from .constants import (
    DEFAULT_SCHEMA,
    ERROR_DATABASE_NOT_INITIALIZED,
    ERROR_DATABASE_URL_EMPTY,
    ERROR_INVALID_DATABASE_URL,
)

# Import trace context from shared tracing library
try:
    from ..tracing.context import get_trace_id
except ImportError:
    # Fallback if tracing is not available
    def get_trace_id() -> Optional[str]:
        return None


class TracedPostgresClient:
    """
    PostgreSQL database client with trace_id logging integration.
    
    Enhanced version of the original PostgresDatabase from analysis-srv with:
    - Automatic trace_id injection in query logs
    - Database logging format: "trace_id={uuid} query={sql}"
    - Graceful handling when no trace_id exists in context
    - Backward compatibility with existing interfaces
    - Async support with SQLAlchemy and asyncpg
    
    Usage:
        # Initialize client
        config = PostgresConfig(database_url="postgresql+asyncpg://...")
        client = TracedPostgresClient(config)
        
        # Use with trace context (automatic logging)
        from smap_shared.tracing import set_trace_id
        set_trace_id("550e8400-e29b-41d4-a716-446655440000")
        
        async with client.get_session() as session:
            result = await session.execute(text("SELECT * FROM users"))
            # Logs: trace_id=550e8400-e29b-41d4-a716-446655440000 query=SELECT * FROM users
    """
    
    def __init__(self, config: PostgresConfig):
        """
        Initialize PostgreSQL client with trace logging.
        
        Args:
            config: PostgresConfig instance
        """
        self.config = config
        self.engine = None
        self.session_factory = None
        self._initialize_engine()
        self._setup_query_logging()
    
    def _initialize_engine(self) -> None:
        """Initialize async database engine and session factory."""
        try:
            # Ensure URL uses asyncpg driver
            url = self.config.database_url
            if url.startswith("postgresql://") or url.startswith("postgres://"):
                url = url.replace("postgresql://", "postgresql+asyncpg://", 1)
                url = url.replace("postgres://", "postgresql+asyncpg://", 1)
            
            # Create async engine with connection pooling
            engine_kwargs = {
                "echo": self.config.echo,
                "echo_pool": self.config.echo_pool,
                "pool_pre_ping": self.config.pool_pre_ping,
                "pool_recycle": self.config.pool_recycle,
            }
            
            # Use NullPool if echo is enabled (debug mode)
            if self.config.echo:
                engine_kwargs["poolclass"] = NullPool
            else:
                engine_kwargs["pool_size"] = self.config.pool_size
                engine_kwargs["max_overflow"] = self.config.max_overflow
            
            self.engine = create_async_engine(url, **engine_kwargs)
            
            # Create session factory
            self.session_factory = async_sessionmaker(
                bind=self.engine,
                class_=AsyncSession,
                expire_on_commit=False,  # Keep objects accessible after commit
                autoflush=True,
                autocommit=False,
            )
            
        except Exception as e:
            logger.error(f"Failed to initialize PostgreSQL engine: {e}")
            raise
    
    def _setup_query_logging(self) -> None:
        """Setup query logging with trace_id injection."""
        if not self.engine:
            return
        
        try:
            @event.listens_for(self.engine.sync_engine, "before_cursor_execute")
            def log_query_with_trace(conn, cursor, statement, parameters, context, executemany):
                """Log SQL queries with trace_id in the specified format."""
                try:
                    # Get current trace_id from context
                    trace_id = get_trace_id()
                    
                    # Format the query for logging (remove extra whitespace)
                    clean_query = " ".join(statement.split())
                    
                    # Log with trace_id if available
                    if trace_id:
                        log_message = f"trace_id={trace_id} query={clean_query}"
                    else:
                        # Log without trace_id when not available (graceful handling)
                        log_message = f"query={clean_query}"
                    
                    logger.info(log_message)
                    
                except Exception as e:
                    # Don't let logging errors affect database operations
                    logger.error(f"Failed to log query with trace_id: {e}")
        except Exception as e:
            # Gracefully handle event setup failures (e.g., during testing)
            logger.warning(f"Failed to setup query logging events: {e}")
            logger.info("Query logging will be disabled, but database operations will continue normally")
    
    @asynccontextmanager
    async def get_session(self) -> AsyncGenerator[AsyncSession, None]:
        """
        Get async database session with automatic cleanup.
        
        Yields:
            AsyncSession instance
            
        Raises:
            RuntimeError: If database not initialized
        """
        if not self.session_factory:
            raise RuntimeError(ERROR_DATABASE_NOT_INITIALIZED)
        
        async with self.session_factory() as session:
            try:
                # Set search_path to use the configured schema
                if self.config.schema and self.config.schema != DEFAULT_SCHEMA:
                    await session.execute(
                        text(f"SET search_path TO {self.config.schema}, {DEFAULT_SCHEMA}")
                    )
                yield session
            except Exception as e:
                logger.error(f"Database session error: {e}")
                await session.rollback()
                raise
            finally:
                await session.close()
    
    async def health_check(self) -> bool:
        """
        Check database connectivity.
        
        Returns:
            True if database is healthy, False otherwise
        """
        try:
            async with self.get_session() as session:
                result = await session.execute(text("SELECT 1"))
                return result.scalar() == 1
        except Exception as e:
            logger.error(f"Database health check failed: {e}")
            return False
    
    async def close(self) -> None:
        """Close database engine and cleanup resources."""
        if self.engine:
            await self.engine.dispose()
            logger.info("PostgreSQL engine closed")
    
    async def execute_raw(self, query: str, params: Optional[Dict[str, Any]] = None) -> Any:
        """
        Execute raw SQL query with trace logging.
        
        Args:
            query: SQL query string
            params: Query parameters (optional)
            
        Returns:
            Query result
        """
        async with self.get_session() as session:
            result = await session.execute(text(query), params or {})
            await session.commit()
            return result
    
    async def get_pool_status(self) -> Dict[str, Any]:
        """
        Get connection pool status.
        
        Returns:
            Dictionary with pool statistics
        """
        if not self.engine or not hasattr(self.engine.pool, "size"):
            return {"error": "Pool not available"}
        
        pool = self.engine.pool
        return {
            "pool_size": pool.size(),
            "checked_in": pool.checkedin(),
            "checked_out": pool.checkedout(),
            "overflow": pool.overflow(),
            "total": pool.size() + pool.overflow(),
        }
    
    # Backward compatibility methods (matching original interface)
    
    async def get_connection(self) -> AsyncGenerator[AsyncSession, None]:
        """
        Alias for get_session() for backward compatibility.
        
        Yields:
            AsyncSession instance
        """
        async with self.get_session() as session:
            yield session
    
    def get_engine(self):
        """
        Get the SQLAlchemy engine.
        
        Returns:
            SQLAlchemy async engine
        """
        return self.engine
    
    def get_session_factory(self):
        """
        Get the session factory.
        
        Returns:
            SQLAlchemy async session factory
        """
        return self.session_factory


# Backward compatibility alias
PostgresDatabase = TracedPostgresClient


__all__ = [
    "TracedPostgresClient",
    "PostgresDatabase",  # Backward compatibility
]