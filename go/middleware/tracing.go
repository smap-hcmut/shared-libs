package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

const (
	XTraceIDHeader = "X-Trace-Id"
)

// Tracing returns a middleware that handles distributed tracing using X-Trace-Id header.
// It extracts or generates a trace ID and sets it in both the Gin context and the
// Go request context so that the logger can pick it up via ctx.Value(traceContextKey{}).
func Tracing() gin.HandlerFunc {
	tracer := tracing.NewTraceContext()
	return func(c *gin.Context) {
		traceID := c.GetHeader(XTraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set in Gin context (for c.GetString(log.TraceIDKey) usage)
		c.Set(log.TraceIDKey, traceID)

		// Set in Go request context so logger.ctx() can read it via tracer.GetTraceID(ctx)
		ctx := tracer.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// Echo back in response header
		c.Header(XTraceIDHeader, traceID)

		c.Next()
	}
}
