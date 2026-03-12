package tracing

import (
	"regexp"
	"strings"
)

// UUID v4 validation regex pattern
// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
// Where x is any hexadecimal digit and y is one of 8, 9, A, or B
var uuidv4Regex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// ValidateUUIDv4 validates if a string is a valid UUID v4 format
func ValidateUUIDv4(traceID string) bool {
	if traceID == "" {
		return false
	}

	// Convert to lowercase for validation
	traceID = strings.ToLower(traceID)

	// Check format using regex
	return uuidv4Regex.MatchString(traceID)
}

// IsValidTraceID checks if a trace_id is valid and non-empty
func IsValidTraceID(traceID string) bool {
	return traceID != "" && ValidateUUIDv4(traceID)
}
