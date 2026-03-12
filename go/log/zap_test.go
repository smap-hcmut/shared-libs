package log

import (
	"context"
	"testing"

	"github.com/smap/shared-libs/go/tracing"
)

func TestZapLogger_BasicLogging(t *testing.T) {
	// Test basic logger creation and logging
	logger := NewZapLogger(DefaultDevelopmentConfig)

	ctx := context.Background()

	// Test basic logging methods
	logger.Info(ctx, "test info message")
	logger.Debug(ctx, "test debug message")
	logger.Warn(ctx, "test warn message")
	logger.Error(ctx, "test error message")

	// Test formatted logging methods
	logger.Infof(ctx, "test info message with format: %s", "value")
	logger.Debugf(ctx, "test debug message with format: %d", 42)
}

func TestZapLogger_WithTraceID(t *testing.T) {
	// Test logger with trace_id integration
	logger := NewZapLogger(DefaultDevelopmentConfig)

	tracer := tracing.NewTraceContext()
	traceID := tracer.GenerateTraceID()

	// Create context with trace_id
	ctx := tracer.WithTraceID(context.Background(), traceID)

	// Test logging with trace_id
	logger.Info(ctx, "message with trace_id")
	logger.Errorf(ctx, "error message with trace_id: %s", "some error")

	// Test WithTrace method
	tracedLogger := logger.WithTrace(ctx)
	tracedLogger.Info(context.Background(), "message from traced logger")
}

func TestZapLogger_WithoutTraceID(t *testing.T) {
	// Test logger without trace_id (should work normally)
	logger := NewZapLogger(DefaultDevelopmentConfig)

	ctx := context.Background()

	// Test logging without trace_id
	logger.Info(ctx, "message without trace_id")
	logger.Warn(ctx, "warning without trace_id")
}

func TestZapLogger_JSONMode(t *testing.T) {
	// Test JSON logging mode
	config := ZapConfig{
		Level:        LevelInfo,
		Mode:         ModeProduction,
		Encoding:     EncodingJSON,
		ColorEnabled: false,
	}

	logger := NewZapLogger(config)

	tracer := tracing.NewTraceContext()
	traceID := tracer.GenerateTraceID()
	ctx := tracer.WithTraceID(context.Background(), traceID)

	// Test JSON logging with trace_id
	logger.Info(ctx, "JSON log message with trace_id")
	logger.Error(ctx, "JSON error message")
}

func TestZapLogger_BackwardCompatibility(t *testing.T) {
	// Test backward compatibility with Init function
	logger := Init(DefaultDevelopmentConfig)

	ctx := context.Background()

	// Test that Init still works
	logger.Info(ctx, "backward compatibility test")
	logger.Debugf(ctx, "formatted message: %s", "test")
}
