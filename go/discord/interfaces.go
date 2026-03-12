package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// IDiscord defines the interface for Discord webhook service with trace integration.
// Implementations are safe for concurrent use.
// Also implements ErrorReporter interface for response package integration.
type IDiscord interface {
	// SendMessage sends a simple text message with trace context
	SendMessage(ctx context.Context, content string) error

	// SendEmbed sends an embed message with options and trace context
	SendEmbed(ctx context.Context, options MessageOptions) error

	// SendError sends an error message with trace context
	SendError(ctx context.Context, title, description string, err error) error

	// SendSuccess sends a success message with trace context
	SendSuccess(ctx context.Context, title, description string) error

	// SendWarning sends a warning message with trace context
	SendWarning(ctx context.Context, title, description string) error

	// SendInfo sends an info message with trace context
	SendInfo(ctx context.Context, title, description string) error

	// ReportBug sends a bug report with trace context (implements ErrorReporter interface)
	ReportBug(ctx context.Context, message string) error

	// SendNotification sends a notification with fields and trace context
	SendNotification(ctx context.Context, title, description string, fields map[string]string) error

	// SendActivityLog sends an activity log with trace context
	SendActivityLog(ctx context.Context, action, user, details string) error

	// GetWebhookURL returns the Discord webhook URL
	GetWebhookURL() string

	// Close closes idle connections
	Close() error
}

// parseWebhookURL extracts id and token from Discord webhook URL.
func parseWebhookURL(webhookURL string) (id, token string, err error) {
	webhookURL = strings.TrimSpace(webhookURL)
	prefix := "https://discord.com/api/webhooks/"
	if !strings.HasPrefix(webhookURL, prefix) {
		return "", "", fmt.Errorf("discord: invalid webhook URL format")
	}
	rest := strings.TrimPrefix(webhookURL, prefix)
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("discord: webhook URL must be .../webhooks/{id}/{token}")
	}
	return parts[0], parts[1], nil
}

// New creates a new Discord service from webhook URL with trace integration.
func New(l log.Logger, webhookURL string) (IDiscord, error) {
	if webhookURL == "" {
		return nil, ErrWebhookRequired
	}
	id, token, err := parseWebhookURL(webhookURL)
	if err != nil {
		return nil, err
	}
	return newImpl(l, id, token)
}

// NewWithTracer creates a new Discord service with custom tracer.
func NewWithTracer(l log.Logger, webhookURL string, tracer tracing.TraceContext) (IDiscord, error) {
	if webhookURL == "" {
		return nil, ErrWebhookRequired
	}
	id, token, err := parseWebhookURL(webhookURL)
	if err != nil {
		return nil, err
	}
	return newImplWithTracer(l, id, token, tracer)
}

// NewWithConfig creates a new Discord service with custom configuration.
func NewWithConfig(l log.Logger, webhookURL string, config Config) (IDiscord, error) {
	if webhookURL == "" {
		return nil, ErrWebhookRequired
	}
	id, token, err := parseWebhookURL(webhookURL)
	if err != nil {
		return nil, err
	}
	return newImplWithConfig(l, id, token, config)
}
