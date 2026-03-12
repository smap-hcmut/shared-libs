"""Constants for Redis cache client with trace logging."""

# Redis Defaults
DEFAULT_HOST = "localhost"
DEFAULT_PORT = 6379
DEFAULT_DB = 0
DEFAULT_SSL = False
DEFAULT_ENCODING = "utf-8"
DEFAULT_DECODE_RESPONSES = True
DEFAULT_MAX_CONNECTIONS = 50
DEFAULT_SOCKET_TIMEOUT = 5
DEFAULT_SOCKET_CONNECT_TIMEOUT = 5
DEFAULT_SOCKET_KEEPALIVE = True
DEFAULT_HEALTH_CHECK_INTERVAL = 30

# Trace Logging Constants
TRACE_LOG_FORMAT = "trace_id={trace_id} operation={operation} key={key}"
OPERATION_LOG_FORMAT = "operation={operation} key={key}"

# Error Messages
ERROR_HOST_EMPTY = "host is required"
ERROR_INVALID_PORT = "port must be between 1 and 65535"
ERROR_INVALID_DB = "db must be between 0 and 15"
ERROR_INVALID_MAX_CONNECTIONS = "max_connections must be between 1 and 1000"
ERROR_INVALID_SOCKET_TIMEOUT = "socket_timeout must be between 1 and 300 seconds"
ERROR_CLIENT_NOT_INITIALIZED = "Redis client not initialized"

# Connection Limits
MIN_PORT = 1
MAX_PORT = 65535
MIN_DB = 0
MAX_DB = 15  # Redis default max databases
MIN_MAX_CONNECTIONS = 1
MAX_MAX_CONNECTIONS = 1000
MIN_SOCKET_TIMEOUT = 1
MAX_SOCKET_TIMEOUT = 300  # 5 minutes

# Redis Operation Names
REDIS_OPERATIONS = {
    "GET": "GET",
    "SET": "SET", 
    "DEL": "DEL",
    "EXISTS": "EXISTS",
    "EXPIRE": "EXPIRE",
    "TTL": "TTL",
    "INCR": "INCR",
    "DECR": "DECR",
    "MGET": "MGET",
    "MSET": "MSET",
    "PING": "PING",
    "INFO": "INFO",
    "LPUSH": "LPUSH",
    "RPOP": "RPOP",
    "LLEN": "LLEN",
}

# Connection Pool Settings
DEFAULT_RETRY_ON_TIMEOUT = True
DEFAULT_SOCKET_KEEPALIVE_OPTIONS = {}

# SSL Settings
DEFAULT_SSL_CERT_REQS = "required"
DEFAULT_SSL_CHECK_HOSTNAME = True

__all__ = [
    # Defaults
    "DEFAULT_HOST",
    "DEFAULT_PORT",
    "DEFAULT_DB",
    "DEFAULT_SSL",
    "DEFAULT_ENCODING",
    "DEFAULT_DECODE_RESPONSES",
    "DEFAULT_MAX_CONNECTIONS",
    "DEFAULT_SOCKET_TIMEOUT",
    "DEFAULT_SOCKET_CONNECT_TIMEOUT",
    "DEFAULT_SOCKET_KEEPALIVE",
    "DEFAULT_HEALTH_CHECK_INTERVAL",
    
    # Trace Logging
    "TRACE_LOG_FORMAT",
    "OPERATION_LOG_FORMAT",
    
    # Error Messages
    "ERROR_HOST_EMPTY",
    "ERROR_INVALID_PORT",
    "ERROR_INVALID_DB",
    "ERROR_INVALID_MAX_CONNECTIONS",
    "ERROR_INVALID_SOCKET_TIMEOUT",
    "ERROR_CLIENT_NOT_INITIALIZED",
    
    # Connection Limits
    "MIN_PORT",
    "MAX_PORT",
    "MIN_DB",
    "MAX_DB",
    "MIN_MAX_CONNECTIONS",
    "MAX_MAX_CONNECTIONS",
    "MIN_SOCKET_TIMEOUT",
    "MAX_SOCKET_TIMEOUT",
    
    # Redis Operations
    "REDIS_OPERATIONS",
    
    # Pool Settings
    "DEFAULT_RETRY_ON_TIMEOUT",
    "DEFAULT_SOCKET_KEEPALIVE_OPTIONS",
    
    # SSL Settings
    "DEFAULT_SSL_CERT_REQS",
    "DEFAULT_SSL_CHECK_HOSTNAME",
]