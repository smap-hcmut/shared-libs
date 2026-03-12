package kafka

import (
	"context"
	"testing"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/smap/shared-libs/go/tracing"
)

// TestTracePropagationIntegration tests end-to-end trace propagation
func TestTracePropagationIntegration(t *testing.T) {
	// Test trace injection in producer
	tracer := tracing.NewTraceContext()
	propagator := tracing.NewKafkaPropagator(tracer)

	// Generate a test trace ID
	testTraceID := uuid.New().String()
	ctx := tracer.WithTraceID(context.Background(), testTraceID)

	// Test header injection
	headers := make(map[string]string)
	propagator.InjectKafka(ctx, headers)

	if headers["X-Trace-Id"] != testTraceID {
		t.Errorf("Expected trace_id %s in headers, got %s", testTraceID, headers["X-Trace-Id"])
	}

	// Test header extraction
	extractedTraceID := propagator.ExtractKafka(headers)
	if extractedTraceID != testTraceID {
		t.Errorf("Expected extracted trace_id %s, got %s", testTraceID, extractedTraceID)
	}
}

// TestTracedProducerHeaderInjection tests that traced producer injects headers correctly
func TestTracedProducerHeaderInjection(t *testing.T) {
	// Create a mock sarama producer to capture messages
	mockProducer := &mockSyncProducer{
		messages: make([]*sarama.ProducerMessage, 0),
	}

	tracer := tracing.NewTraceContext()
	propagator := tracing.NewKafkaPropagator(tracer)

	producer := &tracedProducerImpl{
		producer:   mockProducer,
		topic:      "test-topic",
		tracer:     tracer,
		propagator: propagator,
	}

	// Create context with trace ID
	testTraceID := uuid.New().String()
	ctx := tracer.WithTraceID(context.Background(), testTraceID)

	// Publish message
	err := producer.PublishWithContext(ctx, []byte("key"), []byte("value"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify message was captured
	if len(mockProducer.messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(mockProducer.messages))
	}

	message := mockProducer.messages[0]

	// Verify headers contain trace ID
	var foundTraceID string
	for _, header := range message.Headers {
		if string(header.Key) == "X-Trace-Id" {
			foundTraceID = string(header.Value)
			break
		}
	}

	if foundTraceID != testTraceID {
		t.Errorf("Expected trace_id %s in message headers, got %s", testTraceID, foundTraceID)
	}
}

// TestTracedConsumerHeaderExtraction tests that traced consumer extracts headers correctly
func TestTracedConsumerHeaderExtraction(t *testing.T) {
	tracer := tracing.NewTraceContext()
	propagator := tracing.NewKafkaPropagator(tracer)

	// Create test message with trace ID header
	testTraceID := uuid.New().String()
	message := &sarama.ConsumerMessage{
		Headers: []*sarama.RecordHeader{
			{
				Key:   []byte("X-Trace-Id"),
				Value: []byte(testTraceID),
			},
		},
		Value: []byte("test message"),
	}

	// Test header extraction
	headers := make(map[string]string)
	for _, header := range message.Headers {
		headers[string(header.Key)] = string(header.Value)
	}

	extractedTraceID := propagator.ExtractKafka(headers)
	if extractedTraceID != testTraceID {
		t.Errorf("Expected extracted trace_id %s, got %s", testTraceID, extractedTraceID)
	}
}

// Mock sarama.SyncProducer for testing
type mockSyncProducer struct {
	messages []*sarama.ProducerMessage
}

func (m *mockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	m.messages = append(m.messages, msg)
	return 0, 0, nil
}

func (m *mockSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	m.messages = append(m.messages, msgs...)
	return nil
}

func (m *mockSyncProducer) Close() error {
	return nil
}

func (m *mockSyncProducer) GetMetadata() (*sarama.MetadataResponse, error) {
	return nil, nil
}

func (m *mockSyncProducer) IsTransactional() bool {
	return false
}

func (m *mockSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	return 0
}

func (m *mockSyncProducer) BeginTxn() error {
	return nil
}

func (m *mockSyncProducer) CommitTxn() error {
	return nil
}

func (m *mockSyncProducer) AbortTxn() error {
	return nil
}

func (m *mockSyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	return nil
}

func (m *mockSyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	return nil
}
