package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/smap/shared-libs/go/tracing"
)

// Client wraps sql.DB with trace_id injection in query logs
type Client struct {
	db     *sql.DB
	tracer tracing.TraceContext
	logger Logger
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

// Config holds PostgreSQL configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
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

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Client{
		db:     db,
		tracer: tracing.NewTraceContext(),
		logger: logger,
	}, nil
}

// logQuery logs the query with trace_id if available
func (c *Client) logQuery(ctx context.Context, query string, args ...interface{}) {
	traceID := c.tracer.GetTraceID(ctx)

	var logMessage string
	if traceID != "" {
		// Format: "trace_id={uuid} query={sql}"
		logMessage = fmt.Sprintf("trace_id=%s query=%s", traceID, query)
	} else {
		// Graceful handling when no trace_id exists
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
