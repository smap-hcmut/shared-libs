"""
PostgreSQL package with trace_id logging integration.

Database client with trace_id injection in query logs.
Enhanced version migrated from analysis-srv with automatic trace_id logging.
"""

from .client import TracedPostgresClient, PostgresDatabase
from .config import PostgresConfig
from .interfaces import IDatabaseInterface, IDatabase
from .constants import (
    DEFAULT_SCHEMA,
    DEFAULT_POOL_SIZE,
    DEFAULT_MAX_OVERFLOW,
    TRACE_LOG_FORMAT,
    QUERY_LOG_FORMAT,
)

__all__ = [
    # Main classes
    "TracedPostgresClient",
    "PostgresConfig",
    
    # Interfaces
    "IDatabaseInterface",
    
    # Backward compatibility
    "PostgresDatabase",
    "IDatabase",
    
    # Constants
    "DEFAULT_SCHEMA",
    "DEFAULT_POOL_SIZE", 
    "DEFAULT_MAX_OVERFLOW",
    "TRACE_LOG_FORMAT",
    "QUERY_LOG_FORMAT",
]