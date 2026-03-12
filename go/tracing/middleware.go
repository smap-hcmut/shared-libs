package tracing

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
)

// GinTraceMiddleware creates a Gin middleware for trace_id management
// This builds upon the existing tracing patterns in identity-srv and project-srv
func GinTraceMiddleware(tracer TraceContext, propagator HTTPPropagator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace_id from headers
		traceID := propagator.ExtractHTTP(c.Request)

		// Validate and generate new trace_id if needed
		validTraceID, err := ValidateAndGenerateTraceID(traceID, tracer)
		if err != nil {
			// Log validation failure for debugging (non-blocking)
			log.Printf("WARN: %v, generated new trace_id: %s", err, validTraceID)
		}

		// Set in Gin Context for handlers (maintains compatibility with existing code)
		c.Set("trace_id", validTraceID)

		// Set in request.Context() for standard library and loggers
		ctx := tracer.WithTraceID(c.Request.Context(), validTraceID)
		c.Request = c.Request.WithContext(ctx)

		// Always return it to the client/downstream
		c.Header(TraceIDHeader, validTraceID)

		c.Next()
	}
}

// GetTraceIDFromGinContext extracts trace_id from Gin context
// This maintains compatibility with existing service code
func GetTraceIDFromGinContext(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceIDFromContext extracts trace_id from standard context
// This is the preferred method for new code
func GetTraceIDFromContext(ctx context.Context, tracer TraceContext) string {
	return tracer.GetTraceID(ctx)
}

// WithTraceIDInContext adds trace_id to context
// Helper function for manual context management
func WithTraceIDInContext(ctx context.Context, traceID string, tracer TraceContext) context.Context {
	return tracer.WithTraceID(ctx, traceID)
}
