package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/smap/shared-libs/go/tracing"
)

// TraceMiddleware creates HTTP middleware for trace_id extraction and generation.
// This middleware extracts trace_id from X-Trace-Id header or generates a new one.
// The trace_id is stored in the request context for use throughout the request lifecycle.
func TraceMiddleware() gin.HandlerFunc {
	// Create tracing components
	tracingComponents := tracing.NewTracingComponents()

	return func(c *gin.Context) {
		// Extract trace_id from headers
		traceID := tracingComponents.HTTPPropagator.ExtractHTTP(c.Request)

		// Validate and generate new trace_id if needed
		if traceID == "" || !tracingComponents.TraceContext.ValidateTraceID(traceID) {
			traceID = tracingComponents.TraceContext.GenerateTraceID()
		}

		// Store trace_id in request context
		ctx := tracingComponents.TraceContext.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// Continue with request processing
		c.Next()
	}
}

// StandardTraceMiddleware creates HTTP middleware for standard net/http handlers.
// This is for services that don't use Gin framework.
func StandardTraceMiddleware(next http.Handler) http.Handler {
	// Create tracing components
	tracingComponents := tracing.NewTracingComponents()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace_id from headers
		traceID := tracingComponents.HTTPPropagator.ExtractHTTP(r)

		// Validate and generate new trace_id if needed
		if traceID == "" || !tracingComponents.TraceContext.ValidateTraceID(traceID) {
			traceID = tracingComponents.TraceContext.GenerateTraceID()
		}

		// Store trace_id in request context
		ctx := tracingComponents.TraceContext.WithTraceID(r.Context(), traceID)
		r = r.WithContext(ctx)

		// Continue with request processing
		next.ServeHTTP(w, r)
	})
}
