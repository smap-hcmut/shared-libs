package kafka

import (
	"context"

	"github.com/IBM/sarama"
)

//go:generate mockery --name=IProducer
// IProducer defines the interface for Kafka producer.
// Implementations are safe for concurrent use.
type IProducer interface {
	Publish(key, value []byte) error
	PublishWithContext(ctx context.Context, key, value []byte) error
	Close() error
	HealthCheck() error
}

// IConsumer defines the interface for Kafka consumer group.
// Wraps sarama.ConsumerGroup for easier testing and management.
type IConsumer interface {
	// Consume starts consuming from topics. Uses background context.
	Consume(topics []string, handler sarama.ConsumerGroupHandler) error
	// ConsumeWithContext starts consuming from topics with context. Blocks until context is cancelled.
	ConsumeWithContext(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error
	// Close closes the consumer group
	Close() error
	// Errors returns a channel of errors from the consumer
	Errors() <-chan error
}

// TracedConsumerGroupHandler wraps sarama.ConsumerGroupHandler to provide trace context
type TracedConsumerGroupHandler interface {
	sarama.ConsumerGroupHandler
	// SetupWithTrace is called at the beginning of a new session, before ConsumeClaim
	SetupWithTrace(sarama.ConsumerGroupSession) error
	// CleanupWithTrace is called at the end of a session, once all ConsumeClaim goroutines have exited
	CleanupWithTrace(sarama.ConsumerGroupSession) error
	// ConsumeClaimWithTrace should start a consumer loop of ConsumerGroupClaim's Messages().
	ConsumeClaimWithTrace(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
}

// NewProducer creates a new Kafka producer with trace integration.
func NewProducer(cfg Config) (IProducer, error) {
	if err := validateProducerConfig(cfg); err != nil {
		return nil, err
	}
	return newProducerImpl(cfg)
}

// NewConsumer creates a new Kafka consumer group with trace integration.
func NewConsumer(cfg ConsumerConfig) (IConsumer, error) {
	if err := validateConsumerConfig(cfg); err != nil {
		return nil, err
	}
	return newConsumerImpl(cfg)
}

// NewTracedProducer creates a new Kafka producer with automatic trace_id injection.
// This is the recommended way to create producers for trace-aware applications.
func NewTracedProducer(cfg Config) (IProducer, error) {
	if err := validateProducerConfig(cfg); err != nil {
		return nil, err
	}
	return newTracedProducerImpl(cfg)
}

// NewTracedConsumer creates a new Kafka consumer group with automatic trace_id extraction.
// This is the recommended way to create consumers for trace-aware applications.
func NewTracedConsumer(cfg ConsumerConfig) (IConsumer, error) {
	if err := validateConsumerConfig(cfg); err != nil {
		return nil, err
	}
	return newTracedConsumerImpl(cfg)
}
