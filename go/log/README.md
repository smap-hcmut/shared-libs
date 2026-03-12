# Go Logging Package

Enhanced Zap-based logging package with automatic trace_id integration for distributed tracing.

## Features

- **Trace Integration**: Automatic trace_id injection from context
- **Structured Logging**: JSON and console output formats
- **Backward Compatibility**: Compatible with existing service logging interfaces
- **Performance Optimized**: Minimal overhead for trace_id management
- **Field Ordering**: Consistent field order in JSON logs (trace_id, level, caller, message, service)

## Quick Start

```go
package main

import (
    "context"
    "github.com/smap/shared-libs/go/log"
    "github.com/smap/shared-libs/go/tracing"
)

func main() {
    // Create logger
    logger := log.NewDevelopmentLogger()
    
    // Create context with trace_id
    tracer := tracing.NewTraceContext()
    traceID := tracer.GenerateTraceID()
    ctx := tracer.WithTraceID(context.Background(), traceID)
    
    // Log with automatic trace_id injection
    logger.Info(ctx, "Hello, world!")
    logger.Errorf(ctx, "Error occurred: %s", "something went wrong")
}
```

## Configuration

### Predefined Configurations

```go
// Development (console, colored)
logger := log.NewDevelopmentLogger()

// Production (JSON, structured)
logger := log.NewProductionLogger()

// From environment variables
logger := log.NewLoggerFromEnv()
```

### Custom Configuration

```go
config := log.ZapConfig{
    Level:        log.LevelInfo,
    Mode:         log.ModeProduction,
    Encoding:     log.EncodingJSON,
    ColorEnabled: false,
}
logger := log.NewZapLogger(config)
```

### Environment Variables

- `LOG_LEVEL`: debug, info, warn, error, fatal, panic, dpanic
- `LOG_MODE`: development, production
- `LOG_ENCODING`: console, json
- `LOG_COLOR`: true, false

## Trace Integration

The logger automatically extracts trace_id from context and includes it in all log entries:

```go
// Automatic trace_id injection
logger.Info(ctx, "This log will include trace_id from context")

// Create a traced logger instance
tracedLogger := logger.WithTrace(ctx)
tracedLogger.Info(context.Background(), "This log will include the trace_id")
```

## Log Output Formats

### Console Format (Development)
```
Thu, 12 Mar 2026 13:47:02 +0700 INFO    main.go:15      Hello, world!   {"service": "my-service", "trace_id": "550e8400-e29b-41d4-a716-446655440000"}
```

### JSON Format (Production)
```json
{
  "timestamp": "Thu, 12 Mar 2026 13:47:02 +0700",
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "level": "info",
  "caller": "main.go:15",
  "message": "Hello, world!",
  "service": "my-service"
}
```

## Migration from Service-Specific Loggers

### Before (service-specific)
```go
import "your-service/pkg/log"

logger := log.Init(log.ZapConfig{...})
logger.Info(ctx, "message")
```

### After (shared library)
```go
import "github.com/smap/shared-libs/go/log"

logger := log.NewZapLogger(log.ZapConfig{...})
logger.Info(ctx, "message")
```

## Backward Compatibility

The package maintains full backward compatibility with existing service logging interfaces:

- All existing method signatures are preserved
- `Init()` function is still available (deprecated, use `NewZapLogger()`)
- Context-aware logging methods work exactly the same
- Configuration structure is unchanged

## Performance

- Trace_id extraction: < 0.1ms per operation
- Memory overhead: < 1KB per logger instance
- No performance impact when trace_id is not present in context