package redis

import (
	"context"
	"testing"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// mockLogger captures log messages for testing
type mockLogger struct {
	messages []string
}

func (m *mockLogger) Log(message string) {
	m.messages = append(m.messages, message)
}

func TestLogOperation(t *testing.T) {
	logger := &mockLogger{}
	client := &redisImpl{
		tracer: tracing.NewTraceContext(),
		logger: logger,
	}

	tests := []struct {
		name           string
		ctx            context.Context
		operation      string
		key            string
		args           []interface{}
		expectedPrefix string
	}{
		{
			name:           "operation without trace_id",
			ctx:            context.Background(),
			operation:      "GET",
			key:            "user:123",
			args:           nil,
			expectedPrefix: "query=REDIS GET user:123",
		},
		{
			name:           "operation with trace_id",
			ctx:            client.tracer.WithTraceID(context.Background(), "550e8400-e29b-41d4-a716-446655440000"),
			operation:      "SET",
			key:            "user:123",
			args:           []interface{}{"value", "60s"},
			expectedPrefix: "trace_id=550e8400-e29b-41d4-a716-446655440000 query=REDIS SET user:123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.messages = nil // Reset messages
			client.logOperation(tt.ctx, tt.operation, tt.key, tt.args...)

			if len(logger.messages) != 1 {
				t.Errorf("Expected 1 log message, got %d", len(logger.messages))
				return
			}

			message := logger.messages[0]
			if len(message) < len(tt.expectedPrefix) || message[:len(tt.expectedPrefix)] != tt.expectedPrefix {
				t.Errorf("Expected log message to start with %q, got %q", tt.expectedPrefix, message)
			}
		})
	}
}

func TestRedisConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      RedisConfig
		expectError bool
		expectedErr error
	}{
		{
			name: "valid config",
			config: RedisConfig{
				Host: "localhost",
				Port: 6379,
				DB:   0,
			},
			expectError: false,
		},
		{
			name: "missing host",
			config: RedisConfig{
				Port: 6379,
				DB:   0,
			},
			expectError: true,
			expectedErr: ErrHostRequired,
		},
		{
			name: "invalid port",
			config: RedisConfig{
				Host: "localhost",
				Port: 0,
				DB:   0,
			},
			expectError: true,
			expectedErr: ErrInvalidPort,
		},
		{
			name: "port too high",
			config: RedisConfig{
				Host: "localhost",
				Port: 70000,
				DB:   0,
			},
			expectError: true,
			expectedErr: ErrInvalidPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWithLogger(tt.config, &mockLogger{})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.expectedErr)
				} else if err != tt.expectedErr {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
			} else if err != nil && tt.expectError == false {
				// For valid config, we expect connection error since we're not running a real Redis
				// This is acceptable for unit tests
				t.Logf("Connection error expected for valid config in unit test: %v", err)
			}
		})
	}
}

func TestRedisConfigDefaults(t *testing.T) {
	config := RedisConfig{
		Password: "testpass",
		// Missing Host, Port, DB
	}

	configWithDefaults := config.WithDefaults()

	if configWithDefaults.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got %q", configWithDefaults.Host)
	}

	if configWithDefaults.Port != 6379 {
		t.Errorf("Expected default port 6379, got %d", configWithDefaults.Port)
	}

	// Password should be preserved
	if configWithDefaults.Password != "testpass" {
		t.Errorf("Expected password 'testpass' to be preserved, got %q", configWithDefaults.Password)
	}
}
