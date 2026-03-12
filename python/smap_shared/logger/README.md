# Enhanced Python Logger with Trace Integration

Enhanced logging package migrated from service-specific implementations with automatic trace_id injection from the shared tracing library.

## Features

- **Automatic trace_id injection** from shared tracing context
- **Structured logging** with JSON output for production environments
- **Colored console output** for development environments
- **Backward compatibility** with existing service logger implementations
- **Request ID tracking** for request-specific logging (optional)
- **Cross-service consistency** with Go services
- **Performance optimized** with minimal overhead

## Quick Start

```python
from smap_shared.logger import Logger, LoggerConfig, LogLevel
from smap_shared.tracing.context import set_trace_id, generate_trace_id

# Initialize logger
config = LoggerConfig(
    level=LogLevel.INFO,
    enable_trace_id=True,
    service_name="my-service"
)
logger = Logger(config)

# Set trace context
trace_id = generate_trace_id()
set_trace_id(trace_id)

# Log with automatic trace_id injection
logger.info("Processing request")  # Includes trace_id automatically
logger.error("Error occurred")     # Includes trace_id automatically
```

## Configuration

### Basic Configuration

```python
from smap_shared.logger import LoggerConfig, LogLevel

# Development configuration
dev_config = LoggerConfig(
    level=LogLevel.DEBUG,
    colorize=True,
    json_output=False,
    enable_trace_id=True,
    service_name="my-service"
)

# Production configuration
prod_config = LoggerConfig(
    level=LogLevel.INFO,
    colorize=False,
    json_output=True,
    enable_trace_id=True,
    service_name="my-service"
)
```

### Environment-Based Configuration

```python
import os
from smap_shared.logger import LoggerConfig, LogLevel

def create_logger_config():
    debug = os.getenv("DEBUG", "false").lower() == "true"
    service_name = os.getenv("SERVICE_NAME", "python-service")
    
    return LoggerConfig(
        level=LogLevel.DEBUG if debug else LogLevel.INFO,
        json_output=not debug,
        colorize=debug,
        service_name=service_name,
        enable_trace_id=True,
    )
```

## Usage Examples

### Basic Logging

```python
logger.debug("Debug information")
logger.info("Information message")
logger.warning("Warning message")
logger.error("Error message")
logger.critical("Critical message")
logger.exception("Exception with traceback")
```

### Structured Logging

```python
# JSON output with additional context
logger.info(
    "User action performed",
    extra={
        "user_id": "user123",
        "action": "create_project",
        "project_id": "proj456",
        "duration_ms": 250,
        "success": True
    }
)
```

### Request Tracking

```python
# Enable request_id in configuration
config = LoggerConfig(enable_request_id=True)
logger = Logger(config)

# Use request context
with logger.request_context(request_id="req_123"):
    logger.info("Processing request")  # Includes both trace_id and request_id
```

### Bound Logger

```python
# Create logger with persistent context
bound_logger = logger.bind(component="auth", module="jwt")
bound_logger.info("JWT token validated")
bound_logger.info("User permissions checked")
```

## Output Formats

### Development Mode (Colored Console)

```
2026-03-12 13:52:38 | INFO     | 550e8400-e29b-41d4-a716-446655440000 | service.py:42 - Processing request
```

### Production Mode (JSON)

```json
{
  "timestamp": "Thu, 12 Mar 2026 13:52:38 +0700",
  "level": "info",
  "caller": "service.py:42",
  "message": "Processing request",
  "service": "my-service",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### With Request ID

```
2026-03-12 13:52:38 | INFO     | 550e8400-e29b-41d4-a716-446655440000 | req_123 | service.py:42 - Processing request
```

## Migration from Service-Specific Loggers

### From scapper-srv

```python
# Old
from app.logger import setup_logging, trace_context

# New
from smap_shared.logger.compat import setup_logging, trace_context

# Same API, enhanced functionality
logger = setup_logging(debug=True, service_name="scapper-srv")
```

### From analysis-srv

```python
# Old
from pkg.logger import Logger, LoggerConfig

# New
from smap_shared.logger import Logger, LoggerConfig
# or for gradual migration
from smap_shared.logger.compat import LoggerCompat
```

See [MIGRATION.md](./MIGRATION.md) for detailed migration instructions.

## Integration with FastAPI

```python
from fastapi import FastAPI, Request, Response
from smap_shared.logger import Logger, LoggerConfig
from smap_shared.tracing.context import set_trace_id, generate_trace_id
from smap_shared.tracing.http import HTTPPropagator

app = FastAPI()
logger = Logger(LoggerConfig(service_name="api-service"))
propagator = HTTPPropagator()

@app.middleware("http")
async def trace_middleware(request: Request, call_next):
    # Extract or generate trace_id
    trace_id = propagator.extract_http(dict(request.headers))
    if not trace_id:
        trace_id = generate_trace_id()
    
    # Set in context for automatic injection
    set_trace_id(trace_id)
    
    # Log request
    logger.info(f"Request: {request.method} {request.url.path}")
    
    response = await call_next(request)
    
    # Log response
    logger.info(f"Response: {response.status_code}")
    
    return response
```

## Performance Considerations

- **Minimal overhead**: Trace_id injection adds <0.1ms per log call
- **Efficient caching**: File path computation is cached for performance
- **Lazy evaluation**: JSON formatting only when needed
- **Memory efficient**: Context variables use minimal memory

## Error Handling

The logger gracefully handles various error conditions:

```python
# Invalid trace_id - generates new one and continues
set_trace_id("invalid-uuid")
logger.info("Message")  # Logs with new valid trace_id

# Missing trace_id - logs without trace_id field
logger.info("Message")  # Works fine, shows "-" placeholder

# Configuration errors - raises ValueError with clear message
try:
    LoggerConfig(level="INVALID")
except ValueError as e:
    print(f"Configuration error: {e}")
```

## Testing

```python
# Test configuration
from smap_shared.logger.config import DEFAULT_TESTING_CONFIG

test_logger = Logger(DEFAULT_TESTING_CONFIG)
test_logger.info("Test message")
```

## Dependencies

- `loguru`: High-performance logging library
- `smap_shared.tracing`: Shared tracing context management

## API Reference

### LoggerConfig

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `level` | `LogLevel` | `INFO` | Log level |
| `enable_console` | `bool` | `True` | Enable console output |
| `colorize` | `bool` | `True` | Enable colors (dev mode) |
| `json_output` | `bool` | `False` | Enable JSON output (prod mode) |
| `service_name` | `str` | `"python-service"` | Service identifier |
| `enable_trace_id` | `bool` | `True` | Enable trace_id injection |
| `enable_request_id` | `bool` | `False` | Enable request_id tracking |

### Logger Methods

- `debug(message, **kwargs)`: Log debug message
- `info(message, **kwargs)`: Log info message  
- `warning(message, **kwargs)`: Log warning message
- `error(message, **kwargs)`: Log error message
- `critical(message, **kwargs)`: Log critical message
- `exception(message, **kwargs)`: Log exception with traceback
- `get_trace_id()`: Get current trace_id
- `request_context(request_id)`: Context manager for request tracking
- `bind(**kwargs)`: Create bound logger with persistent context

## Examples

See [examples.py](./examples.py) for comprehensive usage examples covering:
- Basic usage
- Production JSON logging
- Request tracking
- Migration patterns
- Error handling
- Structured logging