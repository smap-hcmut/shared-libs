package postgres

import (
	"fmt"
	"time"
)

const (
	// DefaultConnectTimeout is the timeout for initial connection ping
	DefaultConnectTimeout = 5 * time.Second

	// DefaultMaxOpenConns is the default maximum number of open connections
	DefaultMaxOpenConns = 25

	// DefaultMaxIdleConns is the default maximum number of idle connections
	DefaultMaxIdleConns = 5

	// DefaultConnMaxLifetime is the default maximum lifetime of a connection
	DefaultConnMaxLifetime = 5 * time.Minute
)

// DefaultConfig returns a PostgreSQL configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		Host:    "localhost",
		Port:    5432,
		SSLMode: "disable",
	}
}

// WithDefaults applies default values to missing configuration fields
func (c Config) WithDefaults() Config {
	defaults := DefaultConfig()

	if c.Host == "" {
		c.Host = defaults.Host
	}
	if c.Port == 0 {
		c.Port = defaults.Port
	}
	if c.SSLMode == "" {
		c.SSLMode = defaults.SSLMode
	}

	return c
}

// ConnectionString builds a PostgreSQL connection string from the config
func (c Config) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}
