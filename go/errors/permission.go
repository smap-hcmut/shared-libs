package errors

import (
	"context"
	"fmt"
	"strings"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// NewPermissionError creates a new permission error.
func NewPermissionError(code int, field string, messages ...string) *PermissionError {
	return &PermissionError{
		Code:     code,
		Field:    field,
		Messages: messages,
	}
}

// NewPermissionErrorWithTrace creates a new permission error with trace context.
func NewPermissionErrorWithTrace(ctx context.Context, code int, field string, messages ...string) *PermissionError {
	tracer := tracing.NewTraceContext()
	return &PermissionError{
		Code:     code,
		Field:    field,
		Messages: messages,
		TraceID:  tracer.GetTraceID(ctx),
	}
}

// Error implements the error interface for PermissionError.
func (e *PermissionError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Field, strings.Join(e.Messages, ", "))
	if e.TraceID != "" {
		msg += fmt.Sprintf(" (trace_id=%s)", e.TraceID)
	}
	return msg
}

// WithTraceID adds trace_id to the permission error.
func (e *PermissionError) WithTraceID(traceID string) *PermissionError {
	e.TraceID = traceID
	return e
}

// NewPermissionErrorCollector creates a new permission error collector.
func NewPermissionErrorCollector() *PermissionErrorCollector {
	return &PermissionErrorCollector{
		errors: make([]*PermissionError, 0),
	}
}

// NewPermissionErrorCollectorWithTrace creates a new permission error collector with trace context.
func NewPermissionErrorCollectorWithTrace(ctx context.Context) *PermissionErrorCollector {
	tracer := tracing.NewTraceContext()
	return &PermissionErrorCollector{
		errors:  make([]*PermissionError, 0),
		traceID: tracer.GetTraceID(ctx),
	}
}

// Add adds a permission error to the collector.
func (c *PermissionErrorCollector) Add(err *PermissionError) *PermissionErrorCollector {
	// Add trace_id to error if collector has one and error doesn't
	if c.traceID != "" && err.TraceID == "" {
		err.TraceID = c.traceID
	}
	c.errors = append(c.errors, err)
	return c
}

// AddField adds a permission error for a specific field.
func (c *PermissionErrorCollector) AddField(code int, field string, messages ...string) *PermissionErrorCollector {
	err := &PermissionError{
		Code:     code,
		Field:    field,
		Messages: messages,
		TraceID:  c.traceID,
	}
	c.errors = append(c.errors, err)
	return c
}

// HasError returns true if the collector has any errors.
func (c *PermissionErrorCollector) HasError() bool {
	return len(c.errors) > 0
}

// Errors returns all permission errors.
func (c *PermissionErrorCollector) Errors() []*PermissionError {
	return c.errors
}

// Error implements the error interface for PermissionErrorCollector.
func (c *PermissionErrorCollector) Error() string {
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
func (c *PermissionErrorCollector) WithTraceID(traceID string) *PermissionErrorCollector {
	c.traceID = traceID
	for _, err := range c.errors {
		if err.TraceID == "" {
			err.TraceID = traceID
		}
	}
	return c
}
