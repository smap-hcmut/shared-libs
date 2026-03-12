package kafka

import (
	"github.com/IBM/sarama"
	"github.com/smap/shared-libs/go/tracing"
)

// Config holds configuration for Kafka producer.
type Config struct {
	Brokers []string
	Topic   string
}

// ConsumerConfig holds configuration for Kafka consumer group.
type ConsumerConfig struct {
	Brokers []string
	GroupID string
}

// producerImpl implements IProducer without tracing.
type producerImpl struct {
	producer sarama.SyncProducer
	topic    string
}

// consumerImpl implements IConsumer without tracing.
type consumerImpl struct {
	group sarama.ConsumerGroup
}

// tracedProducerImpl implements IProducer with automatic trace_id injection.
type tracedProducerImpl struct {
	producer   sarama.SyncProducer
	topic      string
	tracer     tracing.TraceContext
	propagator tracing.KafkaPropagator
}

// tracedConsumerImpl implements IConsumer with automatic trace_id extraction.
type tracedConsumerImpl struct {
	group      sarama.ConsumerGroup
	tracer     tracing.TraceContext
	propagator tracing.KafkaPropagator
}

// tracedConsumerGroupHandlerWrapper wraps a regular ConsumerGroupHandler to provide trace context
type tracedConsumerGroupHandlerWrapper struct {
	handler    sarama.ConsumerGroupHandler
	tracer     tracing.TraceContext
	propagator tracing.KafkaPropagator
}
