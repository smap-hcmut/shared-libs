"""
Migration example: From analysis-srv Kafka to smap-shared-libs traced Kafka.

This example shows how to migrate from the existing analysis-srv Kafka implementation
to the enhanced traced Kafka implementation in smap-shared-libs.

Key improvements:
- Automatic trace_id injection for producers
- Automatic trace_id extraction for consumers  
- Backward compatibility with existing interfaces
- Enhanced error handling and logging
- Cross-language compatibility with Go services
"""

import asyncio
import json
from typing import Dict, Any

# OLD: analysis-srv imports
# from pkg.kafka.producer import KafkaProducer
# from pkg.kafka.consumer import KafkaConsumer
# from pkg.kafka.type import KafkaProducerConfig, KafkaConsumerConfig, KafkaMessage

# NEW: smap-shared-libs imports
from smap_shared.kafka import (
    TracedKafkaProducer,
    TracedKafkaConsumer,
    KafkaProducerConfig,
    KafkaConsumerConfig,
    KafkaMessage,
)
from smap_shared.tracing import set_trace_id, get_trace_id


async def migration_example():
    """
    Complete migration example showing before/after patterns.
    """
    
    # =============================================================================
    # PRODUCER MIGRATION EXAMPLE
    # =============================================================================
    
    print("=== PRODUCER MIGRATION EXAMPLE ===")
    
    # OLD: analysis-srv producer configuration
    old_producer_config = KafkaProducerConfig(
        bootstrap_servers="localhost:9092",
        acks="all",
        compression_type="snappy",
        enable_idempotence=True,
        linger_ms=5,
    )
    
    # NEW: Enhanced configuration with tracing support
    new_producer_config = KafkaProducerConfig(
        bootstrap_servers="localhost:9092",
        acks="all", 
        compression_type="snappy",
        enable_idempotence=True,
        linger_ms=5,
        # NEW: Tracing configuration
        enable_trace_injection=True,      # Automatically inject trace_id
        auto_generate_trace_id=True,      # Generate trace_id if missing
    )
    
    # Create traced producer (same interface as old producer)
    producer = TracedKafkaProducer(new_producer_config)
    await producer.start()
    
    # Set trace_id in context (simulating incoming request)
    set_trace_id("550e8400-e29b-41d4-a716-446655440000")
    
    # OLD: Manual message sending
    # await producer.send(
    #     topic="analytics-results",
    #     value=json.dumps({"result": "positive"}).encode(),
    #     key=b"user-123"
    # )
    
    # NEW: Same interface, but with automatic trace_id injection
    await producer.send(
        topic="analytics-results", 
        value=json.dumps({"result": "positive"}).encode(),
        key=b"user-123"
        # trace_id automatically injected into headers!
    )
    
    # NEW: Convenient JSON sending with tracing
    await producer.send_json(
        topic="analytics-results",
        value={"result": "positive", "confidence": 0.95},
        key="user-123"
        # trace_id automatically injected!
    )
    
    print(f"✅ Producer sent messages with trace_id: {get_trace_id()}")
    
    await producer.stop()
    
    # =============================================================================
    # CONSUMER MIGRATION EXAMPLE  
    # =============================================================================
    
    print("\n=== CONSUMER MIGRATION EXAMPLE ===")
    
    # OLD: analysis-srv consumer configuration
    old_consumer_config = KafkaConsumerConfig(
        bootstrap_servers="localhost:9092",
        topics=["analytics-tasks"],
        group_id="analytics-group",
        auto_offset_reset="earliest",
        enable_auto_commit=False,
        max_poll_records=10,
    )
    
    # NEW: Enhanced configuration with tracing support
    new_consumer_config = KafkaConsumerConfig(
        bootstrap_servers="localhost:9092",
        topics=["analytics-tasks"],
        group_id="analytics-group", 
        auto_offset_reset="earliest",
        enable_auto_commit=False,
        max_poll_records=10,
        # NEW: Tracing configuration
        enable_trace_extraction=True,     # Automatically extract trace_id
        auto_generate_trace_id=True,      # Generate trace_id if missing
    )
    
    # Create traced consumer (same interface as old consumer)
    consumer = TracedKafkaConsumer(new_consumer_config)
    await consumer.start()
    
    # OLD: Message handler without tracing
    # async def old_message_handler(message: KafkaMessage) -> None:
    #     data = json.loads(message.value.decode())
    #     print(f"Processing: {data}")
    #     # No trace context available
    
    # NEW: Message handler with automatic trace context
    async def new_message_handler(message: KafkaMessage) -> None:
        # trace_id automatically extracted and set in context!
        current_trace_id = get_trace_id()
        
        data = json.loads(message.value.decode())
        print(f"Processing: {data} with trace_id: {current_trace_id}")
        print(f"Message trace_id: {message.trace_id}")  # Also available on message
        
        # All subsequent operations will have trace context
        # (database queries, HTTP calls, etc.)
    
    print("✅ Consumer ready to process messages with trace extraction")
    
    # Same consumption pattern as before
    # await consumer.consume(new_message_handler)
    
    await consumer.stop()


async def backward_compatibility_example():
    """
    Shows how existing analysis-srv code works without changes.
    """
    
    print("\n=== BACKWARD COMPATIBILITY EXAMPLE ===")
    
    # Existing analysis-srv code pattern works unchanged:
    
    config = KafkaProducerConfig(
        bootstrap_servers="localhost:9092",
        acks="all",
        # Tracing features are optional - defaults to enabled
    )
    
    producer = TracedKafkaProducer(config)
    await producer.start()
    
    # Existing send patterns work exactly the same
    await producer.send(
        topic="test-topic",
        value=b"test message",
        key=b"test-key"
    )
    
    # JSON convenience method (new feature)
    await producer.send_json(
        topic="test-topic", 
        value={"message": "hello"},
        key="test-key"
    )
    
    await producer.stop()
    
    print("✅ Existing code works without any changes")


async def advanced_tracing_example():
    """
    Shows advanced tracing features and patterns.
    """
    
    print("\n=== ADVANCED TRACING EXAMPLE ===")
    
    # Producer with custom tracing configuration
    producer_config = KafkaProducerConfig(
        bootstrap_servers="localhost:9092",
        enable_trace_injection=True,
        auto_generate_trace_id=False,  # Don't auto-generate, require explicit trace_id
    )
    
    producer = TracedKafkaProducer(producer_config)
    await producer.start()
    
    # Scenario 1: Request with existing trace_id
    set_trace_id("550e8400-e29b-41d4-a716-446655440000")
    await producer.send_json(
        topic="user-events",
        value={"event": "login", "user_id": "123"},
        key="user-123"
    )
    print(f"✅ Sent with existing trace_id: {get_trace_id()}")
    
    # Scenario 2: Request without trace_id (will not auto-generate)
    set_trace_id(None)  # Clear trace_id
    await producer.send_json(
        topic="user-events", 
        value={"event": "logout", "user_id": "123"},
        key="user-123"
        # No trace_id will be injected (auto_generate_trace_id=False)
    )
    print("✅ Sent without trace_id (as configured)")
    
    await producer.stop()
    
    # Consumer with trace extraction
    consumer_config = KafkaConsumerConfig(
        bootstrap_servers="localhost:9092",
        topics=["user-events"],
        group_id="event-processor",
        enable_trace_extraction=True,
        auto_generate_trace_id=True,  # Generate if missing
    )
    
    consumer = TracedKafkaConsumer(consumer_config)
    await consumer.start()
    
    async def trace_aware_handler(message: KafkaMessage) -> None:
        trace_id = get_trace_id()
        
        if trace_id:
            print(f"✅ Processing message with trace_id: {trace_id}")
        else:
            print("⚠️  Processing message without trace_id")
        
        # Process message...
        data = json.loads(message.value.decode())
        print(f"   Event: {data}")
    
    # await consumer.consume(trace_aware_handler)
    await consumer.stop()


async def service_integration_example():
    """
    Shows how to integrate with existing analysis-srv patterns.
    """
    
    print("\n=== SERVICE INTEGRATION EXAMPLE ===")
    
    # Simulating analysis-srv Dependencies pattern
    class Dependencies:
        def __init__(self):
            # OLD: self.kafka_producer = KafkaProducer(config)
            # NEW: Drop-in replacement
            self.kafka_producer = TracedKafkaProducer(
                KafkaProducerConfig(
                    bootstrap_servers="localhost:9092",
                    acks="all",
                    compression_type="snappy",
                    enable_trace_injection=True,  # Enable tracing
                )
            )
            
            self.kafka_consumer_config = KafkaConsumerConfig(
                bootstrap_servers="localhost:9092",
                topics=["analytics-tasks"],
                group_id="analytics-group",
                enable_trace_extraction=True,  # Enable tracing
            )
    
    # Simulating ConsumerServer pattern
    class ConsumerServer:
        def __init__(self, deps: Dependencies):
            self.deps = deps
            
        async def start(self):
            # Start producer (same as before)
            await self.deps.kafka_producer.start()
            
            # Create consumer (same interface)
            consumer = TracedKafkaConsumer(self.deps.kafka_consumer_config)
            await consumer.start()
            
            # Message handler with tracing
            async def analytics_handler(message: KafkaMessage) -> None:
                # trace_id automatically available in context
                trace_id = get_trace_id()
                print(f"🔬 Analyzing message with trace_id: {trace_id}")
                
                # Process analytics...
                data = json.loads(message.value.decode())
                
                # Send results with trace propagation
                await self.deps.kafka_producer.send_json(
                    topic="analytics-results",
                    value={"analysis": "completed", "input": data},
                    key=f"result-{data.get('id', 'unknown')}"
                    # trace_id automatically propagated!
                )
            
            # await consumer.consume(analytics_handler)
            await consumer.stop()
            await self.deps.kafka_producer.stop()
    
    # Usage (same as analysis-srv)
    deps = Dependencies()
    server = ConsumerServer(deps)
    # await server.start()
    
    print("✅ Service integration maintains existing patterns with tracing")


if __name__ == "__main__":
    async def main():
        await migration_example()
        await backward_compatibility_example() 
        await advanced_tracing_example()
        await service_integration_example()
        
        print("\n🎉 Migration examples completed!")
        print("\nKey benefits of migrating to smap-shared-libs:")
        print("✅ Automatic trace_id propagation")
        print("✅ Backward compatibility with existing code")
        print("✅ Enhanced error handling and logging")
        print("✅ Cross-language compatibility with Go services")
        print("✅ Comprehensive test coverage")
        print("✅ uv package manager compatibility")
    
    asyncio.run(main())