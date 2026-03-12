package tracing

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidTraceID indicates an invalid trace_id format
	ErrInvalidTraceID = errors.New("invalid trace_id format")

	// ErrEmptyTraceID indicates an empty trace_id
	ErrEmptyTraceID = errors.New("trace_id is empty")

	// ErrContextPropagationFailed indicates context propagation failure
	ErrContextPropagationFailed = errors.New("failed to propagate trace_id in context")

	// ErrHeaderInjectionFailed indicates header injection failure
	ErrHeaderInjectionFailed = errors.New("failed to inject trace_id into headers")

	// ErrHeaderExtractionFailed indicates header extraction failure
	ErrHeaderExtractionFailed = errors.New("failed to extract trace_id from headers")
)

// TraceError wraps tracing-related errors with additional context
type TraceError struct {
	Op      string // Operation that failed
	TraceID string // Related trace_id if available
	Err     error  // Underlying error
}

func (e *TraceError) Error() string {
	if e.TraceID != "" {
		return fmt.Sprintf("tracing operation '%s' failed for trace_id '%s': %v", e.Op, e.TraceID, e.Err)
	}
	return fmt.Sprintf("tracing operation '%s' failed: %v", e.Op, e.Err)
}

func (e *TraceError) Unwrap() error {
	return e.Err
}

// NewTraceError creates a new TraceError
func NewTraceError(op, traceID string, err error) *TraceError {
	return &TraceError{
		Op:      op,
		TraceID: traceID,
		Err:     err,
	}
}

// ValidateAndGenerateTraceID validates a trace_id and generates a new one if invalid
// Returns the valid trace_id and any validation error for logging
func ValidateAndGenerateTraceID(traceID string, tracer TraceContext) (string, error) {
	if traceID == "" {
		return tracer.GenerateTraceID(), ErrEmptyTraceID
	}

	if !tracer.ValidateTraceID(traceID) {
		return tracer.GenerateTraceID(), NewTraceError("validation", traceID, ErrInvalidTraceID)
	}

	return traceID, nil
}
