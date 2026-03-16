package httpclient

import "context"

// Client defines the interface for HTTP client with retry, timeout, and automatic trace injection.
// Implementations are safe for concurrent use.
type Client interface {
	// Get performs a GET request with automatic trace_id injection
	Get(ctx context.Context, url string, headers map[string]string) ([]byte, int, error)

	// Post performs a POST request with JSON body and automatic trace_id injection
	Post(ctx context.Context, url string, body interface{}, headers map[string]string) ([]byte, int, error)
}

// NewClient creates a new HTTP client with tracing capabilities.
// Returns the Client interface with automatic X-Trace-Id header injection.
func NewClient(cfg Config) Client {
	return newClientImpl(cfg)
}

// NewDefaultClient creates a new HTTP client with default configuration and tracing.
func NewDefaultClient() Client {
	return NewClient(DefaultConfig())
}
