package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

//go:generate mockery --name=IRedis
// IRedis defines the interface for Redis operations with trace logging.
// Implementations are safe for concurrent use.
type IRedis interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Close() error
	Ping(ctx context.Context) error
	GetClient() *goredis.Client
}

// Logger interface for Redis operation logging
type Logger interface {
	Log(message string)
}

// defaultLogger provides a simple logging implementation
type defaultLogger struct{}

func (d *defaultLogger) Log(message string) {
	log.Println(message)
}

// redisImpl implements IRedis using go-redis with trace logging
type redisImpl struct {
	client *goredis.Client
	tracer tracing.TraceContext
	logger Logger
}

// New creates a new Redis client with trace logging. Returns an implementation of IRedis.
func New(cfg RedisConfig) (IRedis, error) {
	return NewWithLogger(cfg, &defaultLogger{})
}

// NewWithLogger creates a new Redis client with custom logger
func NewWithLogger(cfg RedisConfig, logger Logger) (IRedis, error) {
	if cfg.Host == "" {
		return nil, ErrHostRequired
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, ErrInvalidPort
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), DefaultConnectTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisImpl{
		client: client,
		tracer: tracing.NewTraceContext(),
		logger: logger,
	}, nil
}

// logOperation logs the Redis operation with trace_id if available
func (c *redisImpl) logOperation(ctx context.Context, operation, key string, args ...interface{}) {
	traceID := c.tracer.GetTraceID(ctx)

	var logMessage string
	if traceID != "" {
		// Format: "trace_id={uuid} query={redis_operation}"
		logMessage = fmt.Sprintf("trace_id=%s query=REDIS %s %s", traceID, operation, key)
	} else {
		// Graceful handling when no trace_id exists
		logMessage = fmt.Sprintf("query=REDIS %s %s", operation, key)
	}

	if len(args) > 0 {
		logMessage += fmt.Sprintf(" args=%v", args)
	}

	c.logger.Log(logMessage)
}

// Set stores a key-value pair with TTL.
func (c *redisImpl) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.logOperation(ctx, "SET", key, value, ttl)
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Get retrieves a value by key.
func (c *redisImpl) Get(ctx context.Context, key string) (string, error) {
	c.logOperation(ctx, "GET", key)
	return c.client.Get(ctx, key).Result()
}

// Delete removes keys.
func (c *redisImpl) Delete(ctx context.Context, keys ...string) error {
	c.logOperation(ctx, "DEL", fmt.Sprintf("[%v]", keys))
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists.
func (c *redisImpl) Exists(ctx context.Context, key string) (bool, error) {
	c.logOperation(ctx, "EXISTS", key)
	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// TTL returns the remaining TTL of a key.
func (c *redisImpl) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.logOperation(ctx, "TTL", key)
	return c.client.TTL(ctx, key).Result()
}

// Close closes the Redis connection.
func (c *redisImpl) Close() error {
	return c.client.Close()
}

// Ping checks if Redis is reachable.
// Not logged — Ping is used for health checks and would generate excessive noise.
func (c *redisImpl) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetClient returns the underlying go-redis Client for advanced operations.
func (c *redisImpl) GetClient() *goredis.Client {
	return c.client
}
