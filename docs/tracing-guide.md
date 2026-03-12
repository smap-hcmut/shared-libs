# Tracing Guide: Distributed Trace Management

This guide explains how to use the SMAP shared library's distributed tracing capabilities for end-to-end request tracking.

## Overview

The tracing system provides UUID v4 trace_id propagation across all service boundaries including HTTP calls, Kafka messages, and database operations. Every request can be tracked from entry point through the entire microservices architecture.

## Core Concepts

### Trace ID
- **Format**: UUID v4 (e.g., `550e8400-e29b-41d4-a716-446655440000`)
- **Header**: `X-Trace-Id`
- **Scope**: Single request across all services
- **Lifecycle**: Generated at entry point, propagated throughout request flow

### Trace Context
- **Go**: Stored in `context.Context`
- **Python**: Stored in `contextvars`
- **Propagation**: Automatic across HTTP, Kafka, database operations
- **Inheritance**: Passed to goroutines/async tasks

## Usage Patterns

### Go Services

#### Basic Setup
```go
package main

import (
    "context"
    "github.com/smap/shared-libs/go/tracing"
    "github.com/smap/shared-libs/go/log"
    "github.com/gin-gonic/gin"
)

func main() {
    // Initialize tracing components
    tracer := tracing.NewTraceContext()
    propagator := tracing.NewHTTPPropagator()
    logger := log.NewLogger()
    
    // Setup HTTP server with trace middleware
    r := gin.Default()
    r.Use(tracing.TraceMiddleware(tracer, propagator))
    
    r.GET("/api/data", func(c *gin.Context) {
        ctx := c.Request.Context()
        
        // Trace ID is automatically available
        traceID := tracer.GetTraceID(ctx)
        logger.Info(ctx, "Processing request", "trace_id", traceID)
        
        c.JSON(200, gin.H{"status": "success"})
    })
    
    r.Run(":8080")
}
```

#### HTTP Client Usage
```go
import "github.com/smap/shared-libs/go/http"

func callExternalService(ctx context.Context) error {
    client := http.NewTracedClient()
    
    // Trace ID automatically injected in X-Trace-Id header
    resp, err := client.Get(ctx, "http://project-srv/api/projects", nil)
    if err != nil {
        return err
    }
    
    // Process response
    return nil
}
```

#### Kafka Producer Usage
```go
import "github.com/smap/shared-libs/go/kafka"

func publishEvent(ctx context.Context, event interface{}) error {
    producer := kafka.NewTracedProducer("localhost:9092")
    
    // Trace ID automatically injected in message headers
    return producer.Publish(ctx, "events", "key", event)
}
```

#### Database Operations
```go
import "github.com/smap/shared-libs/go/postgres"

func getUserByID(ctx context.Context, userID string) (*User, error) {
    db := postgres.NewTracedClient(connectionString)
    
    // Query logs automatically include trace_id
    var user User
    err := db.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", userID).Scan(&user)
    return &user, err
}
```

### Python Services

#### Basic Setup
```python
from fastapi import FastAPI
from smap_shared.tracing import trace_middleware, TraceContext
from smap_shared.logger import Logger

app = FastAPI()
app.middleware("http")(trace_middleware)

trace_context = TraceContext()
logger = Logger()

@app.get("/api/data")
async def get_data():
    # Trace ID is automatically available
    trace_id = trace_context.get_trace_id()
    logger.info("Processing request", extra={"trace_id": trace_id})
    
    return {"status": "success"}
```

#### HTTP Client Usage
```python
from smap_shared.http import TracedHTTPClient

async def call_external_service():
    client = TracedHTTPClient()
    
    # Trace ID automatically injected in X-Trace-Id header
    response = await client.get("http://analysis-srv/api/analyze")
    return response.json()
```

#### Kafka Consumer Usage
```python
from smap_shared.kafka import TracedKafkaConsumer

def process_messages():
    consumer = TracedKafkaConsumer(['events'], bootstrap_servers=['localhost:9092'])
    
    for message in consumer:
        # Trace ID automatically extracted from message headers
        trace_id = trace_context.get_trace_id()
        logger.info("Processing message", extra={"trace_id": trace_id})
        
        # Process message
        process_event(message.value)
```

## Advanced Patterns

### Context Propagation in Concurrent Operations

#### Go Goroutines
```go
func processAsync(ctx context.Context, data []Item) {
    var wg sync.WaitGroup
    
    for _, item := range data {
        wg.Add(1)
        go func(ctx context.Context, item Item) {
            defer wg.Done()
            
            // Context with trace_id is inherited
            logger.Info(ctx, "Processing item", "item_id", item.ID)
            processItem(ctx, item)
        }(ctx, item) // Pass context to goroutine
    }
    
    wg.Wait()
}
```

#### Python Async Tasks
```python
import asyncio
from smap_shared.tracing import TraceContext

async def process_async(items):
    trace_context = TraceContext()
    
    async def process_item(item):
        # Trace context is inherited in async tasks
        trace_id = trace_context.get_trace_id()
        logger.info(f"Processing item {item.id}", extra={"trace_id": trace_id})
        await process_item_logic(item)
    
    # Create tasks with inherited context
    tasks = [process_item(item) for item in items]
    await asyncio.gather(*tasks)
```

### Custom Trace Propagation

#### Manual Header Injection
```go
func customHTTPCall(ctx context.Context, url string) error {
    tracer := tracing.NewTraceContext()
    propagator := tracing.NewHTTPPropagator()
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }
    
    // Manual trace injection
    propagator.InjectHTTP(ctx, req)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    // Handle response
    return err
}
```

#### Manual Header Extraction
```python
from smap_shared.tracing import HTTPPropagator, TraceContext

def handle_webhook(request):
    propagator = HTTPPropagator()
    trace_context = TraceContext()
    
    # Manual trace extraction
    trace_id = propagator.extract_http(dict(request.headers))
    if trace_id:
        trace_context.set_trace_id(trace_id)
    
    # Process webhook with trace context
    process_webhook_data(request.json())
```

## Error Handling and Validation

### Invalid Trace ID Handling
```go
func validateAndSetTraceID(ctx context.Context, traceID string) context.Context {
    tracer := tracing.NewTraceContext()
    
    if !tracer.ValidateTraceID(traceID) {
        // Generate new trace_id for invalid input
        traceID = tracer.GenerateTraceID()
        logger.Warn(ctx, "Invalid trace_id received, generated new one", 
                   "new_trace_id", traceID)
    }
    
    return tracer.WithTraceID(ctx, traceID)
}
```

### Graceful Degradation
```python
def safe_trace_operation():
    try:
        trace_context = TraceContext()
        trace_id = trace_context.get_trace_id()
        
        if not trace_id:
            # Generate new trace_id if missing
            trace_id = trace_context.generate_trace_id()
            trace_context.set_trace_id(trace_id)
            
        return trace_id
    except Exception as e:
        logger.error(f"Trace operation failed: {e}")
        # Continue without trace_id
        return None
```

## Monitoring and Debugging

### Trace Flow Validation
```bash
# Check trace propagation in logs
grep "trace_id=550e8400-e29b-41d4-a716-446655440000" /var/log/smap/*.log

# Monitor trace success rates
curl -s http://identity-srv/metrics | grep trace_propagation_success_rate

# Check Kafka message headers
kafka-console-consumer --topic events --bootstrap-server localhost:9092 \
  --property print.headers=true
```

### Performance Monitoring
```go
// Add trace performance metrics
func measureTraceImpact(ctx context.Context, operation func() error) error {
    start := time.Now()
    err := operation()
    duration := time.Since(start)
    
    // Log if trace operations exceed 1ms threshold
    if duration > time.Millisecond {
        logger.Warn(ctx, "Trace operation exceeded threshold", 
                   "duration", duration)
    }
    
    return err
}
```

## Best Practices

### Do's
- ✅ Always pass context through function calls
- ✅ Use trace-aware logger methods
- ✅ Validate trace_id format before processing
- ✅ Handle missing trace_id gracefully
- ✅ Monitor trace propagation success rates
- ✅ Use structured logging with trace_id field

### Don'ts
- ❌ Don't manually manage trace_id strings
- ❌ Don't skip context propagation in goroutines/async tasks
- ❌ Don't ignore trace validation failures
- ❌ Don't block request processing on trace operations
- ❌ Don't modify trace_id during request processing
- ❌ Don't use trace_id for business logic decisions

### Performance Guidelines
- Keep trace operations under 1ms per request
- Use efficient UUID v4 generation
- Minimize context copying overhead
- Cache trace propagators when possible
- Monitor memory usage in high-concurrency scenarios

## Troubleshooting

### Common Issues

**Trace ID Not Propagating**
- Check middleware configuration order
- Verify context is passed to all function calls
- Ensure HTTP client uses traced version

**Performance Degradation**
- Monitor trace operation latency
- Check context overhead in tight loops
- Verify efficient header injection

**Invalid Trace ID Errors**
- Check UUID v4 format validation
- Verify header extraction logic
- Monitor validation failure rates

**Context Loss in Concurrent Operations**
- Ensure context is passed to goroutines
- Check async task context inheritance
- Verify proper context cancellation

### Debug Logging
```go
// Enable debug logging for trace operations
logger.Debug(ctx, "Trace operation", 
            "operation", "http_inject",
            "trace_id", tracer.GetTraceID(ctx),
            "headers", req.Header)
```

## Integration Examples

See [examples/](examples/) directory for complete integration examples:
- `go-service-example/` - Complete Go service with tracing
- `python-service-example/` - Complete Python service with tracing
- `cross-service-flow/` - End-to-end trace flow example
- `performance-benchmarks/` - Performance testing examples