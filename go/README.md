# SMAP Go Shared Libraries

Go implementation of shared libraries for SMAP services with distributed tracing support.

## Packages

### Core Tracing
- `tracing/` - Core trace_id management and propagation interfaces
- `tracing/context.go` - TraceContext implementation with UUID v4 support
- `tracing/http.go` - HTTPPropagator for X-Trace-Id header management
- `tracing/kafka.go` - KafkaPropagator for message header management
- `tracing/validation.go` - UUID v4 validation utilities
- `tracing/middleware.go` - Ready-to-use HTTP middleware

### Enhanced Shared Packages
- `log/` - Enhanced logging with automatic trace_id injection
- `http/` - HTTP client with automatic trace propagation
- `kafka/` - Kafka producer/consumer with trace headers
- `redis/` - Redis client with trace context
- `postgres/` - PostgreSQL client with trace logging
- `auth/` - JWT utilities with trace context
- `response/` - HTTP response utilities

## Usage

```go
package main

import (
    "context"
    "github.com/smap/shared-libs/go/tracing"
    "github.com/smap/shared-libs/go/log"
    "github.com/smap/shared-libs/go/http"
)

func main() {
    // Initialize tracing
    tracer := tracing.NewTraceContext()
    logger := log.NewLogger()
    httpClient := http.NewTracedClient()
    
    // Use in request context
    ctx := context.Background()
    ctx = tracer.WithTraceID(ctx, tracer.GenerateTraceID())
    
    // Automatic trace propagation
    response, err := httpClient.Get(ctx, "http://api.example.com/data", nil)
    if err != nil {
        logger.Error(ctx, "HTTP request failed", "error", err)
    }
}
```

## Migration from Service Packages

Replace service-specific imports:
```go
// Before
import "identity-srv/pkg/log"
import "project-srv/pkg/http"

// After  
import "github.com/smap/shared-libs/go/log"
import "github.com/smap/shared-libs/go/http"
```

## Testing

```bash
go test ./...
go test -race ./...
go test -bench=. ./...
```