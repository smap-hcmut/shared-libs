// Package llm provides a multi-provider LLM abstraction with automatic fallback.
//
// All providers use the OpenAI Chat Completions API (which Gemini, DeepSeek,
// and Qwen also support). Providers are treated equally — no primary — and
// requests are distributed round-robin with fallback on transient errors.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LLM is the interface consumed by application code.
// It is a drop-in replacement for gemini.IGemini.
type LLM interface {
	// Generate sends a prompt and returns the generated text.
	Generate(ctx context.Context, prompt string) (string, error)
	// Name returns a human-readable identifier (e.g. "gemini", "openai/gpt-4o-mini").
	Name() string
}

// ProviderConfig configures a single LLM backend.
type ProviderConfig struct {
	// Name is a human-readable label, e.g. "gemini", "openai", "deepseek", "qwen".
	Name string
	// BaseURL is the full Chat Completions endpoint URL.
	// Example: "https://api.openai.com/v1/chat/completions"
	BaseURL string
	// APIKey is the bearer token sent in the Authorization header.
	APIKey string
	// Model is the model ID, e.g. "gpt-4o-mini", "gemini-2.0-flash".
	Model string
	// Timeout per request. Default: 60s.
	Timeout time.Duration
	// MaxRetries per single provider call. Default: 1 (total 2 attempts).
	MaxRetries int
}

// provider implements LLM for a single OpenAI-compatible endpoint.
type provider struct {
	name    string
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
	retries int
}

// Well-known provider endpoints and default models.
var KnownProviders = map[string]struct {
	BaseURL      string
	DefaultModel string
}{
	"gemini": {
		BaseURL:      "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
		DefaultModel: "gemini-2.0-flash",
	},
	"openai": {
		BaseURL:      "https://api.openai.com/v1/chat/completions",
		DefaultModel: "gpt-4o-mini",
	},
	"deepseek": {
		BaseURL:      "https://api.deepseek.com/v1/chat/completions",
		DefaultModel: "deepseek-chat",
	},
	"qwen": {
		BaseURL:      "https://dashscope-intl.aliyuncs.com/compatible-mode/v1/chat/completions",
		DefaultModel: "qwen-turbo",
	},
}

// NewProvider creates a single LLM provider. If BaseURL or Model are empty and
// Name matches a known provider, defaults are filled in automatically.
func NewProvider(cfg ProviderConfig) (LLM, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("llm: API key is required for provider %q", cfg.Name)
	}

	// Fill defaults from known providers
	if known, ok := KnownProviders[strings.ToLower(cfg.Name)]; ok {
		if cfg.BaseURL == "" {
			cfg.BaseURL = known.BaseURL
		}
		if cfg.Model == "" {
			cfg.Model = known.DefaultModel
		}
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("llm: base URL is required for provider %q", cfg.Name)
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("llm: model is required for provider %q", cfg.Name)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	retries := cfg.MaxRetries
	if retries <= 0 {
		retries = 1
	}

	return &provider{
		name:    cfg.Name,
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		client:  &http.Client{Timeout: timeout},
		retries: retries,
	}, nil
}

func (p *provider) Name() string {
	return fmt.Sprintf("%s/%s", p.name, p.model)
}

// Generate implements LLM using the OpenAI Chat Completions API.
func (p *provider) Generate(ctx context.Context, prompt string) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= p.retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
		text, err := p.doRequest(ctx, prompt)
		if err == nil {
			return text, nil
		}
		lastErr = err
		if !isRetryable(err) {
			return "", fmt.Errorf("[%s] %w", p.name, err)
		}
	}
	return "", fmt.Errorf("[%s] all retries exhausted: %w", p.name, lastErr)
}

// ── OpenAI Chat Completions request/response types ──────────────────────────

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

func (p *provider) doRequest(ctx context.Context, prompt string) (string, error) {
	reqBody := chatRequest{
		Model: p.model,
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", &httpError{StatusCode: 0, Message: err.Error()}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", &httpError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", &httpError{
			StatusCode: resp.StatusCode,
			Message:    chatResp.Error.Message,
		}
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ── Error classification ────────────────────────────────────────────────────

type httpError struct {
	StatusCode int
	Message    string
}

func (e *httpError) Error() string {
	return e.Message
}

func isRetryable(err error) bool {
	he, ok := err.(*httpError)
	if !ok {
		return true // network errors are retryable
	}
	switch he.StatusCode {
	case 429, 500, 502, 503, 504, 529:
		return true
	}
	return false
}

// isQuotaError returns true if the error indicates the provider's quota is
// exhausted (as opposed to a transient rate-limit that will resolve soon).
func isQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	quotaPatterns := []string{
		"insufficient_quota",
		"insufficient balance",
		"quota exceeded",
		"resource_exhausted",
		"billing",
		"account_deactivated",
	}
	for _, p := range quotaPatterns {
		if strings.Contains(msg, p) {
			return true
		}
	}
	return false
}
