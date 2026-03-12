package discord

import "errors"

// Discord package errors
var (
	ErrWebhookRequired = errors.New("discord: webhook URL is required")
	ErrInvalidWebhook  = errors.New("discord: invalid webhook format")
	ErrMessageTooLong  = errors.New("discord: message exceeds maximum length")
	ErrEmbedTooLong    = errors.New("discord: embed exceeds maximum length")
	ErrTooManyFields   = errors.New("discord: too many embed fields")
	ErrRequestFailed   = errors.New("discord: webhook request failed")
)
