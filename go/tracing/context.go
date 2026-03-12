package tracing

import (
	"context"

	"github.com/google/uuid"
)

// traceContextKey is a private type for context keys to avoid collisions
type traceContextKey struct{}

// traceContextImpl implements the TraceContext interface
type traceContextImpl struct{}

// NewTraceContext creates a new TraceContext implementation
func NewTraceContext() TraceContext {
	return &traceContextImpl{}
}

// GetTraceID returns current trace_id from context
func (t *traceContextImpl) GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceContextKey{}).(string); ok {
		return traceID
	}
	return ""
}

// WithTraceID adds trace_id to context
func (t *traceContextImpl) WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceContextKey{}, traceID)
}

// GenerateTraceID creates new UUID v4 trace_id
func (t *traceContextImpl) GenerateTraceID() string {
	return uuid.New().String()
}

// ValidateTraceID checks if trace_id is valid UUID v4
func (t *traceContextImpl) ValidateTraceID(traceID string) bool {
	return ValidateUUIDv4(traceID)
}
