package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/smap/shared-libs/go/tracing"
)

// validateProducerConfig validates producer configuration
func validateProducerConfig(cfg Config) error {
	if len(cfg.Brokers) == 0 {
		return fmt.Errorf("kafka: at least one broker is required")
	}
	if cfg.Topic == "" {
		return fmt.Errorf("kafka: topic is required")
	}
	return nil
}

// newProducerImpl creates a new Kafka producer implementation without tracing
func newProducerImpl(cfg Config) (*producerImpl, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = ProducerRetryMax
	config.Producer.Timeout = ProducerTimeout
	config.Version = KafkaVersion

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	return &producerImpl{producer: producer, topic: cfg.Topic}, nil
}

// newTracedProducerImpl creates a new Kafka producer implementation with tracing
func newTracedProducerImpl(cfg Config) (*tracedProducerImpl, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = ProducerRetryMax
	config.Producer.Timeout = ProducerTimeout
	config.Version = KafkaVersion

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	tracer := tracing.NewTraceContext()
	propagator := tracing.NewKafkaPropagator(tracer)

	return &tracedProducerImpl{
		producer:   producer,
		topic:      cfg.Topic,
		tracer:     tracer,
		propagator: propagator,
	}, nil
}

// Publish sends a message to the configured topic without trace context.
func (p *producerImpl) Publish(key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to publish message to Kafka: %w", err)
	}
	return nil
}

// PublishWithContext sends a message to the configured topic without trace context.
func (p *producerImpl) PublishWithContext(ctx context.Context, key, value []byte) error {
	return p.Publish(key, value)
}

// Close closes the producer.
func (p *producerImpl) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// HealthCheck verifies the producer is initialized.
func (p *producerImpl) HealthCheck() error {
	if p.producer == nil {
		return fmt.Errorf("producer is not initialized")
	}
	return nil
}

// Publish sends a message to the configured topic with automatic trace_id injection.
func (p *tracedProducerImpl) Publish(key, value []byte) error {
	return p.PublishWithContext(context.Background(), key, value)
}

// PublishWithContext sends a message to the configured topic with automatic trace_id injection.
func (p *tracedProducerImpl) PublishWithContext(ctx context.Context, key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	// Inject trace_id into message headers
	headers := make(map[string]string)
	p.propagator.InjectKafka(ctx, headers)

	// Convert to sarama headers
	for k, v := range headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to publish message to Kafka: %w", err)
	}
	return nil
}

// Close closes the traced producer.
func (p *tracedProducerImpl) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// HealthCheck verifies the traced producer is initialized.
func (p *tracedProducerImpl) HealthCheck() error {
	if p.producer == nil {
		return fmt.Errorf("producer is not initialized")
	}
	return nil
}
