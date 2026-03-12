"""
Migration example from analysis-srv Kafka to shared library.

This example shows how to migrate existing Kafka code from analysis-srv
to use the enhanced shared library with tracing support.
"""

import asyncio
import json
from typing import Dict, Any

# OLD: analysis-srv imports
# from pkg.kafka.producer import KafkaProducer
# from pkg.kafka.consumer import KafkaConsumer
# from pkg.kafka.type import KafkaProducerConfig, KafkaConsumerConfig, KafkaMessage

# NEW: shared library imports
from smap_shared.kafka import (
    KafkaProducer,
    KafkaConsumer,
    KafkaProducerConfig,
    KafkaConsumerConfig,
    KafkaMessage,
)
from smap_shared.tracing import set_trace_id, get_trace_id


async def migration_example():
    """
    Example showing migration from analysis-srv to shared library.
    
    The interface remains the same, but now includes automatic tracing!
    """
    
    # Set up trace context (this is new!)
    set_trace_id("550e8400-e29b-41d4-a716-446655440000")
    
    # Producer configuration (same as before)
    producer_config = KafkaProducerConfig(
        bootstrap_servers="localhost:9092",
        client_id="analysis-srv-producer",
        # NEW: Tracing is enabled by default
        enable_trace_injection=True,
        auto_generate_trace_id=True,
    )
    
    # Consumer configuration (same as before)
    consumer_config = KafkaConsumerConfig(
        bootstrap_servers="localhost:9092",
        topics=["analysis-requests"],
        group_id="analysis-consumer-group",
        client_id="analysis-srv-consumer",
        # NEW: Tracing is enabled by default
        enable_trace_extraction=True,
        auto_generate_trace_id=True,
    )
    
    # Create producer (same interface)
    producer = KafkaProducer(producer_config)
    await producer.start()
    
    # Send JSON message (same interface, but now with trace headers!)
    message_data = {
        "post_id": "12345",
        "content": "Sample social media post",
        "timestamp": "2024-01-15T10:30:00Z"
    }
    
    await producer.send_json(
        topic="analysis-requests",
        value=message_data,
        key="post_12345"
    )
    
    print(f"Sent message with trace_id: {get_trace_id()}")
    
    # Create consumer (same interface)
    consumer = KafkaConsumer(consumer_config)
    await consumer.start()
    
    # Message handler (same interface, but trace_id is automatically extracted!)
    async def handle_message(message: KafkaMessage):
        """Handle incoming Kafka message with automatic trace extraction."""
        
        # Parse JSON message (same as before)
        data = json.loads(message.value.decode('utf-8'))
        
        # NEW: trace_id is automatically available!
        print(f"Processing message with trace_id: {message.trace_id}")
        print(f"Current context trace_id: {get_trace_id()}")
        
        # Process the message (your existing logic)
        post_id = data.get("post_id")
        content = data.get("content")
        
        print(f"Processing post {post_id}: {content}")
        
        # Your analysis logic here...
        
    # Start consuming (same interface)
    # Note: This would run indefinitely in a real application
    # await consumer.consume(handle_message)
    
    # Cleanup
    await producer.stop()
    await consumer.stop()


async def backward_compatibility_example():
    """
    Example showing that existing analysis-srv code works without changes.
    
    The shared library provides drop-in replacements that are fully
    backward compatible with the analysis-srv implementation.
    """
    
    # This is exactly the same code as in analysis-srv
    producer_config = KafkaProducerConfig(
        bootstrap_servers="localhost:9092",
        acks="all",
        compression_type="gzip",
        client_id="analysis-srv"
    )
    
    producer = KafkaProducer(producer_config)
    await producer.start()
    
    # Send message exactly like before
    await producer.send_json(
        topic="test-topic",
        value={"message": "Hello World"},
        key="test-key"
    )
    
    await producer.stop()
    
    print("Backward compatibility confirmed - existing code works unchanged!")


async def enhanced_features_example():
    """
    Example showing new enhanced features available in the shared library.
    """
    
    # Enhanced producer with tracing configuration
    producer_config = KafkaProducerConfig(
        bootstrap_servers="localhost:9092",
        client_id="enhanced-producer",
        # Configure tracing behavior
        enable_trace_injection=True,
        auto_generate_trace_id=True,
    )
    
    # Enhanced consumer with tracing configuration  
    consumer_config = KafkaConsumerConfig(
        bootstrap_servers="localhost:9092",
        topics=["enhanced-topic"],
        group_id="enhanced-group",
        # Configure tracing behavior
        enable_trace_extraction=True,
        auto_generate_trace_id=True,
    )
    
    producer = KafkaProducer(producer_config)
    await producer.start()
    
    # Set trace context
    set_trace_id("enhanced-trace-id-12345")
    
    # Send message with automatic trace injection
    await producer.send_json(
        topic="enhanced-topic",
        value={
            "service": "analysis-srv",
            "operation": "sentiment_analysis",
            "data": {"text": "This is great!"}
        }
    )
    
    print(f"Enhanced message sent with trace_id: {get_trace_id()}")
    
    await producer.stop()


if __name__ == "__main__":
    print("=== Kafka Migration Examples ===")
    
    print("\n1. Migration Example:")
    asyncio.run(migration_example())
    
    print("\n2. Backward Compatibility Example:")
    asyncio.run(backward_compatibility_example())
    
    print("\n3. Enhanced Features Example:")
    asyncio.run(enhanced_features_example())
    
    print("\nMigration complete! Your analysis-srv Kafka code now has tracing support.")