package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// Client wraps sql.DB with trace_id injection in query logs
type Client struct {
	db     *sql.DB
	tracer tracing.TraceContext
	logger Logger

	logQueries bool
}

// Ensure Client implements IPostgres
var _ IPostgres = (*Client)(nil)

// Logger interface for database operation logging
type Logger interface {
	Log(message string)
}

// defaultLogger provides a simple logging implementation
type defaultLogger struct{}

func (d *defaultLogger) Log(message string) {
	log.Println(message)
}

// Config holds PostgreSQL configuration.
// Pool tunables left at the zero value fall back to the Default* constants;
// callers that need bespoke sizing override per-service.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	// Pool tunables. Zero values fall back to the Default* constants.
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	ConnMaxIdleTime  time.Duration
	StatementTimeout time.Duration

	// LogQueries enables per-query trace logging. Off by default because
	// it produces one structured log line per SQL statement which, at
	// steady state across multiple services, generates significant I/O
	// noise through the fluent-bit → Loki pipeline.
	LogQueries bool
}

// New creates a new PostgreSQL client with trace logging
func New(cfg Config) (IPostgres, error) {
	return NewWithLogger(cfg, &defaultLogger{})
}

// NewWithLogger creates a new PostgreSQL client with custom logger
func NewWithLogger(cfg Config, logger Logger) (IPostgres, error) {
	if cfg.Host == "" {
		return nil, ErrHostRequired
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, ErrInvalidPort
	}
	if cfg.User == "" {
		return nil, ErrUserRequired
	}
	if cfg.DBName == "" {
		return nil, ErrDBNameRequired
	}

	// Set default SSL mode if not specified
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	// Build connection string with libpq-style options so the timeouts apply
	// to every connection the pool opens, not just the first one we exec on.
	timeout := cfg.StatementTimeout
	if timeout <= 0 {
		timeout = DefaultStatementTimeout
	}
	timeoutMs := timeout.Milliseconds()
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s "+
			"options='-c statement_timeout=%d -c idle_in_transaction_session_timeout=%d'",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		timeoutMs, timeoutMs,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	applyPoolDefaults(db, cfg)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{
		db:         db,
		tracer:     tracing.NewTraceContext(),
		logger:     logger,
		logQueries: cfg.LogQueries,
	}, nil
}

// applyPoolDefaults wires the Default* constants into the pool when the caller
// did not override them. Without this every Go service was running on the
// stdlib defaults (0 max idle conns, no lifetime cap) regardless of what the
// Default* constants advertised.
func applyPoolDefaults(db *sql.DB, cfg Config) {
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = DefaultMaxOpenConns
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = DefaultMaxIdleConns
	}
	lifetime := cfg.ConnMaxLifetime
	if lifetime <= 0 {
		lifetime = DefaultConnMaxLifetime
	}
	idleTime := cfg.ConnMaxIdleTime
	if idleTime <= 0 {
		idleTime = DefaultConnMaxIdleTime
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(lifetime)
	db.SetConnMaxIdleTime(idleTime)
}


// logQuery logs the query with trace_id if available, only when the
// caller opted into query logging via Config.LogQueries. Defaulting to
// off avoids blasting every SQL statement through structured logging
// (→ fluent-bit → Loki) on every service at steady state.
func (c *Client) logQuery(ctx context.Context, query string, args ...interface{}) {
	if !c.logQueries {
		return
	}
	traceID := c.tracer.GetTraceID(ctx)

	var logMessage string
	if traceID != "" {
		logMessage = fmt.Sprintf("trace_id=%s query=%s", traceID, query)
	} else {
		logMessage = fmt.Sprintf("query=%s", query)
	}

	if len(args) > 0 {
		logMessage += fmt.Sprintf(" args=%v", args)
	}

	c.logger.Log(logMessage)
}

// QueryContext executes a query with trace logging
func (c *Client) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	c.logQuery(ctx, query, args...)
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns at most one row with trace logging
func (c *Client) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	c.logQuery(ctx, query, args...)
	return c.db.QueryRowContext(ctx, query, args...)
}

// ExecContext executes a query without returning any rows with trace logging
func (c *Client) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	c.logQuery(ctx, query, args...)
	return c.db.ExecContext(ctx, query, args...)
}

// PrepareContext creates a prepared statement with trace logging
func (c *Client) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	c.logQuery(ctx, fmt.Sprintf("PREPARE: %s", query))
	return c.db.PrepareContext(ctx, query)
}

// BeginTx starts a transaction with trace logging
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	c.logQuery(ctx, "BEGIN TRANSACTION")
	return c.db.BeginTx(ctx, opts)
}

// Ping verifies a connection to the database is still alive
func (c *Client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// GetDB returns the underlying sql.DB for advanced operations
func (c *Client) GetDB() *sql.DB {
	return c.db
}

// SetMaxOpenConns sets the maximum number of open connections to the database
func (c *Client) SetMaxOpenConns(n int) {
	c.db.SetMaxOpenConns(n)
}

// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
func (c *Client) SetMaxIdleConns(n int) {
	c.db.SetMaxIdleConns(n)
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
func (c *Client) SetConnMaxLifetime(d time.Duration) {
	c.db.SetConnMaxLifetime(d)
}

// Stats returns database statistics
func (c *Client) Stats() sql.DBStats {
	return c.db.Stats()
}
