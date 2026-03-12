package tracing

import (
	"context"
	"net/http"
)

const (
	// TraceIDHeader is the standard header name for trace_id propagation
	TraceIDHeader = "X-Trace-Id"
)

// httpPropagatorImpl implements the HTTPPropagator interface
type httpPropagatorImpl struct {
	tracer TraceContext
}

// NewHTTPPropagator creates a new HTTPPropagator implementation
func NewHTTPPropagator(tracer TraceContext) HTTPPropagator {
	return &httpPropagatorImpl{
		tracer: tracer,
	}
}

// InjectHTTP adds trace_id to outbound HTTP request headers
func (h *httpPropagatorImpl) InjectHTTP(ctx context.Context, req *http.Request) {
	traceID := h.tracer.GetTraceID(ctx)
	if traceID != "" {
		req.Header.Set(TraceIDHeader, traceID)
	}
}

// ExtractHTTP retrieves trace_id from inbound HTTP request headers
func (h *httpPropagatorImpl) ExtractHTTP(req *http.Request) string {
	return req.Header.Get(TraceIDHeader)
}
