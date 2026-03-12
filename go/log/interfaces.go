package log

import "context"

// Logger interface for trace-aware logging with backward compatibility
type Logger interface {
	// Context-aware logging methods (primary interface)
	Debug(ctx context.Context, args ...any)
	Debugf(ctx context.Context, template string, args ...any)
	Info(ctx context.Context, args ...any)
	Infof(ctx context.Context, template string, args ...any)
	Warn(ctx context.Context, args ...any)
	Warnf(ctx context.Context, template string, args ...any)
	Error(ctx context.Context, args ...any)
	Errorf(ctx context.Context, template string, args ...any)
	DPanic(ctx context.Context, args ...any)
	DPanicf(ctx context.Context, template string, args ...any)
	Panic(ctx context.Context, args ...any)
	Panicf(ctx context.Context, template string, args ...any)
	Fatal(ctx context.Context, args ...any)
	Fatalf(ctx context.Context, template string, args ...any)

	// WithTrace returns a logger that includes trace_id from context
	WithTrace(ctx context.Context) Logger
}
