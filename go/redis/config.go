package redis

import "time"

const (
	// DefaultConnectTimeout is the timeout for initial connection ping.
	DefaultConnectTimeout = 5 * time.Second

	// DefaultDB is the default Redis database number
	DefaultDB = 0
)

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// DefaultConfig returns a Redis configuration with sensible defaults
func DefaultConfig() RedisConfig {
	return RedisConfig{
		Host: "localhost",
		Port: 6379,
		DB:   DefaultDB,
	}
}

// WithDefaults applies default values to missing configuration fields
func (c RedisConfig) WithDefaults() RedisConfig {
	defaults := DefaultConfig()

	if c.Host == "" {
		c.Host = defaults.Host
	}
	if c.Port == 0 {
		c.Port = defaults.Port
	}
	// DB can be 0, so we don't set a default for it

	return c
}
