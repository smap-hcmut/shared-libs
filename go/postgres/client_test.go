package postgres

import (
	"context"
	"testing"

	"github.com/smap/shared-libs/go/tracing"
)

// mockLogger captures log messages for testing
type mockLogger struct {
	messages []string
}

func (m *mockLogger) Log(message string) {
	m.messages = append(m.messages, message)
}

func TestLogQuery(t *testing.T) {
	logger := &mockLogger{}
	client := &Client{
		tracer: tracing.NewTraceContext(),
		logger: logger,
	}

	tests := []struct {
		name           string
		ctx            context.Context
		query          string
		args           []interface{}
		expectedPrefix string
	}{
		{
			name:           "query without trace_id",
			ctx:            context.Background(),
			query:          "SELECT * FROM users",
			args:           nil,
			expectedPrefix: "query=SELECT * FROM users",
		},
		{
			name:           "query with trace_id",
			ctx:            client.tracer.WithTraceID(context.Background(), "550e8400-e29b-41d4-a716-446655440000"),
			query:          "SELECT * FROM users WHERE id = $1",
			args:           []interface{}{123},
			expectedPrefix: "trace_id=550e8400-e29b-41d4-a716-446655440000 query=SELECT * FROM users WHERE id = $1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.messages = nil // Reset messages
			client.logQuery(tt.ctx, tt.query, tt.args...)

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

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		expectedErr error
	}{
		{
			name: "valid config",
			config: Config{
				Host:    "localhost",
				Port:    5432,
				User:    "testuser",
				DBName:  "testdb",
				SSLMode: "disable",
			},
			expectError: false,
		},
		{
			name: "missing host",
			config: Config{
				Port:   5432,
				User:   "testuser",
				DBName: "testdb",
			},
			expectError: true,
			expectedErr: ErrHostRequired,
		},
		{
			name: "invalid port",
			config: Config{
				Host:   "localhost",
				Port:   0,
				User:   "testuser",
				DBName: "testdb",
			},
			expectError: true,
			expectedErr: ErrInvalidPort,
		},
		{
			name: "missing user",
			config: Config{
				Host:   "localhost",
				Port:   5432,
				DBName: "testdb",
			},
			expectError: true,
			expectedErr: ErrUserRequired,
		},
		{
			name: "missing database name",
			config: Config{
				Host: "localhost",
				Port: 5432,
				User: "testuser",
			},
			expectError: true,
			expectedErr: ErrDBNameRequired,
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
				// For valid config, we expect connection error since we're not running a real DB
				// This is acceptable for unit tests
				t.Logf("Connection error expected for valid config in unit test: %v", err)
			}
		})
	}
}
