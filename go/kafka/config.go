package kafka

import (
	"fmt"
	"strings"
)

// ConfigBuilder helps build Kafka configurations
type ConfigBuilder struct {
	config Config
}

// ConsumerConfigBuilder helps build Kafka consumer configurations
type ConsumerConfigBuilder struct {
	config ConsumerConfig
}

// NewConfigBuilder creates a new Kafka configuration builder
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: Config{
			Brokers: make([]string, 0),
		},
	}
}

// NewConsumerConfigBuilder creates a new Kafka consumer configuration builder
func NewConsumerConfigBuilder() *ConsumerConfigBuilder {
	return &ConsumerConfigBuilder{
		config: ConsumerConfig{
			Brokers: make([]string, 0),
		},
	}
}

// WithBrokers sets the Kafka brokers
func (b *ConfigBuilder) WithBrokers(brokers ...string) *ConfigBuilder {
	b.config.Brokers = brokers
	return b
}

// WithBrokersFromString sets the Kafka brokers from a comma-separated string
func (b *ConfigBuilder) WithBrokersFromString(brokers string) *ConfigBuilder {
	if brokers != "" {
		b.config.Brokers = strings.Split(brokers, ",")
		// Trim whitespace from each broker
		for i, broker := range b.config.Brokers {
			b.config.Brokers[i] = strings.TrimSpace(broker)
		}
	}
	return b
}

// WithTopic sets the Kafka topic
func (b *ConfigBuilder) WithTopic(topic string) *ConfigBuilder {
	b.config.Topic = topic
	return b
}

// Build returns the built configuration
func (b *ConfigBuilder) Build() (Config, error) {
	if err := validateProducerConfig(b.config); err != nil {
		return Config{}, fmt.Errorf("invalid producer config: %w", err)
	}
	return b.config, nil
}

// WithBrokers sets the Kafka brokers
func (b *ConsumerConfigBuilder) WithBrokers(brokers ...string) *ConsumerConfigBuilder {
	b.config.Brokers = brokers
	return b
}

// WithBrokersFromString sets the Kafka brokers from a comma-separated string
func (b *ConsumerConfigBuilder) WithBrokersFromString(brokers string) *ConsumerConfigBuilder {
	if brokers != "" {
		b.config.Brokers = strings.Split(brokers, ",")
		// Trim whitespace from each broker
		for i, broker := range b.config.Brokers {
			b.config.Brokers[i] = strings.TrimSpace(broker)
		}
	}
	return b
}

// WithGroupID sets the consumer group ID
func (b *ConsumerConfigBuilder) WithGroupID(groupID string) *ConsumerConfigBuilder {
	b.config.GroupID = groupID
	return b
}

// Build returns the built consumer configuration
func (b *ConsumerConfigBuilder) Build() (ConsumerConfig, error) {
	if err := validateConsumerConfig(b.config); err != nil {
		return ConsumerConfig{}, fmt.Errorf("invalid consumer config: %w", err)
	}
	return b.config, nil
}
