# Python Logger Migration Guide

This guide helps migrate from service-specific logger implementations to the enhanced shared logger with automatic trace_id integration.

## Overview

The enhanced shared logger provides:
- **Automatic trace_id injection** from shared tracing context
- **Structured logging** with JSON output for production
- **Backward compatibility** with existing service implementations
- **Request ID tracking** for request-specific logging
- **Cross-service consistency** with Go services

## Migration Paths

### 1. Scapper-srv Migration

**Before (scapper-srv/app/logger.py):**
```python
from app.logger import setup_logging, trace_context, get_trace_id, set_trace_id

# Setup
logger = setup_logging(debug=True)

# Usage
with trace_context(trace_id="uuid"):
    logger.info("Processing")
```

**After (using shared library):**
```python
from smap_shared.logger.compat import setup_logging, trace_context
from smap_shared.tracing.context import get_trace_id, set_trace_id

# Setup (same API)
logger = setup_logging(debug=True, service_name="scapper-srv")

# Usage (same API)
with trace_context(trace_id="uuid"):
    logger.info("Processing")  # Automatic trace_id injection
```

### 2. Analysis-srv Migration

**Before (analysis-srv/pkg/logger/):**
```python
from pkg.logger import Logger, LoggerConfig, LogLevel

# Setup
config = LoggerConfig(level=LogLevel.INFO, enable_trace_id=True)
logger = Logger(config)

# Usage
with logger.trace_context(trace_id="uuid"):
    logger.info("Processing")
```

**After (using shared library):**
```python
from smap_shared.logger import Logger, LoggerConfig, LogLevel
from smap_shared.tracing.context import set_trace_id

# Setup (enhanced configuration)
config = LoggerConfig(
    level=LogLevel.INFO,
    enable_trace_id=True,
    service_name="analysis-srv"
)
logger = Logger(config)

# Usage (automatic trace_id from context)
set_trace_id("uuid")  # Set once in middleware/handler
logger.info("Processing")  # Automatic trace_id injection
```

**Alternative (using compatibility wrapper):**
```python
from smap_shared.logger.compat import LoggerCompat
from smap_shared.logger import LoggerConfig

# Same API as before
config = LoggerConfig(level="INFO", enable_trace_id=True)
logger = LoggerCompat(config)

with logger.trace_context(trace_id="uuid"):
    logger.info("Processing")
```

## Step-by-Step Migration

### Step 1: Update Dependencies

**pyproject.toml:**
```toml
[dependencies]
smap-shared-python = "^1.0.0"
# Remove: loguru (now included in shared library)
```

### Step 2: Update Imports

**Replace service-specific imports:**
```python
# Old
from pkg.logger import Logger, LoggerConfig
from app.logger import setup_logging

# New
from smap_shared.logger import Logger, LoggerConfig, LogLevel
from smap_shared.logger.compat import setup_logging  # For gradual migration
```

### Step 3: Update Configuration

**Enhanced configuration options:**
```python
from smap_shared.logger import LoggerConfig, LogLevel

# Development
config = LoggerConfig(
    level=LogLevel.DEBUG,
    colorize=True,
    json_output=False,
    enable_trace_id=True,
    service_name="your-service"
)

# Production
config = LoggerConfig(
    level=LogLevel.INFO,
    colorize=False,
    json_output=True,
    enable_trace_id=True,
    service_name="your-service"
)
```

### Step 4: Update Trace Context Usage

**Automatic trace_id injection:**
```python
from smap_shared.tracing.context import set_trace_id
from smap_shared.logger import Logger

# Set trace_id once (e.g., in middleware)
set_trace_id("550e8400-e29b-41d4-a716-446655440000")

# All subsequent log calls automatically include trace_id
logger.info("Processing request")  # Includes trace_id
logger.error("Error occurred")     # Includes trace_id
```

### Step 5: Update Middleware Integration

**FastAPI middleware example:**
```python
from fastapi import Request, Response
from smap_shared.tracing.context import set_trace_id, generate_trace_id
from smap_shared.tracing.http import HTTPPropagator
from smap_shared.logger import Logger

async def trace_middleware(request: Request, call_next):
    propagator = HTTPPropagator()
    
    # Extract or generate trace_id
    trace_id = propagator.extract_http(dict(request.headers))
    if not trace_id:
        trace_id = generate_trace_id()
    
    # Set in context for automatic injection
    set_trace_id(trace_id)
    
    response = await call_next(request)
    return response
```

## Configuration Options

### LoggerConfig Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `level` | `LogLevel` | `INFO` | Log level (DEBUG, INFO, WARNING, ERROR, CRITICAL) |
| `enable_console` | `bool` | `True` | Enable console output |
| `colorize` | `bool` | `True` | Enable colored output (development) |
| `json_output` | `bool` | `False` | Enable JSON structured output (production) |
| `service_name` | `str` | `"python-service"` | Service name for structured logging |
| `enable_trace_id` | `bool` | `True` | Enable automatic trace_id injection |
| `enable_request_id` | `bool` | `False` | Enable request_id tracking |

### Environment-Specific Configurations

```python
from smap_shared.logger.config import (
    DEFAULT_DEVELOPMENT_CONFIG,
    DEFAULT_PRODUCTION_CONFIG,
    DEFAULT_TESTING_CONFIG
)

# Use predefined configurations
logger = Logger(DEFAULT_PRODUCTION_CONFIG)
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
  "service": "your-service",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Best Practices

### 1. Service Initialization
```python
import os
from smap_shared.logger import Logger, LoggerConfig, LogLevel

def setup_service_logger():
    """Setup logger based on environment."""
    debug = os.getenv("DEBUG", "false").lower() == "true"
    service_name = os.getenv("SERVICE_NAME", "python-service")
    
    config = LoggerConfig(
        level=LogLevel.DEBUG if debug else LogLevel.INFO,
        json_output=not debug,
        colorize=debug,
        service_name=service_name,
        enable_trace_id=True,
    )
    
    return Logger(config)
```

### 2. Middleware Integration
```python
# Set trace_id early in request lifecycle
async def setup_tracing_middleware(request: Request, call_next):
    # Extract/generate trace_id
    trace_id = extract_or_generate_trace_id(request)
    set_trace_id(trace_id)
    
    # All subsequent logging automatically includes trace_id
    response = await call_next(request)
    return response
```

### 3. Structured Logging
```python
# Use extra fields for structured data
logger.info(
    "User action completed",
    extra={
        "user_id": "user123",
        "action": "create_project",
        "duration_ms": 250,
        "success": True
    }
)
```

### 4. Error Handling
```python
try:
    process_request()
except Exception as e:
    logger.exception("Request processing failed")
    logger.error(f"Error details: {str(e)}")
```

## Troubleshooting

### Common Issues

1. **Missing trace_id in logs**
   - Ensure `enable_trace_id=True` in configuration
   - Verify trace_id is set in context using `set_trace_id()`

2. **Import errors**
   - Ensure `smap-shared-python` is installed
   - Check import paths match new shared library structure

3. **Configuration errors**
   - Validate log level strings (use `LogLevel` enum)
   - Ensure service_name is non-empty string

### Validation Script
```python
from smap_shared.logger import Logger, LoggerConfig
from smap_shared.tracing.context import set_trace_id, generate_trace_id

# Test basic functionality
config = LoggerConfig(service_name="test")
logger = Logger(config)

trace_id = generate_trace_id()
set_trace_id(trace_id)

logger.info("Migration test successful")
print(f"Trace ID: {logger.get_trace_id()}")
```

## Rollback Plan

If issues occur, you can temporarily use the compatibility layer:

```python
# Minimal changes using compatibility wrapper
from smap_shared.logger.compat import LoggerCompat, setup_logging

# Keep existing code mostly unchanged
logger = setup_logging(debug=True)
# or
logger = LoggerCompat(your_existing_config)
```

This provides the same API as your existing implementation while gaining trace integration benefits.