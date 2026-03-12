package tracing

import (
	"context"
	"net/http"
)

// TraceContext manages trace_id throughout request lifecycle
type TraceContext interface {
	// GetTraceID returns current trace_id from context
	GetTraceID(ctx context.Context) string

	// WithTraceID adds trace_id to context
	WithTraceID(ctx context.Context, traceID string) context.Context

	// GenerateTraceID creates new UUID v4 trace_id
	GenerateTraceID() string

	// ValidateTraceID checks if trace_id is valid UUID v4
	ValidateTraceID(traceID string) bool
}

// HTTPPropagator handles HTTP trace_id propagation
type HTTPPropagator interface {
	// InjectHTTP adds trace_id to outbound HTTP request headers
	InjectHTTP(ctx context.Context, req *http.Request)

	// ExtractHTTP retrieves trace_id from inbound HTTP request headers
	ExtractHTTP(req *http.Request) string
}

// KafkaPropagator handles Kafka trace_id propagation
type KafkaPropagator interface {
	// InjectKafka adds trace_id to Kafka message headers
	InjectKafka(ctx context.Context, headers map[string]string)

	// ExtractKafka retrieves trace_id from Kafka message headers
	ExtractKafka(headers map[string]string) string
}
