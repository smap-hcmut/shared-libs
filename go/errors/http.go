package errors

import (
	"context"
	"fmt"
	"net/http"

	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// NewHTTPError creates a new HTTP error.
func NewHTTPError(code int, message string) *HTTPError {
	return &HTTPError{
		Code:       code,
		Message:    message,
		StatusCode: code,
	}
}

// NewHTTPErrorWithTrace creates a new HTTP error with trace context.
func NewHTTPErrorWithTrace(ctx context.Context, code int, message string) *HTTPError {
	tracer := tracing.NewTraceContext()
	return &HTTPError{
		Code:       code,
		Message:    message,
		StatusCode: code,
		TraceID:    tracer.GetTraceID(ctx),
	}
}

// Error implements the error interface for HTTPError.
func (e *HTTPError) Error() string {
	msg := e.Message
	if e.TraceID != "" {
		msg += fmt.Sprintf(" (trace_id=%s)", e.TraceID)
	}
	return msg
}

// WithTraceID adds trace_id to the HTTP error.
func (e *HTTPError) WithTraceID(traceID string) *HTTPError {
	e.TraceID = traceID
	return e
}

// Predefined HTTP errors with trace support

// NewBadRequestError creates a 400 Bad Request error.
func NewBadRequestError(message string) *HTTPError {
	if message == "" {
		message = "Bad Request"
	}
	return NewHTTPError(http.StatusBadRequest, message)
}

// NewBadRequestErrorWithTrace creates a 400 Bad Request error with trace context.
func NewBadRequestErrorWithTrace(ctx context.Context, message string) *HTTPError {
	if message == "" {
		message = "Bad Request"
	}
	return NewHTTPErrorWithTrace(ctx, http.StatusBadRequest, message)
}

// NewUnauthorizedError creates a 401 Unauthorized error.
func NewUnauthorizedError() *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, "Unauthorized")
}

// NewUnauthorizedErrorWithTrace creates a 401 Unauthorized error with trace context.
func NewUnauthorizedErrorWithTrace(ctx context.Context) *HTTPError {
	return NewHTTPErrorWithTrace(ctx, http.StatusUnauthorized, "Unauthorized")
}

// NewForbiddenError creates a 403 Forbidden error.
func NewForbiddenError() *HTTPError {
	return NewHTTPError(http.StatusForbidden, "Forbidden")
}

// NewForbiddenErrorWithTrace creates a 403 Forbidden error with trace context.
func NewForbiddenErrorWithTrace(ctx context.Context) *HTTPError {
	return NewHTTPErrorWithTrace(ctx, http.StatusForbidden, "Forbidden")
}

// NewNotFoundError creates a 404 Not Found error.
func NewNotFoundError(resource string) *HTTPError {
	message := "Not Found"
	if resource != "" {
		message = fmt.Sprintf("%s: Not Found", resource)
	}
	return NewHTTPError(http.StatusNotFound, message)
}

// NewNotFoundErrorWithTrace creates a 404 Not Found error with trace context.
func NewNotFoundErrorWithTrace(ctx context.Context, resource string) *HTTPError {
	message := "Not Found"
	if resource != "" {
		message = fmt.Sprintf("%s: Not Found", resource)
	}
	return NewHTTPErrorWithTrace(ctx, http.StatusNotFound, message)
}

// NewConflictError creates a 409 Conflict error.
func NewConflictError(message string) *HTTPError {
	if message == "" {
		message = "Conflict"
	}
	return NewHTTPError(http.StatusConflict, message)
}

// NewConflictErrorWithTrace creates a 409 Conflict error with trace context.
func NewConflictErrorWithTrace(ctx context.Context, message string) *HTTPError {
	if message == "" {
		message = "Conflict"
	}
	return NewHTTPErrorWithTrace(ctx, http.StatusConflict, message)
}

// NewUnprocessableEntityError creates a 422 Unprocessable Entity error.
func NewUnprocessableEntityError(message string) *HTTPError {
	if message == "" {
		message = "Unprocessable Entity"
	}
	return NewHTTPError(http.StatusUnprocessableEntity, message)
}

// NewUnprocessableEntityErrorWithTrace creates a 422 Unprocessable Entity error with trace context.
func NewUnprocessableEntityErrorWithTrace(ctx context.Context, message string) *HTTPError {
	if message == "" {
		message = "Unprocessable Entity"
	}
	return NewHTTPErrorWithTrace(ctx, http.StatusUnprocessableEntity, message)
}

// NewInternalServerError creates a 500 Internal Server Error.
func NewInternalServerError(message string) *HTTPError {
	if message == "" {
		message = "Internal Server Error"
	}
	return NewHTTPError(http.StatusInternalServerError, message)
}

// NewInternalServerErrorWithTrace creates a 500 Internal Server Error with trace context.
func NewInternalServerErrorWithTrace(ctx context.Context, message string) *HTTPError {
	if message == "" {
		message = "Internal Server Error"
	}
	return NewHTTPErrorWithTrace(ctx, http.StatusInternalServerError, message)
}
