package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// validateConsumerConfig validates consumer configuration
func validateConsumerConfig(cfg ConsumerConfig) error {
	if len(cfg.Brokers) == 0 {
		return fmt.Errorf("kafka: at least one broker is required")
	}
	if cfg.GroupID == "" {
		return fmt.Errorf("kafka: group ID is required")
	}
	return nil
}

// newConsumerImpl creates a new Kafka consumer group implementation without tracing.
func newConsumerImpl(cfg ConsumerConfig) (*consumerImpl, error) {
	config := sarama.NewConfig()
	config.Version = KafkaVersion
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	return &consumerImpl{
		group: group,
	}, nil
}

// newTracedConsumerImpl creates a new Kafka consumer group implementation with tracing.
func newTracedConsumerImpl(cfg ConsumerConfig) (*tracedConsumerImpl, error) {
	config := sarama.NewConfig()
	config.Version = KafkaVersion
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	tracer := tracing.NewTraceContext()
	propagator := tracing.NewKafkaPropagator(tracer)

	return &tracedConsumerImpl{
		group:      group,
		tracer:     tracer,
		propagator: propagator,
	}, nil
}

// ConsumeWithContext starts consuming from topics with context. Blocks until context is cancelled.
func (c *consumerImpl) ConsumeWithContext(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	if c.group == nil {
		return fmt.Errorf("consumer group not initialized")
	}

	// This will block until the context is cancelled
	return c.group.Consume(ctx, topics, handler)
}

// Consume starts consuming from topics. Uses background context.
// For better control, use ConsumeWithContext instead.
func (c *consumerImpl) Consume(topics []string, handler sarama.ConsumerGroupHandler) error {
	return c.ConsumeWithContext(context.Background(), topics, handler)
}

// Close closes the consumer group.
func (c *consumerImpl) Close() error {
	if c.group != nil {
		return c.group.Close()
	}
	return nil
}

// Errors returns a channel of errors from the consumer.
func (c *consumerImpl) Errors() <-chan error {
	if c.group != nil {
		return c.group.Errors()
	}
	return nil
}

// ConsumeWithContext starts consuming from topics with context and automatic trace_id extraction.
func (c *tracedConsumerImpl) ConsumeWithContext(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	if c.group == nil {
		return fmt.Errorf("consumer group not initialized")
	}

	// Wrap the handler to provide trace context
	wrappedHandler := &tracedConsumerGroupHandlerWrapper{
		handler:    handler,
		tracer:     c.tracer,
		propagator: c.propagator,
	}

	// This will block until the context is cancelled
	return c.group.Consume(ctx, topics, wrappedHandler)
}

// Consume starts consuming from topics with automatic trace_id extraction.
func (c *tracedConsumerImpl) Consume(topics []string, handler sarama.ConsumerGroupHandler) error {
	return c.ConsumeWithContext(context.Background(), topics, handler)
}

// Close closes the traced consumer group.
func (c *tracedConsumerImpl) Close() error {
	if c.group != nil {
		return c.group.Close()
	}
	return nil
}

// Errors returns a channel of errors from the traced consumer.
func (c *tracedConsumerImpl) Errors() <-chan error {
	if c.group != nil {
		return c.group.Errors()
	}
	return nil
}

// NewConsumerGroup creates a new Kafka consumer group (legacy function for backward compatibility).
// Prefer using NewConsumer which returns IConsumer interface.
func NewConsumerGroup(cfg ConsumerConfig) (sarama.ConsumerGroup, error) {
	if err := validateConsumerConfig(cfg); err != nil {
		return nil, err
	}
	config := sarama.NewConfig()
	config.Version = KafkaVersion
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}
	return consumer, nil
}
