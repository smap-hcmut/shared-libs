package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

func TestClient_Get(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if trace ID was injected
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			t.Error("Expected X-Trace-Id header to be present")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	// Create client
	client := NewDefaultClient()

	// Create context with trace ID
	tracingComponents := tracing.NewTracingComponents()
	traceID := tracingComponents.TraceContext.GenerateTraceID()
	ctx := tracingComponents.TraceContext.WithTraceID(context.Background(), traceID)

	// Make request
	data, statusCode, err := client.Get(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if statusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, statusCode)
	}

	expected := `{"message": "success"}`
	if string(data) != expected {
		t.Errorf("Expected response %s, got %s", expected, string(data))
	}
}

func TestClient_Post(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if trace ID was injected
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			t.Error("Expected X-Trace-Id header to be present")
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123}`))
	}))
	defer server.Close()

	// Create client
	client := NewDefaultClient()

	// Create context with trace ID
	tracingComponents := tracing.NewTracingComponents()
	traceID := tracingComponents.TraceContext.GenerateTraceID()
	ctx := tracingComponents.TraceContext.WithTraceID(context.Background(), traceID)

	// Make request
	body := map[string]string{"name": "test"}
	data, statusCode, err := client.Post(ctx, server.URL, body, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, statusCode)
	}

	expected := `{"id": 123}`
	if string(data) != expected {
		t.Errorf("Expected response %s, got %s", expected, string(data))
	}
}

func TestClient_WithCustomConfig(t *testing.T) {
	config := Config{
		Timeout:   5 * time.Second,
		Retries:   1,
		RetryWait: 100 * time.Millisecond,
	}

	client := NewClient(config)

	// Verify client was created
	if client == nil {
		t.Error("Expected client to be created")
	}
}

func TestTracedHTTPClient_Wrapper(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if trace ID was injected
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			t.Error("Expected X-Trace-Id header to be present")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create traced client wrapper
	baseClient := &http.Client{Timeout: 10 * time.Second}
	tracedClient := NewTracedHTTPClient(baseClient)

	// Create context with trace ID
	tracingComponents := tracing.NewTracingComponents()
	traceID := tracingComponents.TraceContext.GenerateTraceID()
	ctx := tracingComponents.TraceContext.WithTraceID(context.Background(), traceID)

	// Make request
	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := tracedClient.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
