# Python Kafka Client with Distributed Tracing

Enhanced Kafka producer and consumer with automatic trace_id propagation for distributed tracing.

**✅ Migrated from analysis-srv**: This implementation incorporates the best features from the analysis-srv Kafka package with enhanced tracing capabilities. Provides full backward compatibility while adding automatic trace propagation.

## Features

- **Automatic Trace Propagation**: Seamlessly injects and extracts trace_id from Kafka message headers
- **Backward Compatibility**: Drop-in replacement for existing Kafka clients (including analysis-srv)
- **Cross-Language Support**: Compatible with Go services using the same trace header format
- **Robust Error Handling**: Graceful degradation when trace operations fail
- **Async/Await Support**: Built on aiokafka for high-performance async operations
- **Enhanced from analysis-srv**: Improved error handling, validation, and tracing features

## Installation

The Kafka client requires `aiokafka`:

```bash
pip install aiokafka
```

## Quick Start

### Producer with Trace Injection

```python
import asyncio
from smap_shared.kafka import TracedKafkaProducer, KafkaProducerConfig
from smap_shared.tracing import set_trace_id

async def main():
    # Configure producer
    config = KafkaProducerConfig(
        bootstrap_servers="localhost:9092",
        client_id="my-producer",
        enable_trace_injection=True,
        auto_generate_trace_id=True
    )
    
    # Create producer
    producer = TracedKafkaProducer(config)
    
    try:
        await producer.start()
        
        # Set trace_id in context (optional - will auto-generate if missing)
        set_trace_id("550e8400-e29b-41d4-a716-446655440000")
        
        # Send message with automatic trace_id injection
        await producer.send_json(
            topic="user-events",
            value={"user_id": 123, "action": "login"},
            key="user-123"
        )
        
        print("Message sent with trace_id!")
        
    finally:
        await producer.stop()

asyncio.run(main())
```

### Consumer with Trace Extraction

```python
import asyncio
from smap_shared.kafka import TracedKafkaConsumer, KafkaConsumerConfig, KafkaMessage
from smap_shared.tracing import get_trace_id

async def message_handler(message: KafkaMessage):
    """Process incoming message with trace context."""
    # Trace_id is automatically extracted and set in context
    current_trace_id = get_trace_id()
    
    print(f"Processing message with trace_id: {current_trace_id}")
    print(f"Message: {message.value.decode('utf-8')}")
    print(f"Topic: {message.topic}, Partition: {message.partition}")

async def main():
    # Configure consumer
    config = KafkaConsumerConfig(
        bootstrap_servers="localhost:9092",
        topics=["user-events"],
        group_id="my-consumer-group",
        client_id="my-consumer",
        enable_trace_extraction=True,
        auto_generate_trace_id=True
    )
    
    # Create consumer
    consumer = TracedKafkaConsumer(config)
    
    try:
        await consumer.start()
        
        # Start consuming with automatic trace_id extraction
        await consumer.consume(message_handler)
        
    finally:
        await consumer.stop()

asyncio.run(main())
```

## Configuration

### Producer Configuration

```python
from smap_shared.kafka import KafkaProducerConfig

config = KafkaProducerConfig(
    bootstrap_servers="localhost:9092",           # Required: Kafka brokers
    acks="all",                                   # Acknowledgment level
    compression_type=None,                        # Compression: gzip, snappy, lz4, zstd
    max_batch_size=16384,                        # Max batch size in bytes
    linger_ms=0,                                 # Batch linger time
    client_id="my-producer",                     # Client identifier
    enable_idempotence=True,                     # Idempotent producer
    enable_trace_injection=True,                 # Enable trace_id injection
    auto_generate_trace_id=True                  # Auto-generate if missing
)
```

### Consumer Configuration

```python
from smap_shared.kafka import KafkaConsumerConfig

config = KafkaConsumerConfig(
    bootstrap_servers="localhost:9092",           # Required: Kafka brokers
    topics=["topic1", "topic2"],                 # Required: Topics to subscribe
    group_id="my-group",                         # Required: Consumer group
    auto_offset_reset="earliest",                # Offset reset: earliest, latest
    enable_auto_commit=True,                     # Auto-commit offsets
    max_poll_records=500,                        # Max records per poll
    session_timeout_ms=10000,                    # Session timeout
    client_id="my-consumer",                     # Client identifier
    enable_trace_extraction=True,                # Enable trace_id extraction
    auto_generate_trace_id=True                  # Auto-generate if missing
)
```

## Advanced Usage

### Manual Trace Management

```python
from smap_shared.kafka import TracedKafkaProducer
from smap_shared.tracing import TraceContext, KafkaPropagator

# Create custom trace context
trace_context = TraceContext()
kafka_propagator = KafkaPropagator(trace_context)

# Create producer with custom tracing
producer = TracedKafkaProducer(
    config=config,
    trace_context=trace_context,
    kafka_propagator=kafka_propagator
)

# Manually set trace_id
trace_context.set_trace_id("custom-trace-id")

# Send message
await producer.send(topic="events", value=b"message")
```

### Batch Message Sending

```python
# Prepare batch messages
messages = [
    {
        "topic": "events",
        "value": b"message 1",
        "key": b"key1"
    },
    {
        "topic": "events", 
        "value": b"message 2",
        "key": b"key2"
    }
]

# Send batch with trace_id injection
await producer.send_batch(messages)
```

### Custom Headers

```python
# Send message with custom headers
custom_headers = {
    "Content-Type": b"application/json",
    "Source": b"my-service"
}

await producer.send(
    topic="events",
    value=b"message",
    headers=custom_headers  # Trace_id will be added automatically
)
```

## Migration from Existing Kafka Clients

### From aiokafka

```python
# Before: Direct aiokafka usage
from aiokafka import AIOKafkaProducer

producer = AIOKafkaProducer(bootstrap_servers="localhost:9092")
await producer.start()
await producer.send("topic", b"message")
await producer.stop()

# After: Traced Kafka producer
from smap_shared.kafka import TracedKafkaProducer, KafkaProducerConfig

config = KafkaProducerConfig(bootstrap_servers="localhost:9092")
producer = TracedKafkaProducer(config)
await producer.start()
await producer.send("topic", b"message")  # Automatic trace_id injection
await producer.stop()
```

### From analysis-srv Kafka Package

```python
# Before: analysis-srv/pkg/kafka
from pkg.kafka import KafkaProducer, KafkaProducerConfig

config = KafkaProducerConfig(bootstrap_servers="localhost:9092")
producer = KafkaProducer(config)
await producer.start()
await producer.send_json("topic", {"data": "value"})
await producer.stop()

# After: Shared library with tracing
from smap_shared.kafka import TracedKafkaProducer, KafkaProducerConfig

config = KafkaProducerConfig(bootstrap_servers="localhost:9092")
producer = TracedKafkaProducer(config)
await producer.start()
await producer.send_json("topic", {"data": "value"})  # Automatic trace_id injection
await producer.stop()
```

## Trace Header Format

The client uses the standard `X-Trace-Id` header for cross-service compatibility:

```
X-Trace-Id: 550e8400-e29b-41d4-a716-446655440000
```

This format is compatible with:
- Go services using the shared tracing library
- HTTP services using the same header standard
- External monitoring and tracing systems

## Error Handling

The traced Kafka client provides graceful error handling:

```python
try:
    await producer.send("topic", b"message")
except KafkaProducerError as e:
    print(f"Kafka producer error: {e}")
except Exception as e:
    print(f"Unexpected error: {e}")
```

### Trace Operation Failures

If trace operations fail, the client continues normal operation:

- **Trace injection failure**: Message is sent without trace_id header
- **Trace extraction failure**: Message is processed without trace context
- **Invalid trace_id**: New trace_id is generated automatically

## Performance Considerations

- **Minimal Overhead**: Trace operations add <1ms latency per message
- **Memory Efficient**: Context variables use minimal memory
- **Async Optimized**: Built on aiokafka for high throughput
- **Batch Friendly**: Trace injection works efficiently with batch operations

## Testing

Run the test suite:

```bash
cd smap-shared-libs/python
python -m pytest smap_shared/kafka/test_kafka.py -v
```

## Compatibility

- **Python**: 3.8+
- **aiokafka**: 0.7+
- **Kafka**: 2.0+
- **Cross-language**: Compatible with Go services using smap-shared-libs/go

## Examples

See the `examples/` directory for complete usage examples:

- `kafka_producer_example.py`: Basic producer usage
- `kafka_consumer_example.py`: Basic consumer usage  
- `kafka_batch_example.py`: Batch processing
- `kafka_migration_example.py`: Migration from existing clients