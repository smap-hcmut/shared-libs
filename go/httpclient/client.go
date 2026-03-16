package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// clientImpl implements Client interface with automatic trace injection
type clientImpl struct {
	client     *http.Client
	config     Config
	propagator tracing.HTTPPropagator
}

// newClientImpl creates a new HTTP client implementation with tracing
func newClientImpl(cfg Config) *clientImpl {
	// Create tracing components
	tracingComponents := tracing.NewTracingComponents()

	return &clientImpl{
		client:     defaultHTTPClient(cfg.Timeout),
		config:     cfg,
		propagator: tracingComponents.HTTPPropagator,
	}
}

// defaultHTTPClient creates a standard HTTP client with timeout
func defaultHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

// Get performs a GET request with automatic trace_id injection.
func (c *clientImpl) Get(ctx context.Context, url string, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	return c.do(ctx, req, headers)
}

// Post performs a POST request with JSON body and automatic trace_id injection.
func (c *clientImpl) Post(ctx context.Context, url string, body interface{}, headers map[string]string) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(ctx, req, headers)
}

// do executes the HTTP request with retry logic and automatic trace injection
func (c *clientImpl) do(ctx context.Context, req *http.Request, headers map[string]string) ([]byte, int, error) {
	// Inject trace_id into request headers (this is the key enhancement)
	c.propagator.InjectHTTP(ctx, req)

	// Set custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	var resp *http.Response
	var err error

	// Retry logic
	for i := 0; i <= c.config.Retries; i++ {
		resp, err = c.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}
		if i < c.config.Retries {
			time.Sleep(c.config.RetryWait)
		}
	}

	if err != nil {
		return nil, 0, fmt.Errorf("request failed after %d retries: %w", c.config.Retries, err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, resp.StatusCode, nil
}
