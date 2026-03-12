package errors

import (
	"context"
	"fmt"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// NewBusinessError creates a new business logic error.
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// NewBusinessErrorWithTrace creates a new business logic error with trace context.
func NewBusinessErrorWithTrace(ctx context.Context, code, message string) *BusinessError {
	tracer := tracing.NewTraceContext()
	return &BusinessError{
		Code:    code,
		Message: message,
		TraceID: tracer.GetTraceID(ctx),
	}
}

// NewBusinessErrorWithDetails creates a new business logic error with details.
func NewBusinessErrorWithDetails(code, message, details string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// NewBusinessErrorWithDetailsAndTrace creates a new business logic error with details and trace context.
func NewBusinessErrorWithDetailsAndTrace(ctx context.Context, code, message, details string) *BusinessError {
	tracer := tracing.NewTraceContext()
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: details,
		TraceID: tracer.GetTraceID(ctx),
	}
}

// Error implements the error interface for BusinessError.
func (e *BusinessError) Error() string {
	msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
	if e.Details != "" {
		msg += fmt.Sprintf(": %s", e.Details)
	}
	if e.TraceID != "" {
		msg += fmt.Sprintf(" (trace_id=%s)", e.TraceID)
	}
	return msg
}

// WithTraceID adds trace_id to the business error.
func (e *BusinessError) WithTraceID(traceID string) *BusinessError {
	e.TraceID = traceID
	return e
}

// WithDetails adds details to the business error.
func (e *BusinessError) WithDetails(details string) *BusinessError {
	e.Details = details
	return e
}

// Predefined business errors

// NewValidationFailedError creates a validation failed business error.
func NewValidationFailedError(message string) *BusinessError {
	return NewBusinessError(CodeValidationFailed, message)
}

// NewValidationFailedErrorWithTrace creates a validation failed business error with trace context.
func NewValidationFailedErrorWithTrace(ctx context.Context, message string) *BusinessError {
	return NewBusinessErrorWithTrace(ctx, CodeValidationFailed, message)
}

// NewPermissionDeniedError creates a permission denied business error.
func NewPermissionDeniedError(resource string) *BusinessError {
	message := "Permission denied"
	if resource != "" {
		message = fmt.Sprintf("Permission denied for resource: %s", resource)
	}
	return NewBusinessError(CodePermissionDenied, message)
}

// NewPermissionDeniedErrorWithTrace creates a permission denied business error with trace context.
func NewPermissionDeniedErrorWithTrace(ctx context.Context, resource string) *BusinessError {
	message := "Permission denied"
	if resource != "" {
		message = fmt.Sprintf("Permission denied for resource: %s", resource)
	}
	return NewBusinessErrorWithTrace(ctx, CodePermissionDenied, message)
}

// NewResourceNotFoundError creates a resource not found business error.
func NewResourceNotFoundError(resource string) *BusinessError {
	message := fmt.Sprintf("Resource not found: %s", resource)
	return NewBusinessError(CodeResourceNotFound, message)
}

// NewResourceNotFoundErrorWithTrace creates a resource not found business error with trace context.
func NewResourceNotFoundErrorWithTrace(ctx context.Context, resource string) *BusinessError {
	message := fmt.Sprintf("Resource not found: %s", resource)
	return NewBusinessErrorWithTrace(ctx, CodeResourceNotFound, message)
}

// NewResourceConflictError creates a resource conflict business error.
func NewResourceConflictError(resource string) *BusinessError {
	message := fmt.Sprintf("Resource conflict: %s", resource)
	return NewBusinessError(CodeResourceConflict, message)
}

// NewResourceConflictErrorWithTrace creates a resource conflict business error with trace context.
func NewResourceConflictErrorWithTrace(ctx context.Context, resource string) *BusinessError {
	message := fmt.Sprintf("Resource conflict: %s", resource)
	return NewBusinessErrorWithTrace(ctx, CodeResourceConflict, message)
}
