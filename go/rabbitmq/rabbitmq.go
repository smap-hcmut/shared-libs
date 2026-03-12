package rabbitmq

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// --- connectionImpl: connection management with trace integration ---

func (c *connectionImpl) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	c.isRetrying = false
}

func (c *connectionImpl) IsReady() bool {
	return c.conn != nil && !c.conn.IsClosed()
}

func (c *connectionImpl) IsClosed() bool {
	return !c.IsReady() && !c.isRetrying
}

func (c *connectionImpl) Channel() (IChannel, error) {
	return c.ChannelWithTrace(context.Background())
}

func (c *connectionImpl) ChannelWithTrace(ctx context.Context) (IChannel, error) {
	if traceID := c.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Creating RabbitMQ channel", "trace_id", traceID)
	}

	ch, err := c.channel()
	if err != nil {
		return nil, err
	}
	chImpl := &channelImpl{
		conn:   c,
		ch:     ch,
		tracer: c.tracer,
	}
	chImpl.listenNotifyReconnect()
	return chImpl, nil
}

func (c *connectionImpl) dial(url string, connChan chan *amqp.Connection, cancelChan chan bool) {
	count := 0
	for {
		select {
		case <-cancelChan:
			return
		default:
			log.Printf("Connecting to RabbitMQ, attempt: %d ...\n", count+1)
			conn, err := amqp.Dial(url)
			if err != nil {
				log.Printf("Connection to RabbitMQ failed: %v\n", err)
				time.Sleep(RetryConnectionDelay)
				count++
				continue
			}
			log.Println("Connected to RabbitMQ!")
			connChan <- conn
			return
		}
	}
}

func (c *connectionImpl) connectWithoutTimeout() error {
	connChan := make(chan *amqp.Connection)
	go c.dial(c.url, connChan, make(chan bool))
	conn := <-connChan
	c.conn = conn
	c.listenNotifyClose()
	return nil
}

func (c *connectionImpl) connect() error {
	connChan := make(chan *amqp.Connection)
	cancelChan := make(chan bool)
	go c.dial(c.url, connChan, cancelChan)
	select {
	case conn := <-connChan:
		c.conn = conn
		c.listenNotifyClose()
		return nil
	case <-time.After(RetryConnectionTimeout):
		cancelChan <- true
		return ErrConnectionTimeout
	}
}
func (c *connectionImpl) listenNotifyClose() {
	fn := c.connect
	if c.retryWithoutTimeout {
		fn = c.connectWithoutTimeout
	}
	notifyClose := make(chan *amqp.Error)
	c.conn.NotifyClose(notifyClose)
	go func() {
		for err := range notifyClose {
			if err != nil {
				c.conn = nil
				c.isRetrying = true
				log.Printf("Connection to RabbitMQ closed: %v\n", err)
				if err := fn(); err != nil {
					log.Printf("Connection to RabbitMQ failed: %v\n", err)
				}
				for _, reconnect := range c.reconnects {
					reconnect <- true
				}
				c.isRetrying = false
				return
			}
		}
	}()
}

func (c *connectionImpl) channel() (*amqp.Channel, error) {
	return c.conn.Channel()
}

func (c *connectionImpl) notifyReconnect(receiver chan bool) <-chan bool {
	c.reconnects = append(c.reconnects, receiver)
	return receiver
}

// --- channelImpl: channel operations with trace integration ---

func (ch *channelImpl) ExchangeDeclare(exc ExchangeArgs) error {
	return ch.ExchangeDeclareWithTrace(context.Background(), exc)
}

func (ch *channelImpl) ExchangeDeclareWithTrace(ctx context.Context, exc ExchangeArgs) error {
	if traceID := ch.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Declaring exchange", "trace_id", traceID, "exchange", exc.Name)
	}
	return ch.ch.ExchangeDeclare(exc.spread())
}

func (ch *channelImpl) QueueDeclare(queue QueueArgs) (amqp.Queue, error) {
	return ch.QueueDeclareWithTrace(context.Background(), queue)
}

func (ch *channelImpl) QueueDeclareWithTrace(ctx context.Context, queue QueueArgs) (amqp.Queue, error) {
	if traceID := ch.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Declaring queue", "trace_id", traceID, "queue", queue.Name)
	}
	return ch.ch.QueueDeclare(queue.spread())
}

func (ch *channelImpl) QueueBind(queueBind QueueBindArgs) error {
	return ch.QueueBindWithTrace(context.Background(), queueBind)
}

func (ch *channelImpl) QueueBindWithTrace(ctx context.Context, queueBind QueueBindArgs) error {
	if traceID := ch.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Binding queue", "trace_id", traceID, "queue", queueBind.Queue, "exchange", queueBind.Exchange)
	}
	return ch.ch.QueueBind(queueBind.spread())
}
func (ch *channelImpl) Publish(ctx context.Context, publish PublishArgs) error {
	return ch.PublishWithTrace(ctx, publish)
}

func (ch *channelImpl) PublishWithTrace(ctx context.Context, publish PublishArgs) error {
	// Inject trace_id into message headers
	if traceID := ch.tracer.GetTraceID(ctx); traceID != "" {
		if publish.Msg.Headers == nil {
			publish.Msg.Headers = make(amqp.Table)
		}
		publish.Msg.Headers[TraceIDHeader] = traceID
		// Could add trace logging here: log.Info("Publishing message", "trace_id", traceID, "exchange", publish.Exchange, "routing_key", publish.RoutingKey)
	}

	return ch.ch.PublishWithContext(publish.spread(ctx))
}

func (ch *channelImpl) Consume(consume ConsumeArgs) (<-chan amqp.Delivery, error) {
	return ch.ConsumeWithTrace(context.Background(), consume)
}

func (ch *channelImpl) ConsumeWithTrace(ctx context.Context, consume ConsumeArgs) (<-chan amqp.Delivery, error) {
	if traceID := ch.tracer.GetTraceID(ctx); traceID != "" {
		// Could add trace logging here: log.Info("Starting consumer", "trace_id", traceID, "queue", consume.Queue, "consumer", consume.Consumer)
	}

	deliveries, err := ch.ch.Consume(consume.spread())
	if err != nil {
		return nil, err
	}

	// Wrap deliveries to extract trace_id from headers
	tracedDeliveries := make(chan amqp.Delivery)
	go func() {
		defer close(tracedDeliveries)
		for delivery := range deliveries {
			// Extract trace_id from message headers if available
			if delivery.Headers != nil {
				if traceID, ok := delivery.Headers[TraceIDHeader].(string); ok && traceID != "" {
					// Could add trace logging here: log.Info("Received message", "trace_id", traceID, "queue", consume.Queue)
				}
			}
			tracedDeliveries <- delivery
		}
	}()

	return tracedDeliveries, nil
}

func (ch *channelImpl) Close() error {
	return ch.ch.Close()
}

func (ch *channelImpl) NotifyReconnect(receiver chan bool) <-chan bool {
	ch.reconnects = append(ch.reconnects, receiver)
	return receiver
}

func (ch *channelImpl) listenNotifyReconnect() {
	reconnNoti := make(chan bool)
	ch.conn.notifyReconnect(reconnNoti)
	go func() {
		for {
			<-reconnNoti
			log.Println("Retry creating RabbitMQ channel...")
			channel, err := ch.conn.channel()
			if err != nil {
				log.Printf("RabbitMQ channel failed: %v\n", err)
				continue
			}
			_ = ch.ch.Close()
			ch.ch = channel
		}
	}()
}
