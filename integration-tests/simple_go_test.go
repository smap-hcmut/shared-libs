package integration_tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smap/shared-libs/go/tracing"
)

// TestBasicGoTracing tests basic Go tracing functionality
func TestBasicGoTracing(t *testing.T) {
	tracer := tracing.NewTraceContext()

	// Test trace ID generation
	traceID := tracer.GenerateTraceID()
	require.NotEmpty(t, traceID, "Trace ID should not be empty")
	require.True(t, tracer.ValidateTraceID(traceID), "Generated trace ID should be valid")

	// Test context integration
	ctx := tracer.WithTraceID(context.Background(), traceID)
	retrievedID := tracer.GetTraceID(ctx)
	assert.Equal(t, traceID, retrievedID, "Retrieved trace ID should match original")

	t.Logf("✓ Basic Go tracing: %s", traceID)
}

// TestGoHTTPPropagation tests HTTP trace propagation
func TestGoHTTPPropagation(t *testing.T) {
	tracer := tracing.NewTraceContext()
	httpPropagator := tracing.NewHTTPPropagator(tracer)

	// Generate trace ID
	traceID := tracer.GenerateTraceID()
	ctx := tracer.WithTraceID(context.Background(), traceID)

	// Create HTTP request for injection
	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	// Test injection
	httpPropagator.InjectHTTP(ctx, req)
	assert.Equal(t, traceID, req.Header.Get("X-Trace-Id"), "Injected trace ID should match")

	// Test extraction
	extractedID := httpPropagator.ExtractHTTP(req)
	assert.Equal(t, traceID, extractedID, "Extracted trace ID should match")

	t.Logf("✓ Go HTTP propagation: %s", traceID)
}

// TestGoKafkaPropagation tests Kafka trace propagation
func TestGoKafkaPropagation(t *testing.T) {
	tracer := tracing.NewTraceContext()
	kafkaPropagator := tracing.NewKafkaPropagator(tracer)

	// Generate trace ID
	traceID := tracer.GenerateTraceID()
	ctx := tracer.WithTraceID(context.Background(), traceID)

	// Test injection
	headers := make(map[string]string)
	kafkaPropagator.InjectKafka(ctx, headers)

	assert.Contains(t, headers, "X-Trace-Id", "X-Trace-Id header should be injected")
	assert.Equal(t, traceID, headers["X-Trace-Id"], "Injected trace ID should match")

	// Test extraction
	extractedID := kafkaPropagator.ExtractKafka(headers)
	assert.Equal(t, traceID, extractedID, "Extracted trace ID should match")

	t.Logf("✓ Go Kafka propagation: %s", traceID)
}

// TestGoEndToEndFlow tests end-to-end trace flow
func TestGoEndToEndFlow(t *testing.T) {
	tracer := tracing.NewTraceContext()
	httpPropagator := tracing.NewHTTPPropagator(tracer)
	kafkaPropagator := tracing.NewKafkaPropagator(tracer)

	// Original trace ID
	originalTraceID := tracer.GenerateTraceID()
	ctx := tracer.WithTraceID(context.Background(), originalTraceID)

	// HTTP propagation
	req, err := http.NewRequest("GET", "http://example.com", nil)
	require.NoError(t, err)

	httpPropagator.InjectHTTP(ctx, req)
	httpTraceID := httpPropagator.ExtractHTTP(req)

	// Kafka propagation
	kafkaHeaders := make(map[string]string)
	kafkaPropagator.InjectKafka(ctx, kafkaHeaders)
	kafkaTraceID := kafkaPropagator.ExtractKafka(kafkaHeaders)

	// Verify consistency
	assert.Equal(t, originalTraceID, httpTraceID, "HTTP trace ID should match original")
	assert.Equal(t, originalTraceID, kafkaTraceID, "Kafka trace ID should match original")

	t.Logf("✓ Go end-to-end flow: %s", originalTraceID)
}
