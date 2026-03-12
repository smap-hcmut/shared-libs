package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

func TestTraceMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ExtractExistingTraceID", func(t *testing.T) {
		// Setup
		router := gin.New()
		router.Use(TraceMiddleware())

		var extractedTraceID string
		router.GET("/test", func(c *gin.Context) {
			tracingComponents := tracing.NewTracingComponents()
			extractedTraceID = tracingComponents.TraceContext.GetTraceID(c.Request.Context())
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Create request with trace ID
		traceID := "550e8400-e29b-41d4-a716-446655440000"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Trace-Id", traceID)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify
		if extractedTraceID != traceID {
			t.Errorf("Expected trace ID %s, got %s", traceID, extractedTraceID)
		}
	})

	t.Run("GenerateNewTraceID", func(t *testing.T) {
		// Setup
		router := gin.New()
		router.Use(TraceMiddleware())

		var extractedTraceID string
		router.GET("/test", func(c *gin.Context) {
			tracingComponents := tracing.NewTracingComponents()
			extractedTraceID = tracingComponents.TraceContext.GetTraceID(c.Request.Context())
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Create request without trace ID
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify new trace ID was generated
		if extractedTraceID == "" {
			t.Error("Expected trace ID to be generated")
		}

		// Verify it's a valid UUID v4
		tracingComponents := tracing.NewTracingComponents()
		if !tracingComponents.TraceContext.ValidateTraceID(extractedTraceID) {
			t.Errorf("Generated trace ID %s is not valid UUID v4", extractedTraceID)
		}
	})

	t.Run("ReplaceInvalidTraceID", func(t *testing.T) {
		// Setup
		router := gin.New()
		router.Use(TraceMiddleware())

		var extractedTraceID string
		router.GET("/test", func(c *gin.Context) {
			tracingComponents := tracing.NewTracingComponents()
			extractedTraceID = tracingComponents.TraceContext.GetTraceID(c.Request.Context())
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Create request with invalid trace ID
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Trace-Id", "invalid-trace-id")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify new trace ID was generated (not the invalid one)
		if extractedTraceID == "invalid-trace-id" {
			t.Error("Invalid trace ID should have been replaced")
		}

		if extractedTraceID == "" {
			t.Error("Expected new trace ID to be generated")
		}

		// Verify it's a valid UUID v4
		tracingComponents := tracing.NewTracingComponents()
		if !tracingComponents.TraceContext.ValidateTraceID(extractedTraceID) {
			t.Errorf("Generated trace ID %s is not valid UUID v4", extractedTraceID)
		}
	})
}

func TestStandardTraceMiddleware(t *testing.T) {
	t.Run("ExtractExistingTraceID", func(t *testing.T) {
		var extractedTraceID string

		// Create handler that extracts trace ID
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracingComponents := tracing.NewTracingComponents()
			extractedTraceID = tracingComponents.TraceContext.GetTraceID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		// Wrap with trace middleware
		wrappedHandler := StandardTraceMiddleware(handler)

		// Create request with trace ID
		traceID := "550e8400-e29b-41d4-a716-446655440000"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Trace-Id", traceID)

		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Verify
		if extractedTraceID != traceID {
			t.Errorf("Expected trace ID %s, got %s", traceID, extractedTraceID)
		}
	})

	t.Run("GenerateNewTraceID", func(t *testing.T) {
		var extractedTraceID string

		// Create handler that extracts trace ID
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracingComponents := tracing.NewTracingComponents()
			extractedTraceID = tracingComponents.TraceContext.GetTraceID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		// Wrap with trace middleware
		wrappedHandler := StandardTraceMiddleware(handler)

		// Create request without trace ID
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(w, req)

		// Verify new trace ID was generated
		if extractedTraceID == "" {
			t.Error("Expected trace ID to be generated")
		}

		// Verify it's a valid UUID v4
		tracingComponents := tracing.NewTracingComponents()
		if !tracingComponents.TraceContext.ValidateTraceID(extractedTraceID) {
			t.Errorf("Generated trace ID %s is not valid UUID v4", extractedTraceID)
		}
	})
}
