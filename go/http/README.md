# HTTP Package

HTTP client with automatic trace_id propagation and middleware for incoming requests.

## Features

- Automatic X-Trace-Id header injection for outbound requests
- HTTP middleware for trace_id extraction from incoming requests
- Backward compatible with existing HTTP client usage
- Built on standard net/http package
- Retry logic with configurable timeout and retry attempts
- Support for both Gin and standard net/http middleware

## Components

- `client.go` - Enhanced HTTP client with automatic trace injection and retry logic
- `wrapper.go` - TracedHTTPClient wrapper for existing http.Client instances
- `middleware.go` - HTTP middleware for trace extraction (Gin and standard)
- `interfaces.go` - Client interface definitions
- `config.go` - HTTP client configuration

## Usage

### Basic HTTP Client

```go
import "github.com/smap/smap-shared-libs/go/http"

// Create client with default config
client := http.NewDefaultClient()

// Make requests with automatic trace injection
data, statusCode, err := client.Get(ctx, "https://api.example.com/users", nil)
```

### Custom Configuration

```go
config := http.Config{
    Timeout:   60 * time.Second,
    Retries:   5,
    RetryWait: 2 * time.Second,
}
client := http.NewClient(config)
```

### Wrapping Existing http.Client

```go
existingClient := &http.Client{Timeout: 30 * time.Second}
tracedClient := http.NewTracedHTTPClient(existingClient)

req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
resp, err := tracedClient.Do(req)
```

### Middleware Usage

```go
// For Gin framework
router := gin.New()
router.Use(http.TraceMiddleware())

// For standard net/http
mux := http.NewServeMux()
handler := http.StandardTraceMiddleware(mux)
```