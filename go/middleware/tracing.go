package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/smap-hcmut/shared-libs/go/log"
)

const (
	XTraceIDHeader = "X-Trace-Id"
)

// Tracing returns a middleware that handles distributed tracing using X-Trace-Id header.
// It extracts or generates a trace ID and sets it in the request context and response headers.
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(XTraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set trace id in context for pkg/log to pick up
		c.Set(log.TraceIDKey, traceID)

		// Also set in response header
		c.Header(XTraceIDHeader, traceID)

		c.Next()
	}
}
