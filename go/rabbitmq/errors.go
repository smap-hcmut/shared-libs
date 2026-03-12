package rabbitmq

import "errors"

var (
	// ErrConnectionTimeout is returned when connection to RabbitMQ times out
	ErrConnectionTimeout = errors.New("connection timeout")
	// ErrChannelClosed is returned when trying to use a closed channel
	ErrChannelClosed = errors.New("channel is closed")
	// ErrConnectionClosed is returned when trying to use a closed connection
	ErrConnectionClosed = errors.New("connection is closed")
)
