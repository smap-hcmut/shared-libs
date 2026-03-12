package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// IRabbitMQ is the RabbitMQ interface with trace integration
// Implementations are safe for concurrent use
type IRabbitMQ interface {
	Close()
	IsReady() bool
	IsClosed() bool
	Channel() (IChannel, error)
	ChannelWithTrace(ctx context.Context) (IChannel, error)
}

// IChannel is the RabbitMQ channel interface with trace integration
// Implementations are safe for concurrent use
type IChannel interface {
	ExchangeDeclare(exc ExchangeArgs) error
	ExchangeDeclareWithTrace(ctx context.Context, exc ExchangeArgs) error
	QueueDeclare(queue QueueArgs) (amqp.Queue, error)
	QueueDeclareWithTrace(ctx context.Context, queue QueueArgs) (amqp.Queue, error)
	QueueBind(queueBind QueueBindArgs) error
	QueueBindWithTrace(ctx context.Context, queueBind QueueBindArgs) error
	Publish(ctx context.Context, publish PublishArgs) error
	PublishWithTrace(ctx context.Context, publish PublishArgs) error
	Consume(consume ConsumeArgs) (<-chan amqp.Delivery, error)
	ConsumeWithTrace(ctx context.Context, consume ConsumeArgs) (<-chan amqp.Delivery, error)
	Close() error
	NotifyReconnect(receiver chan bool) <-chan bool
}

// NewRabbitMQ creates a new RabbitMQ connection with trace integration
func NewRabbitMQ(url string, retryWithoutTimeout bool) (IRabbitMQ, error) {
	return NewRabbitMQWithTracer(url, retryWithoutTimeout, nil)
}

// NewRabbitMQWithTracer creates a new RabbitMQ connection with custom tracer
func NewRabbitMQWithTracer(url string, retryWithoutTimeout bool, tracer tracing.TraceContext) (IRabbitMQ, error) {
	if tracer == nil {
		tracer = tracing.NewTraceContext()
	}

	conn := &connectionImpl{
		url:                 url,
		retryWithoutTimeout: retryWithoutTimeout,
		tracer:              tracer,
	}
	if err := conn.connect(); err != nil {
		return nil, err
	}
	return conn, nil
}
