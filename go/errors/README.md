# Errors Package

Unified error handling with distributed tracing support for SMAP services.

## Features

- **Trace Integration**: Automatic trace_id injection in all error types
- **Multiple Error Types**: Validation, Permission, HTTP, Business, System errors
- **Error Collectors**: Collect multiple validation/permission errors
- **Predefined Errors**: Common HTTP status codes and business logic errors
- **Error Unwrapping**: Support for Go 1.13+ error unwrapping
- **Structured Logging**: JSON-serializable error structures

## Error Types

### ValidationError
For input validation errors with field-level details:
```go
err := errors.NewValidationErrorWithTrace(ctx, 400, "email", "Invalid email format")
collector := errors.NewValidationErrorCollectorWithTrace(ctx)
collector.AddField(400, "name", "Name is required")
```

### HTTPError  
For HTTP status code errors:
```go
err := errors.NewUnauthorizedErrorWithTrace(ctx)
err := errors.NewNotFoundErrorWithTrace(ctx, "user")
```

### BusinessError
For business logic errors:
```go
err := errors.NewPermissionDeniedErrorWithTrace(ctx, "admin_panel")
err := errors.NewResourceConflictErrorWithTrace(ctx, "username")
```

### SystemError
For system-level errors with component context:
```go
err := errors.NewDatabaseErrorWithTrace(ctx, "insert", "Connection failed")
err := errors.NewExternalServiceErrorWithTrace(ctx, "api_call", "Timeout")
```

## Trace Integration

All error types automatically include trace_id when created with trace context:
- Errors include trace_id in JSON serialization
- Error messages include trace_id for logging
- Collectors propagate trace_id to all contained errors