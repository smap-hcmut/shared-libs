package tracing

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestTraceContext(t *testing.T) {
	tracer := NewTraceContext()

	t.Run("GenerateTraceID", func(t *testing.T) {
		traceID := tracer.GenerateTraceID()
		if traceID == "" {
			t.Error("GenerateTraceID should not return empty string")
		}

		// Validate it's a proper UUID v4
		if !tracer.ValidateTraceID(traceID) {
			t.Errorf("Generated trace_id %s is not valid UUID v4", traceID)
		}
	})

	t.Run("ValidateTraceID", func(t *testing.T) {
		validUUID := uuid.New().String()
		if !tracer.ValidateTraceID(validUUID) {
			t.Errorf("Valid UUID %s should pass validation", validUUID)
		}

		invalidCases := []string{
			"",
			"invalid-uuid",
			"550e8400-e29b-41d4-3716-446655440000", // wrong version
			"550e8400-e29b-41d4-a716-446655440000-extra", // too long
		}

		for _, invalid := range invalidCases {
			if tracer.ValidateTraceID(invalid) {
				t.Errorf("Invalid UUID %s should fail validation", invalid)
			}
		}
	})

	t.Run("ContextOperations", func(t *testing.T) {
		ctx := context.Background()
		traceID := tracer.GenerateTraceID()

		// Test WithTraceID
		ctxWithTrace := tracer.WithTraceID(ctx, traceID)

		// Test GetTraceID
		retrievedID := tracer.GetTraceID(ctxWithTrace)
		if retrievedID != traceID {
			t.Errorf("Expected trace_id %s, got %s", traceID, retrievedID)
		}

		// Test empty context
		emptyID := tracer.GetTraceID(ctx)
		if emptyID != "" {
			t.Errorf("Expected empty trace_id from empty context, got %s", emptyID)
		}
	})
}

func TestHTTPPropagator(t *testing.T) {
	tracer := NewTraceContext()
	propagator := NewHTTPPropagator(tracer)

	t.Run("ExtractHTTP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		traceID := uuid.New().String()
		req.Header.Set(TraceIDHeader, traceID)

		extracted := propagator.ExtractHTTP(req)
		if extracted != traceID {
			t.Errorf("Expected trace_id %s, got %s", traceID, extracted)
		}
	})

	t.Run("ExtractHTTP_Missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		extracted := propagator.ExtractHTTP(req)
		if extracted != "" {
			t.Errorf("Expected empty trace_id, got %s", extracted)
		}
	})

	t.Run("InjectHTTP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		traceID := uuid.New().String()
		ctx := tracer.WithTraceID(context.Background(), traceID)

		propagator.InjectHTTP(ctx, req)

		injected := req.Header.Get(TraceIDHeader)
		if injected != traceID {
			t.Errorf("Expected injected trace_id %s, got %s", traceID, injected)
		}
	})

	t.Run("InjectHTTP_EmptyContext", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.Background()

		propagator.InjectHTTP(ctx, req)

		injected := req.Header.Get(TraceIDHeader)
		if injected != "" {
			t.Errorf("Expected no injected trace_id, got %s", injected)
		}
	})
}

func TestKafkaPropagator(t *testing.T) {
	tracer := NewTraceContext()
	propagator := NewKafkaPropagator(tracer)

	t.Run("ExtractKafka", func(t *testing.T) {
		traceID := uuid.New().String()
		headers := map[string]string{
			TraceIDHeader: traceID,
		}

		extracted := propagator.ExtractKafka(headers)
		if extracted != traceID {
			t.Errorf("Expected trace_id %s, got %s", traceID, extracted)
		}
	})

	t.Run("ExtractKafka_Missing", func(t *testing.T) {
		headers := map[string]string{}
		extracted := propagator.ExtractKafka(headers)
		if extracted != "" {
			t.Errorf("Expected empty trace_id, got %s", extracted)
		}
	})

	t.Run("InjectKafka", func(t *testing.T) {
		traceID := uuid.New().String()
		ctx := tracer.WithTraceID(context.Background(), traceID)
		headers := make(map[string]string)

		propagator.InjectKafka(ctx, headers)

		injected := headers[TraceIDHeader]
		if injected != traceID {
			t.Errorf("Expected injected trace_id %s, got %s", traceID, injected)
		}
	})

	t.Run("InjectKafka_EmptyContext", func(t *testing.T) {
		ctx := context.Background()
		headers := make(map[string]string)

		propagator.InjectKafka(ctx, headers)

		if len(headers) != 0 {
			t.Errorf("Expected no headers injected, got %v", headers)
		}
	})
}

func TestValidateAndGenerateTraceID(t *testing.T) {
	tracer := NewTraceContext()

	t.Run("ValidTraceID", func(t *testing.T) {
		validID := uuid.New().String()
		result, err := ValidateAndGenerateTraceID(validID, tracer)

		if err != nil {
			t.Errorf("Expected no error for valid trace_id, got %v", err)
		}
		if result != validID {
			t.Errorf("Expected same trace_id %s, got %s", validID, result)
		}
	})

	t.Run("EmptyTraceID", func(t *testing.T) {
		result, err := ValidateAndGenerateTraceID("", tracer)

		if err != ErrEmptyTraceID {
			t.Errorf("Expected ErrEmptyTraceID, got %v", err)
		}
		if result == "" {
			t.Error("Expected generated trace_id, got empty string")
		}
		if !tracer.ValidateTraceID(result) {
			t.Errorf("Generated trace_id %s should be valid", result)
		}
	})

	t.Run("InvalidTraceID", func(t *testing.T) {
		invalidID := "invalid-uuid"
		result, err := ValidateAndGenerateTraceID(invalidID, tracer)

		if err == nil {
			t.Error("Expected error for invalid trace_id")
		}
		if result == invalidID {
			t.Error("Should not return invalid trace_id")
		}
		if !tracer.ValidateTraceID(result) {
			t.Errorf("Generated trace_id %s should be valid", result)
		}
	})
}

func TestGinTraceMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracer := NewTraceContext()
	propagator := NewHTTPPropagator(tracer)
	middleware := GinTraceMiddleware(tracer, propagator)

	t.Run("WithValidTraceID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		traceID := uuid.New().String()
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set(TraceIDHeader, traceID)

		middleware(c)

		// Check Gin context
		ginTraceID := GetTraceIDFromGinContext(c)
		if ginTraceID != traceID {
			t.Errorf("Expected trace_id %s in Gin context, got %s", traceID, ginTraceID)
		}

		// Check request context
		ctxTraceID := tracer.GetTraceID(c.Request.Context())
		if ctxTraceID != traceID {
			t.Errorf("Expected trace_id %s in request context, got %s", traceID, ctxTraceID)
		}

		// Check response header
		responseTraceID := w.Header().Get(TraceIDHeader)
		if responseTraceID != traceID {
			t.Errorf("Expected trace_id %s in response header, got %s", traceID, responseTraceID)
		}
	})

	t.Run("WithoutTraceID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		middleware(c)

		// Check that a trace_id was generated
		ginTraceID := GetTraceIDFromGinContext(c)
		if ginTraceID == "" {
			t.Error("Expected generated trace_id in Gin context")
		}
		if !tracer.ValidateTraceID(ginTraceID) {
			t.Errorf("Generated trace_id %s should be valid", ginTraceID)
		}

		// Check consistency across contexts
		ctxTraceID := tracer.GetTraceID(c.Request.Context())
		responseTraceID := w.Header().Get(TraceIDHeader)

		if ginTraceID != ctxTraceID || ginTraceID != responseTraceID {
			t.Errorf("Trace_id should be consistent: gin=%s, ctx=%s, response=%s",
				ginTraceID, ctxTraceID, responseTraceID)
		}
	})

	t.Run("WithInvalidTraceID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set(TraceIDHeader, "invalid-uuid")

		middleware(c)

		// Check that a new valid trace_id was generated
		ginTraceID := GetTraceIDFromGinContext(c)
		if ginTraceID == "" {
			t.Error("Expected generated trace_id in Gin context")
		}
		if ginTraceID == "invalid-uuid" {
			t.Error("Should not use invalid trace_id")
		}
		if !tracer.ValidateTraceID(ginTraceID) {
			t.Errorf("Generated trace_id %s should be valid", ginTraceID)
		}
	})
}

func TestTracingComponents(t *testing.T) {
	components := NewTracingComponents()

	if components.TraceContext == nil {
		t.Error("TraceContext should not be nil")
	}
	if components.HTTPPropagator == nil {
		t.Error("HTTPPropagator should not be nil")
	}
	if components.KafkaPropagator == nil {
		t.Error("KafkaPropagator should not be nil")
	}

	// Test that components work together
	traceID := components.TraceContext.GenerateTraceID()
	ctx := components.TraceContext.WithTraceID(context.Background(), traceID)

	// Test HTTP propagation
	req := httptest.NewRequest("GET", "/test", nil)
	components.HTTPPropagator.InjectHTTP(ctx, req)

	extracted := components.HTTPPropagator.ExtractHTTP(req)
	if extracted != traceID {
		t.Errorf("Expected trace_id %s, got %s", traceID, extracted)
	}

	// Test Kafka propagation
	headers := make(map[string]string)
	components.KafkaPropagator.InjectKafka(ctx, headers)

	kafkaExtracted := components.KafkaPropagator.ExtractKafka(headers)
	if kafkaExtracted != traceID {
		t.Errorf("Expected trace_id %s, got %s", traceID, kafkaExtracted)
	}
}
