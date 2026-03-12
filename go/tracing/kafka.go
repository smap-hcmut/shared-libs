package tracing

import (
	"context"
)

// kafkaPropagatorImpl implements the KafkaPropagator interface
type kafkaPropagatorImpl struct {
	tracer TraceContext
}

// NewKafkaPropagator creates a new KafkaPropagator implementation
func NewKafkaPropagator(tracer TraceContext) KafkaPropagator {
	return &kafkaPropagatorImpl{
		tracer: tracer,
	}
}

// InjectKafka adds trace_id to Kafka message headers
func (k *kafkaPropagatorImpl) InjectKafka(ctx context.Context, headers map[string]string) {
	traceID := k.tracer.GetTraceID(ctx)
	if traceID != "" {
		headers[TraceIDHeader] = traceID
	}
}

// ExtractKafka retrieves trace_id from Kafka message headers
func (k *kafkaPropagatorImpl) ExtractKafka(headers map[string]string) string {
	return headers[TraceIDHeader]
}
