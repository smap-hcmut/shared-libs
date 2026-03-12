"""
Tests for traced Kafka producer and consumer.

Validates trace_id propagation through Kafka messages.
"""

import asyncio
import json
import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from typing import Dict, Any

from ..tracing import TraceContext, KafkaPropagator
from .config import KafkaProducerConfig, KafkaConsumerConfig, KafkaMessage
from .producer import TracedKafkaProducer
from .consumer import TracedKafkaConsumer


class TestTracedKafkaProducer:
    """Test cases for TracedKafkaProducer."""

    @pytest.fixture
    def producer_config(self):
        """Create test producer configuration."""
        return KafkaProducerConfig(
            bootstrap_servers="localhost:9092",
            client_id="test-producer",
            enable_trace_injection=True,
            auto_generate_trace_id=True,
        )

    @pytest.fixture
    def trace_context(self):
        """Create test trace context."""
        return TraceContext()

    @pytest.fixture
    def kafka_propagator(self, trace_context):
        """Create test kafka propagator."""
        return KafkaPropagator(trace_context)

    @pytest.fixture
    def producer(self, producer_config, trace_context, kafka_propagator):
        """Create test producer."""
        return TracedKafkaProducer(producer_config, trace_context, kafka_propagator)

    @patch('smap_shared.kafka.producer.AIOKafkaProducer')
    async def test_producer_start_stop(self, mock_producer_class, producer):
        """Test producer start and stop lifecycle."""
        mock_producer = AsyncMock()
        mock_producer_class.return_value = mock_producer

        # Test start
        await producer.start()
        assert producer.is_running()
        mock_producer.start.assert_called_once()

        # Test stop
        await producer.stop()
        assert not producer.is_running()
        mock_producer.flush.assert_called_once()
        mock_producer.stop.assert_called_once()

    @patch('smap_shared.kafka.producer.AIOKafkaProducer')
    async def test_send_with_trace_injection(self, mock_producer_class, producer, trace_context):
        """Test message sending with trace_id injection."""
        mock_producer = AsyncMock()
        mock_producer_class.return_value = mock_producer

        # Set trace_id in context
        test_trace_id = "550e8400-e29b-41d4-a716-446655440000"
        trace_context.set_trace_id(test_trace_id)

        await producer.start()

        # Send message
        await producer.send(
            topic="test-topic",
            value=b"test message",
            key=b"test-key"
        )

        # Verify send was called with trace_id header
        mock_producer.send.assert_called_once()
        call_args = mock_producer.send.call_args
        
        assert call_args[1]["topic"] == "test-topic"
        assert call_args[1]["value"] == b"test message"
        assert call_args[1]["key"] == b"test-key"
        
        # Check headers contain trace_id
        headers = call_args[1]["headers"]
        assert headers is not None
        header_dict = dict(headers)
        assert b"X-Trace-Id" in header_dict or "X-Trace-Id" in header_dict
        
        # Extract trace_id from headers
        trace_header_value = None
        for key, value in headers:
            if key == "X-Trace-Id":
                trace_header_value = value.decode('utf-8') if isinstance(value, bytes) else value
                break
        
        assert trace_header_value == test_trace_id

    @patch('smap_shared.kafka.producer.AIOKafkaProducer')
    async def test_send_json_with_trace(self, mock_producer_class, producer, trace_context):
        """Test JSON message sending with trace_id injection."""
        mock_producer = AsyncMock()
        mock_producer_class.return_value = mock_producer

        # Set trace_id in context
        test_trace_id = "550e8400-e29b-41d4-a716-446655440000"
        trace_context.set_trace_id(test_trace_id)

        await producer.start()

        # Send JSON message
        test_data = {"message": "hello", "timestamp": 1234567890}
        await producer.send_json(
            topic="test-topic",
            value=test_data,
            key="test-key"
        )

        # Verify send was called
        mock_producer.send.assert_called_once()
        call_args = mock_producer.send.call_args
        
        # Check JSON serialization
        expected_json = json.dumps(test_data, ensure_ascii=False).encode("utf-8")
        assert call_args[1]["value"] == expected_json
        assert call_args[1]["key"] == b"test-key"

    @patch('smap_shared.kafka.producer.AIOKafkaProducer')
    async def test_auto_generate_trace_id(self, mock_producer_class, producer, trace_context):
        """Test automatic trace_id generation when missing."""
        mock_producer = AsyncMock()
        mock_producer_class.return_value = mock_producer

        # Ensure no trace_id in context
        trace_context.clear_trace_id()
        assert trace_context.get_trace_id() is None

        await producer.start()

        # Send message (should auto-generate trace_id)
        await producer.send(
            topic="test-topic",
            value=b"test message"
        )

        # Verify trace_id was generated and set
        generated_trace_id = trace_context.get_trace_id()
        assert generated_trace_id is not None
        assert trace_context.validate_trace_id(generated_trace_id)

        # Verify send was called with generated trace_id
        mock_producer.send.assert_called_once()
        call_args = mock_producer.send.call_args
        headers = call_args[1]["headers"]
        
        # Find trace_id in headers
        trace_header_value = None
        for key, value in headers:
            if key == "X-Trace-Id":
                trace_header_value = value.decode('utf-8') if isinstance(value, bytes) else value
                break
        
        assert trace_header_value == generated_trace_id

    def test_producer_config_validation(self):
        """Test producer configuration validation."""
        # Test empty bootstrap_servers
        with pytest.raises(ValueError, match="bootstrap_servers cannot be empty"):
            KafkaProducerConfig(bootstrap_servers="")

        # Test invalid acks
        with pytest.raises(ValueError, match="acks must be"):
            KafkaProducerConfig(bootstrap_servers="localhost:9092", acks="invalid")

        # Test invalid compression_type
        with pytest.raises(ValueError, match="compression_type must be"):
            KafkaProducerConfig(
                bootstrap_servers="localhost:9092", 
                compression_type="invalid"
            )


class TestTracedKafkaConsumer:
    """Test cases for TracedKafkaConsumer."""

    @pytest.fixture
    def consumer_config(self):
        """Create test consumer configuration."""
        return KafkaConsumerConfig(
            bootstrap_servers="localhost:9092",
            topics=["test-topic"],
            group_id="test-group",
            client_id="test-consumer",
            enable_trace_extraction=True,
            auto_generate_trace_id=True,
        )

    @pytest.fixture
    def trace_context(self):
        """Create test trace context."""
        return TraceContext()

    @pytest.fixture
    def kafka_propagator(self, trace_context):
        """Create test kafka propagator."""
        return KafkaPropagator(trace_context)

    @pytest.fixture
    def consumer(self, consumer_config, trace_context, kafka_propagator):
        """Create test consumer."""
        return TracedKafkaConsumer(consumer_config, trace_context, kafka_propagator)

    @patch('smap_shared.kafka.consumer.AIOKafkaConsumer')
    async def test_consumer_start_stop(self, mock_consumer_class, consumer):
        """Test consumer start and stop lifecycle."""
        mock_consumer = AsyncMock()
        mock_consumer_class.return_value = mock_consumer

        # Test start
        await consumer.start()
        assert consumer.is_running()
        mock_consumer.start.assert_called_once()

        # Test stop
        await consumer.stop()
        assert not consumer.is_running()
        mock_consumer.stop.assert_called_once()

    @patch('smap_shared.kafka.consumer.AIOKafkaConsumer')
    async def test_consume_with_trace_extraction(self, mock_consumer_class, consumer, trace_context):
        """Test message consumption with trace_id extraction."""
        mock_consumer = AsyncMock()
        mock_consumer_class.return_value = mock_consumer

        # Create mock message with trace_id header
        test_trace_id = "550e8400-e29b-41d4-a716-446655440000"
        mock_message = MagicMock()
        mock_message.topic = "test-topic"
        mock_message.partition = 0
        mock_message.offset = 123
        mock_message.value = b"test message"
        mock_message.key = b"test-key"
        mock_message.timestamp = 1234567890
        mock_message.headers = [("X-Trace-Id", test_trace_id.encode('utf-8'))]

        # Mock async iteration
        async def mock_iter(self):
            yield mock_message

        mock_consumer.__aiter__ = mock_iter

        await consumer.start()

        # Create message handler
        received_messages = []
        async def message_handler(msg: KafkaMessage):
            received_messages.append(msg)
            # Verify trace_id is set in context
            current_trace_id = trace_context.get_trace_id()
            assert current_trace_id == test_trace_id

        # Start consuming (will process one message then stop)
        consume_task = asyncio.create_task(consumer.consume(message_handler))
        
        # Give it a moment to process the message
        await asyncio.sleep(0.1)
        consume_task.cancel()
        
        try:
            await consume_task
        except asyncio.CancelledError:
            pass

        # Verify message was processed
        assert len(received_messages) == 1
        msg = received_messages[0]
        assert msg.topic == "test-topic"
        assert msg.partition == 0
        assert msg.offset == 123
        assert msg.value == b"test message"
        assert msg.key == b"test-key"
        assert msg.trace_id == test_trace_id

    @patch('smap_shared.kafka.consumer.AIOKafkaConsumer')
    async def test_consume_auto_generate_trace_id(self, mock_consumer_class, consumer, trace_context):
        """Test automatic trace_id generation when missing from message."""
        mock_consumer = AsyncMock()
        mock_consumer_class.return_value = mock_consumer

        # Create mock message without trace_id header
        mock_message = MagicMock()
        mock_message.topic = "test-topic"
        mock_message.partition = 0
        mock_message.offset = 123
        mock_message.value = b"test message"
        mock_message.key = b"test-key"
        mock_message.timestamp = 1234567890
        mock_message.headers = []  # No headers

        # Mock async iteration
        async def mock_iter(self):
            yield mock_message

        mock_consumer.__aiter__ = mock_iter

        await consumer.start()

        # Create message handler
        received_messages = []
        async def message_handler(msg: KafkaMessage):
            received_messages.append(msg)
            # Verify trace_id was generated and set in context
            current_trace_id = trace_context.get_trace_id()
            assert current_trace_id is not None
            assert trace_context.validate_trace_id(current_trace_id)

        # Start consuming (will process one message then stop)
        consume_task = asyncio.create_task(consumer.consume(message_handler))
        
        # Give it a moment to process the message
        await asyncio.sleep(0.1)
        consume_task.cancel()
        
        try:
            await consume_task
        except asyncio.CancelledError:
            pass

        # Verify message was processed with generated trace_id
        assert len(received_messages) == 1
        msg = received_messages[0]
        assert msg.trace_id is not None
        assert trace_context.validate_trace_id(msg.trace_id)

    def test_consumer_config_validation(self):
        """Test consumer configuration validation."""
        # Test empty bootstrap_servers
        with pytest.raises(ValueError, match="bootstrap_servers cannot be empty"):
            KafkaConsumerConfig(
                bootstrap_servers="",
                topics=["test"],
                group_id="test-group"
            )

        # Test empty topics
        with pytest.raises(ValueError, match="topics cannot be empty"):
            KafkaConsumerConfig(
                bootstrap_servers="localhost:9092",
                topics=[],
                group_id="test-group"
            )

        # Test empty group_id
        with pytest.raises(ValueError, match="group_id cannot be empty"):
            KafkaConsumerConfig(
                bootstrap_servers="localhost:9092",
                topics=["test"],
                group_id=""
            )


class TestKafkaMessage:
    """Test cases for KafkaMessage data model."""

    def test_kafka_message_creation(self):
        """Test KafkaMessage creation and attributes."""
        msg = KafkaMessage(
            topic="test-topic",
            partition=0,
            offset=123,
            value=b"test message",
            key=b"test-key",
            timestamp=1234567890,
            headers={"Content-Type": b"application/json"},
            trace_id="550e8400-e29b-41d4-a716-446655440000"
        )

        assert msg.topic == "test-topic"
        assert msg.partition == 0
        assert msg.offset == 123
        assert msg.value == b"test message"
        assert msg.key == b"test-key"
        assert msg.timestamp == 1234567890
        assert msg.headers == {"Content-Type": b"application/json"}
        assert msg.trace_id == "550e8400-e29b-41d4-a716-446655440000"

    def test_kafka_message_defaults(self):
        """Test KafkaMessage with default values."""
        msg = KafkaMessage(
            topic="test-topic",
            partition=0,
            offset=123,
            value=b"test message"
        )

        assert msg.key is None
        assert msg.timestamp is None
        assert msg.headers == {}
        assert msg.trace_id is None


if __name__ == "__main__":
    pytest.main([__file__])