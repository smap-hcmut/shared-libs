package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// Setup is called at the beginning of a new session, before ConsumeClaim
func (w *tracedConsumerGroupHandlerWrapper) Setup(session sarama.ConsumerGroupSession) error {
	return w.handler.Setup(session)
}

// Cleanup is called at the end of a session, once all ConsumeClaim goroutines have exited
func (w *tracedConsumerGroupHandlerWrapper) Cleanup(session sarama.ConsumerGroupSession) error {
	return w.handler.Cleanup(session)
}

// ConsumeClaim should start a consumer loop of ConsumerGroupClaim's Messages().
func (w *tracedConsumerGroupHandlerWrapper) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE: The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29

	// Create a new claim that provides messages with trace context
	tracedClaim := &tracedConsumerGroupClaim{
		ConsumerGroupClaim: claim,
		tracer:             w.tracer,
		propagator:         w.propagator,
		messagesChan:       make(chan *sarama.ConsumerMessage),
	}

	// Start a goroutine to process messages and inject trace context
	go func() {
		defer close(tracedClaim.messagesChan)
		for {
			select {
			case message := <-claim.Messages():
				if message == nil {
					return
				}

				// Extract trace_id from message headers
				headers := make(map[string]string)
				for _, header := range message.Headers {
					headers[string(header.Key)] = string(header.Value)
				}

				traceID := w.propagator.ExtractKafka(headers)
				if traceID == "" || !w.tracer.ValidateTraceID(traceID) {
					traceID = w.tracer.GenerateTraceID()
				}

				// Create context with trace_id and store it in the message
				ctx := w.tracer.WithTraceID(context.Background(), traceID)

				// Create a new message with trace context
				tracedMessage := &tracedConsumerMessage{
					ConsumerMessage: message,
					ctx:             ctx,
				}

				select {
				case tracedClaim.messagesChan <- tracedMessage.ConsumerMessage:
				case <-session.Context().Done():
					return
				}

			case <-session.Context().Done():
				return
			}
		}
	}()

	// Call the original handler with traced claim
	return w.handler.ConsumeClaim(session, tracedClaim)
}

// tracedConsumerGroupClaim wraps sarama.ConsumerGroupClaim to provide trace context
type tracedConsumerGroupClaim struct {
	sarama.ConsumerGroupClaim
	tracer       tracing.TraceContext
	propagator   tracing.KafkaPropagator
	messagesChan chan *sarama.ConsumerMessage
}

// tracedConsumerMessage wraps sarama.ConsumerMessage with trace context
type tracedConsumerMessage struct {
	*sarama.ConsumerMessage
	ctx context.Context
}

// Messages returns a channel of messages with trace context
func (t *tracedConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage {
	return t.messagesChan
}
