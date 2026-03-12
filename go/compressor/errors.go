package compressor

import "errors"

var (
	// ErrUnsupportedImplementation is returned when an unknown implementation is requested
	ErrUnsupportedImplementation = errors.New("unsupported compressor implementation")
	
	// ErrInvalidConfig is returned when the provided configuration is invalid
	ErrInvalidConfig = errors.New("invalid compressor configuration")
)
