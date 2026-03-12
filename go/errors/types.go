package errors

// ValidationError represents a validation error with trace integration.
type ValidationError struct {
	Code     int      `json:"code"`
	Field    string   `json:"field"`
	Messages []string `json:"messages"`
	TraceID  string   `json:"trace_id,omitempty"`
}

// ValidationErrorCollector collects multiple validation errors with trace context.
type ValidationErrorCollector struct {
	errors  []*ValidationError
	traceID string
}

// PermissionError represents a permission error with trace integration.
type PermissionError struct {
	Code     int      `json:"code"`
	Field    string   `json:"field"`
	Messages []string `json:"messages"`
	TraceID  string   `json:"trace_id,omitempty"`
}

// PermissionErrorCollector collects multiple permission errors with trace context.
type PermissionErrorCollector struct {
	errors  []*PermissionError
	traceID string
}

// HTTPError represents an HTTP error with status code, message and trace integration.
type HTTPError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	TraceID    string `json:"trace_id,omitempty"`
}

// BusinessError represents a business logic error with trace integration.
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

// SystemError represents a system-level error with trace integration.
type SystemError struct {
	Component string `json:"component"`
	Operation string `json:"operation"`
	Message   string `json:"message"`
	Cause     error  `json:"cause,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
}
