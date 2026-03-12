# Cron Package

The `cron` package provides a modular and extensible interface for scheduling periodic tasks. It follows the standard architectural patterns of the shared libraries, using a factory-based approach with support for functional wrappers (middleware) for cross-cutting concerns like tracing and logging.

## Features

- **Standard Interface**: Common `Cron` interface for job management.
- **Support for Seconds**: Custom parser supporting the standard 6-field cron format (including seconds).
- **Middleware Support**: Easily wrap job execution with functions for tracing, recovery, or logging.
- **Modular Design**: Separated concerns for configuration, interfaces, and implementation.
- **Backward Compatibility**: Maintains support for legacy initialization and global functions.

## New Structure

- `interfaces.go`: Core `Cron` interface, `JobInfo`, and `HandleFunc`.
- `constants.go`: Implementation type constants.
- `config.go`: Configuration structures.
- `factory.go`: Factory function for scheduler instances.
- `errors.go`: Package-specific errors.
- `robfig.go`: Implementation using `robfig/cron/v3`.
- `cron.go`: Global instance and backward compatibility layer.

## Usage

### Using the Global Instance

```go
import "github.com/smap-hcmut/shared-libs/go/cron"

// Add a job
err := cron.AddJob(cron.JobInfo{
    Name:     "CleanupTask",
    CronTime: "0 0 * * * *", // Every hour
    Handler: func() {
        // Your logic here
    },
})

// Start the scheduler
cron.Start()
defer cron.Stop()
```

### Using a Custom Wrapper (Middleware)

```go
cron.SetFuncWrapper(func(f cron.HandleFunc) {
    // Start tracing span
    // ...
    f()
    // Close span
})
```

### Custom Scheduler Instance

```go
cfg := cron.Config{
    Implementation: cron.ImplementationRobfig,
    Robfig: &cron.RobfigConfig{
        WithSeconds: false,
    },
}

c, err := cron.NewCron(cfg)
if err != nil {
    // handle error
}

c.Start()
```
