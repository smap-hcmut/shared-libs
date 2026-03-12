package rabbitmq

import "time"

const (
	// Connection retry configuration
	RetryConnectionDelay   = 2 * time.Second
	RetryConnectionTimeout = 20 * time.Second

	// Content types
	ContentTypePlainText = "text/plain"
	ContentTypeJSON      = "application/json"

	// Exchange types
	ExchangeTypeDirect = "direct"
	ExchangeTypeFanout = "fanout"
	ExchangeTypeTopic  = "topic"

	// Trace header key for RabbitMQ messages
	TraceIDHeader = "X-Trace-Id"
)
