# SMAP Python Shared Libraries

Python implementation of shared libraries for SMAP services with distributed tracing support.

## Installation

```bash
pip install smap-shared-python
```

## Packages

### Core Tracing
- `smap_shared.tracing` - Core trace_id management and propagation
- `TraceContext` - Trace context management with contextvars
- `HTTPPropagator` - X-Trace-Id header management for HTTP requests
- `KafkaPropagator` - Trace header management for Kafka messages
- `FastAPIMiddleware` - Ready-to-use FastAPI middleware

### Enhanced Shared Packages
- `smap_shared.logger` - Enhanced logging with automatic trace_id injection
- `smap_shared.http` - HTTP client with automatic trace propagation
- `smap_shared.kafka` - Kafka producer/consumer with trace headers
- `smap_shared.redis` - Redis client with trace context
- `smap_shared.postgres` - PostgreSQL client with trace logging

## Usage

```python
from smap_shared.tracing import TraceContext, HTTPPropagator
from smap_shared.logger import Logger
from smap_shared.http import TracedHTTPClient

# Initialize components
trace_context = TraceContext()
logger = Logger()
http_client = TracedHTTPClient()

# Set trace context
trace_id = trace_context.generate_trace_id()
trace_context.set_trace_id(trace_id)

# Automatic trace propagation
async def make_request():
    response = await http_client.get("http://api.example.com/data")
    logger.info("Request completed", extra={"response_status": response.status_code})
```

## FastAPI Integration

```python
from fastapi import FastAPI
from smap_shared.tracing import trace_middleware

app = FastAPI()
app.middleware("http")(trace_middleware)

@app.get("/api/data")
async def get_data():
    # trace_id automatically available in context
    return {"message": "Hello World"}
```

## Migration from Service Packages

Replace service-specific imports:
```python
# Before
from analysis_srv.pkg.logger import Logger
from scapper_srv.app.logger import logger

# After
from smap_shared.logger import Logger
```

## Testing

```bash
pytest
pytest --cov=smap_shared
pytest --cov=smap_shared --cov-report=html
```