package kafka_test

import (
	"fmt"
	"log"

	"github.com/smap-hcmut/shared-libs/go/kafka"
)

// ExampleNewTracedProducer demonstrates how to create and use a traced Kafka producer
func ExampleNewTracedProducer() {
	// Build configuration using the builder pattern
	config, err := kafka.NewConfigBuilder().
		WithBrokers("localhost:9092").
		WithTopic("user-events").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Producer config created: %d brokers, topic: %s\n",
		len(config.Brokers), config.Topic)

	// Output: Producer config created: 1 brokers, topic: user-events
}

// ExampleNewTracedConsumer demonstrates how to create and use a traced Kafka consumer
func ExampleNewTracedConsumer() {
	// Build consumer configuration
	config, err := kafka.NewConsumerConfigBuilder().
		WithBrokers("localhost:9092").
		WithGroupID("event-processor").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Consumer config created: %d brokers, group: %s\n",
		len(config.Brokers), config.GroupID)

	// Output: Consumer config created: 1 brokers, group: event-processor
}

// Example_configBuilder demonstrates configuration building
func Example_configBuilder() {
	// Producer configuration
	producerConfig, err := kafka.NewConfigBuilder().
		WithBrokers("broker1:9092", "broker2:9092").
		WithTopic("my-topic").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Producer config: %d brokers, topic: %s\n",
		len(producerConfig.Brokers), producerConfig.Topic)

	// Consumer configuration
	consumerConfig, err := kafka.NewConsumerConfigBuilder().
		WithBrokersFromString("broker1:9092,broker2:9092").
		WithGroupID("my-consumer-group").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Consumer config: %d brokers, group: %s\n",
		len(consumerConfig.Brokers), consumerConfig.GroupID)

	// Output:
	// Producer config: 2 brokers, topic: my-topic
	// Consumer config: 2 brokers, group: my-consumer-group
}

// Example_migrationFromServicePackage shows how to migrate from service-specific packages
func Example_migrationFromServicePackage() {
	// Build configuration for migration example
	config, err := kafka.NewConfigBuilder().
		WithBrokers("localhost:9092").
		WithTopic("events").
		Build()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Migration config: %d brokers, topic: %s\n",
		len(config.Brokers), config.Topic)

	// Output: Migration config: 1 brokers, topic: events
}
