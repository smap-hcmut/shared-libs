package discord

import "time"

// Discord API constants
const (
	// Webhook URL template
	webhookURLTemplate = "https://discord.com/api/webhooks/%s/%s"

	// User agent for requests
	UserAgent = "SMAP-Discord-Bot/1.0"

	// Message length limits
	MaxMessageLength  = 2000
	MaxEmbedLength    = 6000
	MaxTitleLen       = 256
	MaxDescriptionLen = 4096
	MaxFieldValueLen  = 1024
	MaxFieldsCount    = 25

	// Default configuration values
	DefaultTimeout    = 30 * time.Second
	DefaultRetryCount = 3
	DefaultRetryDelay = 1 * time.Second
	DefaultUsername   = "SMAP Bot"

	// Embed colors
	ColorInfo    = 0x3498db // Blue
	ColorSuccess = 0x2ecc71 // Green
	ColorWarning = 0xf39c12 // Orange
	ColorError   = 0xe74c3c // Red

	// Special message constants
	ReportBugTitle   = "🐛 Bug Report"
	ActivityLogTitle = "📋 Activity Log"
	ReportBugDescLen = 1900 // Leave room for formatting
)
