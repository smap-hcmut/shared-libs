package cron

import "errors"

var (
	// ErrUnsupportedImplementation is returned when an unknown implementation is requested
	ErrUnsupportedImplementation = errors.New("unsupported cron implementation")

	// ErrInvalidCronTime is returned when the cron time format is invalid
	ErrInvalidCronTime = errors.New("invalid cron time format")

	// ErrJobHandlerRequired is returned when a job handler is not provided
	ErrJobHandlerRequired = errors.New("job handler is required")
)
