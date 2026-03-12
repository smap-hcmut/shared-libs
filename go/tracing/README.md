# Go Tracing Library

This package provides comprehensive distributed tracing capabilities for SMAP microservices, enabling end-to-end request tracking across HTTP calls, Kafka messages, and database operations.

## Features

- **UUID v4 Trace ID Management**: Generate and validate UUID v4 trace identifiers
- **HTTP Propagation**: Automatic trace_id injection/extraction for HTTP requests
- **Kafka Propagation**: Trace_id management for Kafka message headers
- **Context Integration**: Seamless integration with Go's context.Context
- **Gin Middleware**: Ready-to-use middleware for Gin web framework
- **Error Handling**: Graceful handling of invalid or missing trace_ids
- **Backward Compatibility**: Compatible with existing service tracing patterns

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/smap/shared-libs/go/tracing"
)

func main() {
    // Create tracing components
    components := tracing.NewTracingComponents()
    
    // Set up Gin router with tracing middleware
    r := gin.Default()
    r.Use(tracing.GinTraceMiddleware(
        components.TraceContext, 
        components.HTTPPropagator,
    ))
    
    r.GET("/api/test", func(c *gin.Context) {
        // Get trace_id from context
        traceID := tracing.GetTraceIDFromGinContext(c)
        c.JSON(200, gin.H{"trace_id": traceID})
    })
    
    r.Run(":8080")
}
```

## Core Interfaces

### TraceContext
Manages trace_id throughout request lifecycle:
- `GetTraceID(ctx)` - Extract trace_id from context
- `WithTraceID(ctx, traceID)` - Add trace_id to context
- `GenerateTraceID()` - Create new UUID v4 trace_id
- `ValidateTraceID(traceID)` - Validate UUID v4 format

### HTTPPropagator
Handles HTTP trace_id propagation:
- `InjectHTTP(ctx, req)` - Add trace_id to outbound request headers
- `ExtractHTTP(req)` - Extract trace_id from inbound request headers

### KafkaPropagator
Handles Kafka trace_id propagation:
- `InjectKafka(ctx, headers)` - Add trace_id to message headers
- `ExtractKafka(headers)` - Extract trace_id from message headers

## Components

- `interfaces.go` - Core interface definitions
- `context.go` - TraceContext implementation with UUID v4 support
- `http.go` - HTTPPropagator for X-Trace-Id header management  
- `kafka.go` - KafkaPropagator for message header management
- `validation.go` - UUID v4 validation utilities
- `middleware.go` - Ready-to-use Gin middleware
- `factory.go` - Component factory functions
- `errors.go` - Error handling and validation utilities

## Requirements Satisfied

This implementation satisfies the following requirements:
- **1.1, 1.2**: HTTP trace_id extraction and generation
- **1.4**: Trace_id storage in request context
- **4.1, 4.4**: UUID v4 format consistency
- **5.1, 5.2**: Context management and accessibility
- **6.1, 6.2**: Trace_id validation and error handling