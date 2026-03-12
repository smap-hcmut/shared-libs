# RabbitMQ Package

The RabbitMQ package provides message queue operations with distributed tracing integration for SMAP services.

## Features

- **Connection Management**: Automatic reconnection with configurable retry
- **Trace Integration**: Automatic trace_id injection in message headers
- **Channel Operations**: Exchange, queue, and binding management
- **Message Publishing**: Publish messages with trace context
- **Message Consuming**: Consume messages with trace extraction
- **Backward Compatibility**: Drop-in replacement for existing RabbitMQ packages
- **Concurrent Safe**: All operations are safe for concurrent use

## Usage

### Basic Usage (Backward Compatible)

```go
import "github.com/smap-hcmut/shared-libs/go/rabbitmq"

// Create connection
conn, err := rabbitmq.NewRabbitMQ("amqp://guest:guest@localhost:5672/", false)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Create channel
ch, err := conn.Channel()
if err != nil {
    log.Fatal(err)
}
defer ch.Close()

// Declare exchange
err = ch.ExchangeDeclare(rabbitmq.ExchangeArgs{
    Name:    "my-exchange",
    Type:    rabbitmq.ExchangeTypeDirect,
    Durable: true,
})

// Declare queue
queue, err := ch.QueueDeclare(rabbitmq.QueueArgs{
    Name:    "my-queue",
    Durable: true,
})

// Bind queue to exchange
err = ch.QueueBind(rabbitmq.QueueBindArgs{
    Queue:      queue.Name,
    Exchange:   "my-exchange",
    RoutingKey: "my-key",
})
```

### Advanced Usage with Trace Integration

```go
import (
    "github.com/smap-hcmut/shared-libs/go/rabbitmq"
    "github.com/smap-hcmut/shared-libs/go/tracing"
)

// Create connection with custom tracer
tracer := tracing.NewTraceContext()
conn, err := rabbitmq.NewRabbitMQWithTracer("amqp://localhost:5672/", false, tracer)

// Create channel with trace context
ch, err := conn.ChannelWithTrace(ctx)

// All operations with trace context
err = ch.ExchangeDeclareWithTrace(ctx, rabbitmq.ExchangeArgs{...})
queue, err := ch.QueueDeclareWithTrace(ctx, rabbitmq.QueueArgs{...})
err = ch.QueueBindWithTrace(ctx, rabbitmq.QueueBindArgs{...})
```

### Publishing Messages

```go
// Basic publishing
err = ch.Publish(ctx, rabbitmq.PublishArgs{
    Exchange:   "my-exchange",
    RoutingKey: "my-key",
    Msg: rabbitmq.Publishing{
        ContentType: rabbitmq.ContentTypeJSON,
        Body:        []byte(`{"message": "hello"}`),
    },
})

// Publishing with trace integration (automatic trace_id injection)
err = ch.PublishWithTrace(ctx, rabbitmq.PublishArgs{
    Exchange:   "my-exchange",
    RoutingKey: "my-key",
    Msg: rabbitmq.Publishing{
        ContentType: rabbitmq.ContentTypeJSON,
        Body:        []byte(`{"message": "hello"}`),
        // trace_id will be automatically added to Headers
    },
})
```

### Consuming Messages

```go
// Basic consuming
deliveries, err := ch.Consume(rabbitmq.ConsumeArgs{
    Queue:    "my-queue",
    Consumer: "my-consumer",
    AutoAck:  false,
})

// Process messages
for delivery := range deliveries {
    // Extract trace_id from headers if available
    if traceID, ok := delivery.Headers["X-Trace-Id"].(string); ok {
        // Use trace_id for request correlation
        ctx := tracer.SetTraceID(context.Background(), traceID)
        // Process message with trace context
    }
    
    // Acknowledge message
    delivery.Ack(false)
}

// Consuming with trace integration (automatic trace extraction)
deliveries, err := ch.ConsumeWithTrace(ctx, rabbitmq.ConsumeArgs{
    Queue:    "my-queue",
    Consumer: "my-consumer",
    AutoAck:  false,
})
```

### Connection Management

```go
// Check connection status
if conn.IsReady() {
    // Connection is active
}

if conn.IsClosed() {
    // Connection is closed and not retrying
}

// Listen for reconnection events
reconnectChan := make(chan bool)
ch.NotifyReconnect(reconnectChan)

go func() {
    for range reconnectChan {
        log.Println("RabbitMQ reconnected, reinitializing...")
        // Reinitialize exchanges, queues, etc.
    }
}()
```

## API Reference

### Connection Interface

#### IRabbitMQ
- `Close()`: Close the connection
- `IsReady() bool`: Check if connection is active
- `IsClosed() bool`: Check if connection is closed
- `Channel() (IChannel, error)`: Create new channel
- `ChannelWithTrace(ctx) (IChannel, error)`: Create channel with trace context

#### Constructor Functions
- `NewRabbitMQ(url, retryWithoutTimeout) (IRabbitMQ, error)`: Create connection
- `NewRabbitMQWithTracer(url, retryWithoutTimeout, tracer) (IRabbitMQ, error)`: Create with custom tracer

### Channel Interface

#### IChannel
- `ExchangeDeclare(args) error`: Declare exchange
- `ExchangeDeclareWithTrace(ctx, args) error`: Declare with trace context
- `QueueDeclare(args) (Queue, error)`: Declare queue
- `QueueDeclareWithTrace(ctx, args) (Queue, error)`: Declare with trace context
- `QueueBind(args) error`: Bind queue to exchange
- `QueueBindWithTrace(ctx, args) error`: Bind with trace context
- `Publish(ctx, args) error`: Publish message
- `PublishWithTrace(ctx, args) error`: Publish with trace injection
- `Consume(args) (<-chan Delivery, error)`: Start consuming
- `ConsumeWithTrace(ctx, args) (<-chan Delivery, error)`: Consume with trace extraction
- `Close() error`: Close the channel
- `NotifyReconnect(chan bool) <-chan bool`: Listen for reconnection events

### Argument Types

#### ExchangeArgs
```go
type ExchangeArgs struct {
    Name       string                 // Exchange name
    Type       string                 // Exchange type (direct, fanout, topic)
    Durable    bool                   // Survive server restart
    AutoDelete bool                   // Delete when unused
    Internal   bool                   // Internal exchange
    NoWait     bool                   // Don't wait for server response
    Args       map[string]interface{} // Additional arguments
}
```

#### QueueArgs
```go
type QueueArgs struct {
    Name       string                 // Queue name
    Durable    bool                   // Survive server restart
    AutoDelete bool                   // Delete when unused
    Exclusive  bool                   // Exclusive to this connection
    NoWait     bool                   // Don't wait for server response
    Args       map[string]interface{} // Additional arguments
}
```

#### PublishArgs
```go
type PublishArgs struct {
    Exchange   string      // Target exchange
    RoutingKey string      // Routing key
    Mandatory  bool        // Return if unroutable
    Immediate  bool        // Return if no consumers
    Msg        Publishing  // Message to publish
}
```

## Constants

### Connection
- `RetryConnectionDelay`: Delay between connection attempts (2s)
- `RetryConnectionTimeout`: Connection timeout (20s)

### Content Types
- `ContentTypePlainText`: "text/plain"
- `ContentTypeJSON`: "application/json"

### Exchange Types
- `ExchangeTypeDirect`: "direct"
- `ExchangeTypeFanout`: "fanout"
- `ExchangeTypeTopic`: "topic"

### Trace Integration
- `TraceIDHeader`: "X-Trace-Id" (header key for trace propagation)

## Migration Guide

### From Local RabbitMQ Package

1. Update imports:
```go
// Before
import "your-service/pkg/rabbitmq"

// After
import "github.com/smap-hcmut/shared-libs/go/rabbitmq"
```

2. No code changes needed for basic usage
3. Optional: Add trace integration for enhanced debugging

### Trace Integration Benefits

- **Message Tracking**: Follow messages across service boundaries
- **Request Correlation**: Link messages to specific request flows
- **Performance Monitoring**: Measure message processing latency
- **Debugging**: Easier troubleshooting with trace context
- **Service Mesh**: Better observability in microservice architecture

## Best Practices

1. **Connection Management**: Use connection pooling for high-throughput applications
2. **Error Handling**: Always check for connection/channel errors
3. **Graceful Shutdown**: Properly close channels and connections
4. **Message Acknowledgment**: Use manual ACK for reliable message processing
5. **Trace Integration**: Use trace-aware methods for better observability
6. **Reconnection Handling**: Listen for reconnection events and reinitialize resources
7. **Queue Durability**: Use durable queues for important messages

## Error Handling

```go
// Connection errors
if err == rabbitmq.ErrConnectionTimeout {
    // Handle connection timeout
}

// Channel errors
if err == rabbitmq.ErrChannelClosed {
    // Handle closed channel
}

// Connection status
if !conn.IsReady() {
    // Connection is not ready, handle accordingly
}
```