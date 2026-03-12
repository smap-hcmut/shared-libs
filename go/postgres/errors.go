package postgres

import "errors"

var (
	// Configuration errors
	ErrHostRequired   = errors.New("postgres: host is required")
	ErrInvalidPort    = errors.New("postgres: invalid port")
	ErrUserRequired   = errors.New("postgres: user is required")
	ErrDBNameRequired = errors.New("postgres: database name is required")

	// UUID validation errors (migrated from existing implementations)
	ErrInvalidObjectIDs = errors.New("postgres: invalid object ids")
	ErrInvalidUUID      = errors.New("postgres: invalid UUID format")
)
