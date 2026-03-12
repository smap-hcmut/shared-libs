package errors

import (
	"context"
	"fmt"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// NewSystemError creates a new system error.
func NewSystemError(component, operation, message string) *SystemError {
	return &SystemError{
		Component: component,
		Operation: operation,
		Message:   message,
	}
}

// NewSystemErrorWithTrace creates a new system error with trace context.
func NewSystemErrorWithTrace(ctx context.Context, component, operation, message string) *SystemError {
	tracer := tracing.NewTraceContext()
	return &SystemError{
		Component: component,
		Operation: operation,
		Message:   message,
		TraceID:   tracer.GetTraceID(ctx),
	}
}

// NewSystemErrorWithCause creates a new system error with underlying cause.
func NewSystemErrorWithCause(component, operation, message string, cause error) *SystemError {
	return &SystemError{
		Component: component,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// NewSystemErrorWithCauseAndTrace creates a new system error with cause and trace context.
func NewSystemErrorWithCauseAndTrace(ctx context.Context, component, operation, message string, cause error) *SystemError {
	tracer := tracing.NewTraceContext()
	return &SystemError{
		Component: component,
		Operation: operation,
		Message:   message,
		Cause:     cause,
		TraceID:   tracer.GetTraceID(ctx),
	}
}

// Error implements the error interface for SystemError.
func (e *SystemError) Error() string {
	msg := fmt.Sprintf("[%s.%s] %s", e.Component, e.Operation, e.Message)
	if e.Cause != nil {
		msg += fmt.Sprintf(": %v", e.Cause)
	}
	if e.TraceID != "" {
		msg += fmt.Sprintf(" (trace_id=%s)", e.TraceID)
	}
	return msg
}

// WithTraceID adds trace_id to the system error.
func (e *SystemError) WithTraceID(traceID string) *SystemError {
	e.TraceID = traceID
	return e
}

// WithCause adds underlying cause to the system error.
func (e *SystemError) WithCause(cause error) *SystemError {
	e.Cause = cause
	return e
}

// Unwrap returns the underlying cause for error unwrapping.
func (e *SystemError) Unwrap() error {
	return e.Cause
}

// Predefined system errors

// NewDatabaseError creates a database system error.
func NewDatabaseError(operation, message string) *SystemError {
	return NewSystemError(ComponentDatabase, operation, message)
}

// NewDatabaseErrorWithTrace creates a database system error with trace context.
func NewDatabaseErrorWithTrace(ctx context.Context, operation, message string) *SystemError {
	return NewSystemErrorWithTrace(ctx, ComponentDatabase, operation, message)
}

// NewDatabaseErrorWithCause creates a database system error with cause.
func NewDatabaseErrorWithCause(operation, message string, cause error) *SystemError {
	return NewSystemErrorWithCause(ComponentDatabase, operation, message, cause)
}

// NewDatabaseErrorWithCauseAndTrace creates a database system error with cause and trace context.
func NewDatabaseErrorWithCauseAndTrace(ctx context.Context, operation, message string, cause error) *SystemError {
	return NewSystemErrorWithCauseAndTrace(ctx, ComponentDatabase, operation, message, cause)
}

// NewCacheError creates a cache system error.
func NewCacheError(operation, message string) *SystemError {
	return NewSystemError(ComponentCache, operation, message)
}

// NewCacheErrorWithTrace creates a cache system error with trace context.
func NewCacheErrorWithTrace(ctx context.Context, operation, message string) *SystemError {
	return NewSystemErrorWithTrace(ctx, ComponentCache, operation, message)
}

// NewQueueError creates a queue system error.
func NewQueueError(operation, message string) *SystemError {
	return NewSystemError(ComponentQueue, operation, message)
}

// NewQueueErrorWithTrace creates a queue system error with trace context.
func NewQueueErrorWithTrace(ctx context.Context, operation, message string) *SystemError {
	return NewSystemErrorWithTrace(ctx, ComponentQueue, operation, message)
}

// NewExternalServiceError creates an external service system error.
func NewExternalServiceError(operation, message string) *SystemError {
	return NewSystemError(ComponentExternal, operation, message)
}

// NewExternalServiceErrorWithTrace creates an external service system error with trace context.
func NewExternalServiceErrorWithTrace(ctx context.Context, operation, message string) *SystemError {
	return NewSystemErrorWithTrace(ctx, ComponentExternal, operation, message)
}
