package http

import (
	"context"
	"net/http"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// TracedHTTPClient wraps an existing http.Client with automatic trace_id injection.
// This provides backward compatibility for services that already use http.Client directly.
type TracedHTTPClient struct {
	client     *http.Client
	propagator tracing.HTTPPropagator
}

// NewTracedHTTPClient wraps an existing http.Client with tracing capabilities.
// This is useful for services that need to add tracing to existing http.Client instances.
func NewTracedHTTPClient(client *http.Client) *TracedHTTPClient {
	if client == nil {
		client = &http.Client{}
	}

	// Create tracing components
	tracingComponents := tracing.NewTracingComponents()

	return &TracedHTTPClient{
		client:     client,
		propagator: tracingComponents.HTTPPropagator,
	}
}

// Do executes an HTTP request with automatic trace_id injection.
// This method maintains the same signature as http.Client.Do() for compatibility.
func (t *TracedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Inject trace_id into request headers
	t.propagator.InjectHTTP(req.Context(), req)

	// Execute the request using the wrapped client
	return t.client.Do(req)
}

// DoWithContext executes an HTTP request with context and automatic trace_id injection.
// This is a convenience method that creates a new request with the provided context.
func (t *TracedHTTPClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Create new request with context
	reqWithCtx := req.WithContext(ctx)

	// Inject trace_id into request headers
	t.propagator.InjectHTTP(ctx, reqWithCtx)

	// Execute the request using the wrapped client
	return t.client.Do(reqWithCtx)
}

// Get is a convenience method for GET requests with automatic trace injection
func (t *TracedHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return t.Do(req)
}

// Post is a convenience method for POST requests with automatic trace injection
func (t *TracedHTTPClient) Post(ctx context.Context, url, contentType string, body interface{}) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return t.Do(req)
}
