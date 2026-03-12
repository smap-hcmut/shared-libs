"""Constants for PostgreSQL database client with trace logging."""

# PostgreSQL Defaults
DEFAULT_POOL_SIZE = 20
DEFAULT_MAX_OVERFLOW = 10
DEFAULT_POOL_RECYCLE = 3600
DEFAULT_POOL_PRE_PING = True
DEFAULT_ECHO = False
DEFAULT_ECHO_POOL = False
DEFAULT_SCHEMA = "schema_analysis"

# Trace Logging Constants
TRACE_LOG_FORMAT = "trace_id={trace_id} query={query}"
QUERY_LOG_FORMAT = "query={query}"

# Error Messages
ERROR_DATABASE_URL_EMPTY = "database_url is required"
ERROR_INVALID_DATABASE_URL = "database_url must be a PostgreSQL connection string"
ERROR_POOL_SIZE_POSITIVE = "pool_size must be > 0"
ERROR_MAX_OVERFLOW_NON_NEGATIVE = "max_overflow must be >= 0"
ERROR_POOL_RECYCLE_POSITIVE = "pool_recycle must be > 0"
ERROR_SCHEMA_EMPTY = "schema cannot be empty"
ERROR_DATABASE_NOT_INITIALIZED = "Database not initialized"

# Connection Constants
ASYNCPG_DRIVER_PREFIX = "postgresql+asyncpg://"
POSTGRES_PREFIXES = ("postgresql://", "postgres://", "postgresql+asyncpg://")

# Pool Configuration
MIN_POOL_SIZE = 1
MAX_POOL_SIZE = 100
MIN_MAX_OVERFLOW = 0
MAX_MAX_OVERFLOW = 50
MIN_POOL_RECYCLE = 60  # 1 minute
MAX_POOL_RECYCLE = 86400  # 24 hours

__all__ = [
    # Defaults
    "DEFAULT_POOL_SIZE",
    "DEFAULT_MAX_OVERFLOW", 
    "DEFAULT_POOL_RECYCLE",
    "DEFAULT_POOL_PRE_PING",
    "DEFAULT_ECHO",
    "DEFAULT_ECHO_POOL",
    "DEFAULT_SCHEMA",
    
    # Trace Logging
    "TRACE_LOG_FORMAT",
    "QUERY_LOG_FORMAT",
    
    # Error Messages
    "ERROR_DATABASE_URL_EMPTY",
    "ERROR_INVALID_DATABASE_URL",
    "ERROR_POOL_SIZE_POSITIVE",
    "ERROR_MAX_OVERFLOW_NON_NEGATIVE",
    "ERROR_POOL_RECYCLE_POSITIVE",
    "ERROR_SCHEMA_EMPTY",
    "ERROR_DATABASE_NOT_INITIALIZED",
    
    # Connection Constants
    "ASYNCPG_DRIVER_PREFIX",
    "POSTGRES_PREFIXES",
    
    # Pool Limits
    "MIN_POOL_SIZE",
    "MAX_POOL_SIZE",
    "MIN_MAX_OVERFLOW",
    "MAX_MAX_OVERFLOW",
    "MIN_POOL_RECYCLE",
    "MAX_POOL_RECYCLE",
]