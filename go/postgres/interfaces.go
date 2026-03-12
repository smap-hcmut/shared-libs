package postgres

import (
	"context"
	"database/sql"
	"time"
)

// IPostgres defines the interface for PostgreSQL operations with trace logging.
// Implementations are safe for concurrent use.
type IPostgres interface {
	// Query operations with trace logging
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)

	// Transaction operations
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	// Connection management
	Ping(ctx context.Context) error
	Close() error

	// Advanced operations
	GetDB() *sql.DB
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	SetConnMaxLifetime(d time.Duration)
	Stats() sql.DBStats
}

// Ensure Client implements IPostgres
var _ IPostgres = (*Client)(nil)
