package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/log"
)

// Logger returns a middleware that logs HTTP requests at the appropriate log level
// based on response status code:
//   - 5xx → Error
//   - 4xx → Warn
//   - 2xx/3xx → Info
//
// Health check endpoints (/health, /ready, /live) are skipped.
func Logger(l log.Logger, environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		if path == "/health" || path == "/ready" || path == "/live" {
			return
		}

		latency := time.Since(start)
		status := c.Writer.Status()
		ctx := c.Request.Context()

		if environment == "production" {
			msg := "HTTP Request - Method: %s, Path: %s, Status: %d, IP: %s, Latency: %v, UserAgent: %s, Query: %s"
			args := []any{c.Request.Method, path, status, c.ClientIP(), latency, c.Request.UserAgent(), query}
			switch {
			case status >= 500:
				l.Errorf(ctx, msg, args...)
			case status >= 400:
				l.Warnf(ctx, msg, args...)
			default:
				l.Infof(ctx, msg, args...)
			}
		} else {
			msg := "%s %s %d %s %s"
			args := []any{c.Request.Method, path, status, latency, c.ClientIP()}
			switch {
			case status >= 500:
				l.Errorf(ctx, msg, args...)
			case status >= 400:
				l.Warnf(ctx, msg, args...)
			default:
				l.Infof(ctx, msg, args...)
			}
		}
	}
}
