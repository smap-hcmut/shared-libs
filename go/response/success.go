package response

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smap-hcmut/shared-libs/go/tracing"
)

// NewOKResp creates a new OK response with the given data.
func NewOKResp(data any) Resp {
	return Resp{
		ErrorCode: 0,
		Message:   MessageSuccess,
		Data:      data,
	}
}

// NewOKRespWithTrace creates a new OK response with trace context.
func NewOKRespWithTrace(ctx context.Context, data any) Resp {
	tracer := tracing.NewTraceContext()
	return Resp{
		ErrorCode: 0,
		Message:   MessageSuccess,
		Data:      data,
		TraceID:   tracer.GetTraceID(ctx),
	}
}

// OK sends 200 JSON with data and trace context.
func OK(c *gin.Context, data any) {
	resp := NewOKRespWithTrace(c.Request.Context(), data)
	c.JSON(http.StatusOK, resp)
}

// Created sends 201 JSON with data and trace context.
func Created(c *gin.Context, data any) {
	tracer := tracing.NewTraceContext()
	resp := Resp{
		ErrorCode: 0,
		Message:   MessageCreated,
		Data:      data,
		TraceID:   tracer.GetTraceID(c.Request.Context()),
	}
	c.JSON(http.StatusCreated, resp)
}

// Updated sends 200 JSON with updated data and trace context.
func Updated(c *gin.Context, data any) {
	tracer := tracing.NewTraceContext()
	resp := Resp{
		ErrorCode: 0,
		Message:   MessageUpdated,
		Data:      data,
		TraceID:   tracer.GetTraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, resp)
}

// Deleted sends 200 JSON with deletion confirmation and trace context.
func Deleted(c *gin.Context) {
	tracer := tracing.NewTraceContext()
	resp := Resp{
		ErrorCode: 0,
		Message:   MessageDeleted,
		TraceID:   tracer.GetTraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, resp)
}

// NoContent sends 204 No Content with trace headers.
func NoContent(c *gin.Context) {
	tracer := tracing.NewTraceContext()
	if traceID := tracer.GetTraceID(c.Request.Context()); traceID != "" {
		c.Header("X-Trace-Id", traceID)
	}
	c.Status(http.StatusNoContent)
}

// Paginated sends paginated response with trace context.
func Paginated(c *gin.Context, data any, pagination *Pagination) {
	tracer := tracing.NewTraceContext()
	resp := PaginatedResp{
		ErrorCode:  0,
		Message:    MessageSuccess,
		Data:       data,
		Pagination: pagination,
		TraceID:    tracer.GetTraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, resp)
}

// Health sends health check response with trace context.
func Health(c *gin.Context, status string, checks map[string]string, version string) {
	tracer := tracing.NewTraceContext()

	var statusCode int
	switch status {
	case StatusHealthy:
		statusCode = http.StatusOK
	case StatusDegraded:
		statusCode = http.StatusOK
	case StatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusOK
	}

	resp := HealthResp{
		Status:    status,
		Timestamp: time.Now(),
		Version:   version,
		Checks:    checks,
		TraceID:   tracer.GetTraceID(c.Request.Context()),
	}
	c.JSON(statusCode, resp)
}

// NewPagination creates pagination metadata.
func NewPagination(page, limit int, total int64) *Pagination {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	return &Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
