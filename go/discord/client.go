package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// newHTTPClient creates a new HTTP client with timeout configuration.
func newHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}
}

// newImpl creates a new Discord implementation with default tracer.
func newImpl(l log.Logger, id, token string) (IDiscord, error) {
	return newImplWithTracer(l, id, token, tracing.NewTraceContext())
}

// newImplWithTracer creates a new Discord implementation with custom tracer.
func newImplWithTracer(l log.Logger, id, token string, tracer tracing.TraceContext) (IDiscord, error) {
	return newImplWithConfig(l, id, token, DefaultConfig())
}

// newImplWithConfig creates a new Discord implementation with custom configuration.
func newImplWithConfig(l log.Logger, id, token string, config Config) (IDiscord, error) {
	if id == "" || token == "" {
		return nil, ErrWebhookRequired
	}

	client := newHTTPClient(config.Timeout)

	return &discordImpl{
		l:       l,
		webhook: &webhookInfo{id: id, token: token},
		config:  config,
		client:  client,
		tracer:  tracing.NewTraceContext(),
	}, nil
}

// GetWebhookURL returns the Discord webhook URL.
func (d *discordImpl) GetWebhookURL() string {
	return fmt.Sprintf(webhookURLTemplate, d.webhook.id, d.webhook.token)
}

// Close closes idle connections in the HTTP client.
func (d *discordImpl) Close() error {
	if d.client != nil {
		d.client.CloseIdleConnections()
	}
	return nil
}

// logOperation logs Discord operations with trace_id if available.
func (d *discordImpl) logOperation(ctx context.Context, operation, details string) {
	if d.l == nil {
		return
	}

	traceID := d.tracer.GetTraceID(ctx)
	if traceID != "" {
		d.l.Infof(ctx, "discord.%s: %s (trace_id=%s)", operation, details, traceID)
	} else {
		d.l.Infof(ctx, "discord.%s: %s", operation, details)
	}
}

// sendWithRetry sends a webhook payload with retry logic and trace integration.
func (d *discordImpl) sendWithRetry(ctx context.Context, payload *WebhookPayload) error {
	d.logOperation(ctx, "sendWithRetry", fmt.Sprintf("attempting to send webhook message"))

	var lastErr error
	for attempt := 0; attempt <= d.config.RetryCount; attempt++ {
		if attempt > 0 {
			d.logOperation(ctx, "sendWithRetry", fmt.Sprintf("retrying attempt %d/%d", attempt, d.config.RetryCount))
			time.Sleep(d.config.RetryDelay)
		}

		err := d.sendRequest(ctx, payload)
		if err == nil {
			d.logOperation(ctx, "sendWithRetry", "webhook message sent successfully")
			return nil
		}

		lastErr = err
		if d.l != nil {
			d.l.Warnf(ctx, "discord.sendWithRetry: attempt %d failed: %v", attempt+1, err)
		}
	}

	return fmt.Errorf("failed after %d attempts, last error: %w", d.config.RetryCount+1, lastErr)
}

// sendRequest sends a single HTTP request to Discord webhook.
func (d *discordImpl) sendRequest(ctx context.Context, payload *WebhookPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := d.GetWebhookURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	// Add trace_id to request headers if available
	if traceID := d.tracer.GetTraceID(ctx); traceID != "" {
		req.Header.Set("X-Trace-Id", traceID)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("discord webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Validation methods
func (d *discordImpl) validateMessageLength(content string) error {
	if len(content) > MaxMessageLength {
		return fmt.Errorf("message too long: %d characters (max: %d)", len(content), MaxMessageLength)
	}
	return nil
}

func (d *discordImpl) validateEmbedLength(embed *Embed) error {
	total := len(embed.Title) + len(embed.Description)
	for _, f := range embed.Fields {
		total += len(f.Name) + len(f.Value)
	}
	if total > MaxEmbedLength {
		return fmt.Errorf("embed too long: %d characters (max: %d)", total, MaxEmbedLength)
	}
	if len(embed.Fields) > MaxFieldsCount {
		return fmt.Errorf("too many fields: %d (max: %d)", len(embed.Fields), MaxFieldsCount)
	}
	return nil
}

// Utility methods
func (d *discordImpl) getColorForType(msgType MessageType) int {
	switch msgType {
	case MessageTypeInfo:
		return ColorInfo
	case MessageTypeSuccess:
		return ColorSuccess
	case MessageTypeWarning:
		return ColorWarning
	case MessageTypeError:
		return ColorError
	default:
		return ColorInfo
	}
}

func (d *discordImpl) formatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (d *discordImpl) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
