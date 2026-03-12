package log

import (
	"context"
	"os"
	"testing"
)

func TestFactoryFunctions(t *testing.T) {
	ctx := context.Background()

	// Test NewDevelopmentLogger
	devLogger := NewDevelopmentLogger()
	devLogger.Info(ctx, "development logger test")

	// Test NewProductionLogger
	prodLogger := NewProductionLogger()
	prodLogger.Info(ctx, "production logger test")

	// Test NewLogger with custom config
	customConfig := ZapConfig{
		Level:        LevelDebug,
		Mode:         ModeDevelopment,
		Encoding:     EncodingConsole,
		ColorEnabled: true,
	}
	customLogger := NewLogger(customConfig)
	customLogger.Debug(ctx, "custom logger test")
}

func TestNewLoggerFromEnv(t *testing.T) {
	// Test with default environment (no env vars set)
	logger1 := NewLoggerFromEnv()
	logger1.Info(context.Background(), "default env logger test")

	// Test with custom environment variables
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_MODE", "production")
	os.Setenv("LOG_ENCODING", "json")
	os.Setenv("LOG_COLOR", "false")

	logger2 := NewLoggerFromEnv()
	logger2.Debug(context.Background(), "custom env logger test")

	// Clean up environment variables
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_MODE")
	os.Unsetenv("LOG_ENCODING")
	os.Unsetenv("LOG_COLOR")
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with non-existent env var
	value1 := getEnvOrDefault("NON_EXISTENT_VAR", "default_value")
	if value1 != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", value1)
	}

	// Test with existing env var
	os.Setenv("TEST_VAR", "test_value")
	value2 := getEnvOrDefault("TEST_VAR", "default_value")
	if value2 != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value2)
	}

	// Clean up
	os.Unsetenv("TEST_VAR")
}
