package kafka

import (
	"testing"

	"github.com/IBM/sarama"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

func TestConfigBuilder(t *testing.T) {
	config, err := NewConfigBuilder().
		WithBrokers("localhost:9092", "localhost:9093").
		WithTopic("test-topic").
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(config.Brokers) != 2 {
		t.Errorf("Expected 2 brokers, got %d", len(config.Brokers))
	}

	if config.Topic != "test-topic" {
		t.Errorf("Expected topic 'test-topic', got %s", config.Topic)
	}
}

func TestConsumerConfigBuilder(t *testing.T) {
	config, err := NewConsumerConfigBuilder().
		WithBrokers("localhost:9092").
		WithGroupID("test-group").
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(config.Brokers) != 1 {
		t.Errorf("Expected 1 broker, got %d", len(config.Brokers))
	}

	if config.GroupID != "test-group" {
		t.Errorf("Expected group ID 'test-group', got %s", config.GroupID)
	}
}

func TestConfigBuilderFromString(t *testing.T) {
	config, err := NewConfigBuilder().
		WithBrokersFromString("localhost:9092, localhost:9093, localhost:9094").
		WithTopic("test-topic").
		Build()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(config.Brokers) != 3 {
		t.Errorf("Expected 3 brokers, got %d", len(config.Brokers))
	}

	// Check that whitespace is trimmed
	if config.Brokers[1] != "localhost:9093" {
		t.Errorf("Expected 'localhost:9093', got '%s'", config.Brokers[1])
	}
}

func TestValidateProducerConfig(t *testing.T) {
	// Test valid config
	validConfig := Config{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
	}
	if err := validateProducerConfig(validConfig); err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}

	// Test missing brokers
	invalidConfig1 := Config{
		Brokers: []string{},
		Topic:   "test-topic",
	}
	if err := validateProducerConfig(invalidConfig1); err == nil {
		t.Error("Expected error for missing brokers")
	}

	// Test missing topic
	invalidConfig2 := Config{
		Brokers: []string{"localhost:9092"},
		Topic:   "",
	}
	if err := validateProducerConfig(invalidConfig2); err == nil {
		t.Error("Expected error for missing topic")
	}
}
func TestValidateConsumerConfig(t *testing.T) {
	// Test valid config
	validConfig := ConsumerConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "test-group",
	}
	if err := validateConsumerConfig(validConfig); err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}

	// Test missing brokers
	invalidConfig1 := ConsumerConfig{
		Brokers: []string{},
		GroupID: "test-group",
	}
	if err := validateConsumerConfig(invalidConfig1); err == nil {
		t.Error("Expected error for missing brokers")
	}

	// Test missing group ID
	invalidConfig2 := ConsumerConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "",
	}
	if err := validateConsumerConfig(invalidConfig2); err == nil {
		t.Error("Expected error for missing group ID")
	}
}

// Mock consumer group handler for testing
type mockConsumerGroupHandler struct {
	setupCalled   bool
	cleanupCalled bool
	consumeCalled bool
	lastMessage   *sarama.ConsumerMessage
}

func (m *mockConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	m.setupCalled = true
	return nil
}

func (m *mockConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	m.cleanupCalled = true
	return nil
}

func (m *mockConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	m.consumeCalled = true
	for message := range claim.Messages() {
		m.lastMessage = message
		session.MarkMessage(message, "")
	}
	return nil
}

func TestTracedConsumerGroupHandlerWrapper(t *testing.T) {
	tracer := tracing.NewTraceContext()
	propagator := tracing.NewKafkaPropagator(tracer)

	mockHandler := &mockConsumerGroupHandler{}
	wrapper := &tracedConsumerGroupHandlerWrapper{
		handler:    mockHandler,
		tracer:     tracer,
		propagator: propagator,
	}

	// Test Setup
	if err := wrapper.Setup(nil); err != nil {
		t.Errorf("Expected no error from Setup, got %v", err)
	}
	if !mockHandler.setupCalled {
		t.Error("Expected Setup to be called on wrapped handler")
	}

	// Test Cleanup
	if err := wrapper.Cleanup(nil); err != nil {
		t.Errorf("Expected no error from Cleanup, got %v", err)
	}
	if !mockHandler.cleanupCalled {
		t.Error("Expected Cleanup to be called on wrapped handler")
	}
}
