# Postgres Package

PostgreSQL client with trace_id injection in query logs.

## Features

- Trace_id logging format: "trace_id={uuid} query={sql}"
- Context-aware database operations
- Graceful handling when no trace_id exists
- Built on lib/pq driver

## Components

- `client.go` - PostgreSQL client with trace logging
- `config.go` - Database configuration utilities
- `interfaces.go` - Interface definitions
- `utils.go` - UUID utilities migrated from services
- `errors.go` - Error definitions

## Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/smap/shared-libs/go/postgres"
    "github.com/smap/shared-libs/go/tracing"
)

func main() {
    // Create configuration
    cfg := postgres.Config{
        Host:     "localhost",
        Port:     5432,
        User:     "myuser",
        Password: "mypass",
        DBName:   "mydb",
        SSLMode:  "disable",
    }
    
    // Create client
    client, err := postgres.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Create context with trace_id
    tracer := tracing.NewTraceContext()
    ctx := tracer.WithTraceID(context.Background(), tracer.GenerateTraceID())
    
    // Execute query - will log with trace_id
    rows, err := client.QueryContext(ctx, "SELECT id, name FROM users WHERE active = $1", true)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    // Process results...
}
```

## Log Format

With trace_id:
```
trace_id=550e8400-e29b-41d4-a716-446655440000 query=SELECT id, name FROM users WHERE active = $1 args=[true]
```

Without trace_id:
```
query=SELECT id, name FROM users WHERE active = $1 args=[true]
```