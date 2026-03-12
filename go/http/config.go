package http

import "time"

const (
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
	// DefaultRetries is the default number of retries.
	DefaultRetries = 3
	// DefaultRetryWait is the default wait between retries.
	DefaultRetryWait = 1 * time.Second
)

// Config holds configuration for the HTTP client.
type Config struct {
	Timeout   time.Duration
	Retries   int
	RetryWait time.Duration
}

// DefaultConfig returns default Config.
func DefaultConfig() Config {
	return Config{
		Timeout:   DefaultTimeout,
		Retries:   DefaultRetries,
		RetryWait: DefaultRetryWait,
	}
}
