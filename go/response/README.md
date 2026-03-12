# Response Package

The response package provides standardized HTTP response handling with distributed tracing integration for SMAP services.

## Features

- **Trace Integration**: Automatic trace_id injection in all responses
- **Error Reporting**: Optional external error reporting (Discord, etc.)
- **Backward Compatibility**: Drop-in replacement for existing response packages
- **Structured Responses**: Consistent response format across all services
- **Stack Trace Capture**: Detailed error reporting with stack traces

## Usage

### Basic Usage (Backward Compatible)

```go
import "github.com/smap-hcmut/shared-libs/go/response"

// Simple responses (no trace integration)
response.OK(c, data)
response.Unauthorized(c)
response.Forbidden(c)
response.Error(c, err)
```

### Advanced Usage with Trace Integration

```go
import (
    "github.com/smap-hcmut/shared-libs/go/response"
    "github.com/smap-hcmut/shared-libs/go/tracing"
)

// Create response manager with trace integration
tracer := tracing.NewTraceContext()
manager := response.NewResponseManager(tracer, nil)

// All responses will include trace_id
manager.OK(c, data)
manager.Unauthorized(c)
manager.Error(c, err)
```

### With Error Reporting

```go
// Implement ErrorReporter interface
type DiscordReporter struct {
    // Discord webhook implementation
}

func (d *DiscordReporter) ReportBug(ctx context.Context, message string) error {
    // Send error report to Discord
    return nil
}

// Create manager with error reporting
reporter := &DiscordReporter{}
manager := response.NewResponseManager(tracer, reporter)

// Internal server errors will be automatically reported
manager.Error(c, err)
```

## Response Format

All responses follow this structure:

```json
{
    "error_code": 0,
    "message": "Success",
    "data": {...},
    "trace_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Error Codes

- `0`: Success
- `400`: Validation Error
- `401`: Unauthorized
- `403`: Permission Denied
- `500`: Internal Server Error

## Migration Guide

### From Local Response Package

1. Update imports:
```go
// Before
import "your-service/pkg/response"

// After
import "github.com/smap-hcmut/shared-libs/go/response"
```

2. No code changes needed for basic usage
3. Optional: Add trace integration for enhanced debugging

### Trace Integration Benefits

- **Request Tracking**: Follow requests across service boundaries
- **Error Correlation**: Link errors to specific request flows
- **Performance Monitoring**: Measure request latency across services
- **Debugging**: Easier troubleshooting with trace context