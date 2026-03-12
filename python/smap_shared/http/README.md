# SMAP HTTP Client with Trace Injection

This package provides HTTP client and middleware implementations with automatic trace_id propagation for distributed tracing across SMAP services.

## Features

- **TracedHTTPClient**: HTTP client wrapper with automatic X-Trace-Id header injection
- **TracingMiddleware**: FastAPI middleware for automatic trace_id extraction and generation
- **Dual Backend Support**: Works with both httpx (async) and requests (sync)
- **Configurable**: Timeout, retries, base headers, and tracing components
- **Error Handling**: Graceful handling of missing/invalid trace_ids
- **Cross-Service Compatibility**: Works seamlessly with Go services

## Quick Start

### HTTP Client Usage

```python
from smap_shared.http import TracedHTTPClient, create_async_client
from smap_shared.tracing import set_trace_id, generate_trace_id

# Set trace_id in context (usually done by middleware)
set_trace_id(generate_trace_id())

# Async usage
async with create_async_client() as client:
    response_body, status_code = await client.get(
        "https://api.example.com/data",
        headers={"Authorization": "Bearer token"}
    )

# Sync usage
client = TracedHTTPClient(use_async=False)
try:
    response_body, status_code = client.get_sync(
        "https://api.example.com/data",
        headers={"Authorization": "Bearer token"}
    )
finally:
    client.close()
```

### FastAPI Middleware Usage

```python
from fastapi import FastAPI
from smap_shared.http import setup_tracing_middleware

app = FastAPI()

# Add tracing middleware
setup_tracing_middleware(app)

@app.get("/api/data")
async def get_data():
    # trace_id is automatically available in context
    from smap_shared.tracing import get_trace_id
    current_trace_id = get_trace_id()
    
    # Make outbound calls with automatic trace propagation
    async with create_async_client() as client:
        response_body, status_code = await client.get(
            "http://other-service/api/endpoint"
        )
    
    return {"data": "response"}
```

## API Reference

### TracedHTTPClient

The main HTTP client class with automatic trace_id injection.

#### Constructor

```python
TracedHTTPClient(
    use_async: bool = True,
    timeout: float = 30.0,
    retries: int = 3,
    base_headers: Optional[Dict[str, str]] = None,
    trace_context: Optional[TraceContext] = None,
    http_propagator: Optional[HTTPPropagator] = None,
    **client_kwargs
)
```

**Parameters:**
- `use_async`: Whether to use async httpx client (True) or sync requests (False)
- `timeout`: Request timeout in seconds
- `retries`: Number of retry attempts for failed requests
- `base_headers`: Default headers to include in all requests
- `trace_context`: TraceContext instance (uses global if None)
- `http_propagator`: HTTPPropagator instance (uses global if None)
- `**client_kwargs`: Additional arguments passed to underlying client

#### Async Methods

```python
async def get(url, headers=None, params=None, **kwargs) -> Tuple[bytes, int]
async def post(url, data=None, json=None, headers=None, **kwargs) -> Tuple[bytes, int]
async def put(url, data=None, json=None, headers=None, **kwargs) -> Tuple[bytes, int]
async def delete(url, headers=None, **kwargs) -> Tuple[bytes, int]
```

#### Sync Methods

```python
def get_sync(url, headers=None, params=None, **kwargs) -> Tuple[bytes, int]
def post_sync(url, data=None, json=None, headers=None, **kwargs) -> Tuple[bytes, int]
def put_sync(url, data=None, json=None, headers=None, **kwargs) -> Tuple[bytes, int]
def delete_sync(url, headers=None, **kwargs) -> Tuple[bytes, int]
```

All methods return a tuple of `(response_body: bytes, status_code: int)`.

### TracingMiddleware

FastAPI middleware for automatic trace_id management.

```python
TracingMiddleware(
    app,
    trace_context: Optional[TraceContext] = None,
    http_propagator: Optional[HTTPPropagator] = None,
    log_trace_extraction: bool = True,
)
```

**Features:**
- Extracts trace_id from X-Trace-Id header
- Generates new UUID v4 if trace_id is missing or invalid
- Stores trace_id in context for request processing
- Logs trace_id extraction events (configurable)

### Convenience Functions

```python
# Client creation
create_async_client(**kwargs) -> TracedHTTPClient
create_sync_client(**kwargs) -> TracedHTTPClient

# Context managers
async with traced_httpx_client(**kwargs) as client:
    # Use client

# Middleware setup
setup_tracing_middleware(app, **kwargs)
```

## Configuration Examples

### Service-to-Service Communication

```python
from smap_shared.http import TracedHTTPClient

# Configure client for service communication
client = TracedHTTPClient(
    use_async=True,
    timeout=60.0,
    retries=5,
    base_headers={
        "Service-Name": "analysis-srv",
        "User-Agent": "SMAP-Analysis/1.0",
        "Accept": "application/json"
    }
)

# All requests will include base headers + trace_id
async with client:
    response_body, status_code = await client.post(
        "http://project-srv:8080/api/projects",
        json={"name": "New Project"},
        headers={"Authorization": "Bearer token"}
    )
```

### Error Handling

```python
from smap_shared.http import TracedHTTPClient
from smap_shared.tracing import set_trace_id, generate_trace_id

try:
    async with TracedHTTPClient() as client:
        response_body, status_code = await client.get(
            "http://unreliable-service/api/data"
        )
        
        if status_code >= 400:
            print(f"HTTP error: {status_code}")
            
except Exception as e:
    print(f"Request failed: {e}")
    # Client automatically retries based on configuration
```

## Integration with Existing Services

### Analysis Service Integration

```python
# analysis-srv/internal/usecase/analysis.py
from smap_shared.http import create_async_client
from smap_shared.tracing import get_trace_id

class AnalysisUseCase:
    async def process_data(self, project_id: str):
        # Get project details from project-srv
        async with create_async_client(
            base_headers={"Service-Name": "analysis-srv"}
        ) as client:
            project_data, status = await client.get(
                f"http://project-srv:8080/api/projects/{project_id}"
            )
            
            if status != 200:
                raise Exception(f"Failed to get project: {status}")
            
            # Process analysis...
            return analysis_result
```

### Scapper Service Integration

```python
# scapper-srv/app/main.py
from fastapi import FastAPI
from smap_shared.http import setup_tracing_middleware

app = FastAPI()

# Add tracing middleware
setup_tracing_middleware(app, log_trace_extraction=True)

@app.post("/tasks/{platform}")
async def submit_task(platform: str, request: SubmitTaskRequest):
    # trace_id automatically available from middleware
    
    # Make calls to other services with trace propagation
    async with create_async_client() as client:
        response_body, status = await client.post(
            "http://identity-srv:8080/api/validate",
            headers={"Authorization": request.auth_token}
        )
    
    return {"task_id": "generated_id"}
```

## Dependencies

### Required
- `smap_shared.tracing`: Core tracing functionality

### Optional
- `httpx`: For async HTTP client support
- `requests`: For sync HTTP client support  
- `fastapi`: For FastAPI middleware support

Install dependencies as needed:

```bash
# For async support
pip install httpx

# For sync support  
pip install requests

# For FastAPI middleware
pip install fastapi
```

## Error Handling

The HTTP client handles various error scenarios gracefully:

1. **Missing Dependencies**: Clear error messages when httpx/requests not installed
2. **Invalid Trace IDs**: Automatic validation and fallback to new trace_id generation
3. **Network Failures**: Configurable retry logic with exponential backoff
4. **Context Issues**: Graceful handling when trace_id is not in context

## Performance Considerations

- **Minimal Overhead**: Trace injection adds <1ms per request
- **Connection Pooling**: Uses underlying client connection pooling
- **Retry Logic**: Configurable retry strategy to handle transient failures
- **Memory Efficient**: Context variables provide efficient trace_id storage

## Migration from Service-Specific HTTP Clients

To migrate from existing HTTP client implementations:

1. **Replace imports**:
   ```python
   # Old
   from pkg.http_client import HTTPClient
   
   # New
   from smap_shared.http import TracedHTTPClient
   ```

2. **Update client creation**:
   ```python
   # Old
   client = HTTPClient(timeout=30)
   
   # New
   client = TracedHTTPClient(use_async=True, timeout=30)
   ```

3. **Add middleware** (for FastAPI services):
   ```python
   from smap_shared.http import setup_tracing_middleware
   setup_tracing_middleware(app)
   ```

4. **Update method calls** (if needed):
   ```python
   # Async methods return (body, status_code) tuple
   response_body, status_code = await client.get(url)
   
   # Sync methods use _sync suffix
   response_body, status_code = client.get_sync(url)
   ```

The new client maintains backward compatibility with most existing patterns while adding automatic trace propagation.