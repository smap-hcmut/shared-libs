package response

import (
	"encoding/json"
	"time"

	"github.com/smap-hcmut/shared-libs/go/errors"
)

// Resp is the standard JSON response body with trace integration.
type Resp struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Errors    any    `json:"errors,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
}

// ErrorMapping maps errors to HTTPError for ErrorWithMap.
type ErrorMapping map[error]*errors.HTTPError

// Date is a date that marshals as DateFormat.
type Date time.Time

// MarshalJSON implements json.Marshaler for Date.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Local().Format(DateFormat))
}

// DateTime is a datetime that marshals as DateTimeFormat.
type DateTime time.Time

// MarshalJSON implements json.Marshaler for DateTime.
func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Local().Format(DateTimeFormat))
}

// PaginatedResp is a paginated response with metadata.
type PaginatedResp struct {
	ErrorCode  int         `json:"error_code"`
	Message    string      `json:"message"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	TraceID    string      `json:"trace_id,omitempty"`
}

// Pagination contains pagination metadata.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// HealthResp is a health check response.
type HealthResp struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
	TraceID   string            `json:"trace_id,omitempty"`
}
