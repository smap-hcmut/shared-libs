package response

import (
	"encoding/json"
	"time"
)

// ErrorMapping maps errors to HTTP status codes for ErrorWithMap functionality
type ErrorMapping map[error]int

// Date is a date that marshals as DateFormat
type Date time.Time

// MarshalJSON implements json.Marshaler for Date
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Local().Format(DateFormat))
}

// DateTime is a datetime that marshals as DateTimeFormat
type DateTime time.Time

// MarshalJSON implements json.Marshaler for DateTime
func (d DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Local().Format(DateTimeFormat))
}
