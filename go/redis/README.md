# Redis Package

Redis client with trace context integration for logging and debugging.

## Features

- Trace_id integration in Redis operation logs
- Context-aware Redis operations
- Backward compatible with existing Redis usage
- Built on go-redis/redis

## Components

- `client.go` - Redis client with trace context logging
- `config.go` - Redis configuration utilities
- `interfaces.go` - Interface definitions
- `errors.go` - Error definitions

## Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/smap/shared-libs/go/redis"
    "github.com/smap/shared-libs/go/tracing"
)

func main() {
    // Create configuration
    cfg := redis.RedisConfig{
        Host:     "localhost",
        Port:     6379,
        Password: "",
        DB:       0,
    }
    
    // Create client
    client, err := redis.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Create context with trace_id
    tracer := tracing.NewTraceContext()
    ctx := tracer.WithTraceID(context.Background(), tracer.GenerateTraceID())
    
    // Execute operations - will log with trace_id
    err = client.Set(ctx, "user:123", "john_doe", 5*time.Minute)
    if err != nil {
        log.Fatal(err)
    }
    
    value, err := client.Get(ctx, "user:123")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Retrieved value: %s", value)
}
```

## Log Format

With trace_id:
```
trace_id=550e8400-e29b-41d4-a716-446655440000 query=REDIS SET user:123 args=[john_doe 5m0s]
trace_id=550e8400-e29b-41d4-a716-446655440000 query=REDIS GET user:123
```

Without trace_id:
```
query=REDIS SET user:123 args=[john_doe 5m0s]
query=REDIS GET user:123
```