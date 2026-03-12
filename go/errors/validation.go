package errors

import (
	"context"
	"fmt"
	"strings"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// NewValidationError creates a new validation error with trace context.
func NewValidationError(code int, field string, messages ...string) *ValidationError {
	return &ValidationError{
		Code:     code,
		Field:    field,
		Messages: messages,
	}
}

// NewValidationErrorWithTrace creates a new validation error with trace context.
func NewValidationErrorWithTrace(ctx context.Context, code int, field string, messages ...string) *ValidationError {
	tracer := tracing.NewTraceContext()
	return &ValidationError{
		Code:     code,
		Field:    field,
		Messages: messages,
		TraceID:  tracer.GetTraceID(ctx),
	}
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Field, strings.Join(e.Messages, ", "))
	if e.TraceID != "" {
		msg += fmt.Sprintf(" (trace_id=%s)", e.TraceID)
	}
	return msg
}

// WithTraceID adds trace_id to the validation error.
func (e *ValidationError) WithTraceID(traceID string) *ValidationError {
	e.TraceID = traceID
	return e
}

// NewValidationErrorCollector creates a new validation error collector with trace context.
func NewValidationErrorCollector() *ValidationErrorCollector {
	return &ValidationErrorCollector{
		errors: make([]*ValidationError, 0),
	}
}

// NewValidationErrorCollectorWithTrace creates a new validation error collector with trace context.
func NewValidationErrorCollectorWithTrace(ctx context.Context) *ValidationErrorCollector {
	tracer := tracing.NewTraceContext()
	return &ValidationErrorCollector{
		errors:  make([]*ValidationError, 0),
		traceID: tracer.GetTraceID(ctx),
	}
}

// Add adds a validation error to the collector.
func (c *ValidationErrorCollector) Add(err *ValidationError) *ValidationErrorCollector {
	// Add trace_id to error if collector has one and error doesn't
	if c.traceID != "" && err.TraceID == "" {
		err.TraceID = c.traceID
	}
	c.errors = append(c.errors, err)
	return c
}

// AddField adds a validation error for a specific field.
func (c *ValidationErrorCollector) AddField(code int, field string, messages ...string) *ValidationErrorCollector {
	err := &ValidationError{
		Code:     code,
		Field:    field,
		Messages: messages,
		TraceID:  c.traceID,
	}
	c.errors = append(c.errors, err)
	return c
}

// HasError returns true if the collector has any errors.
func (c *ValidationErrorCollector) HasError() bool {
	return len(c.errors) > 0
}

// Errors returns all validation errors.
func (c *ValidationErrorCollector) Errors() []*ValidationError {
	return c.errors
}

// Error implements the error interface for ValidationErrorCollector.
func (c *ValidationErrorCollector) Error() string {
	var msgs []string
	for _, err := range c.errors {
		msgs = append(msgs, err.Error())
	}
	result := strings.Join(msgs, ", ")
	if c.traceID != "" {
		result += fmt.Sprintf(" (trace_id=%s)", c.traceID)
	}
	return result
}

// WithTraceID adds trace_id to the collector and all its errors.
func (c *ValidationErrorCollector) WithTraceID(traceID string) *ValidationErrorCollector {
	c.traceID = traceID
	for _, err := range c.errors {
		if err.TraceID == "" {
			err.TraceID = traceID
		}
	}
	return c
}
