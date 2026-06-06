package redis

import "time"

const (
	// DefaultConnectTimeout is the timeout for initial connection ping.
	DefaultConnectTimeout = 5 * time.Second

	// DefaultDB is the default Redis database number
	DefaultDB = 0

	// DefaultPoolSize is the maximum number of socket connections per CPU.
	// Bumped from go-redis default of 10 (per CPU) so notification-srv pub/sub
	// fan-out and analytics cache reads do not block on contention.
	DefaultPoolSize = 50

	// DefaultMinIdleConns is the minimum number of idle connections to keep
	// warm. Eliminates connect-on-first-use latency for hot paths.
	DefaultMinIdleConns = 5

	// DefaultPoolTimeout caps how long a request waits for a free connection
	// before failing fast instead of piling up behind a stalled pool.
	DefaultPoolTimeout = 4 * time.Second
)

// RedisConfig holds Redis configuration.
// Pool tunables left at the zero value fall back to the Default* constants.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int

	PoolSize     int
	MinIdleConns int
	PoolTimeout  time.Duration
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
