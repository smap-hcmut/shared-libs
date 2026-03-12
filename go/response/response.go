package response

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smap/shared-libs/go/tracing"
)

// Resp is the standard response format with trace integration
type Resp struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	Errors    any    `json:"errors,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`
}

// ErrorReporter interface for external error reporting (Discord, etc.)
type ErrorReporter interface {
	ReportBug(ctx context.Context, message string) error
}

// ResponseManager handles responses with trace integration
type ResponseManager struct {
	tracer   tracing.TraceContext
	reporter ErrorReporter // Optional
}

// NewResponseManager creates a new response manager with trace integration
func NewResponseManager(tracer tracing.TraceContext, reporter ErrorReporter) *ResponseManager {
	if tracer == nil {
		tracer = tracing.NewTraceContext()
	}
	return &ResponseManager{
		tracer:   tracer,
		reporter: reporter,
	}
}

// NewOKResp returns a new OK response with trace_id
func (rm *ResponseManager) NewOKResp(ctx context.Context, data any) Resp {
	resp := Resp{
		ErrorCode: 0,
		Message:   MessageSuccess,
		Data:      data,
	}

	// Add trace_id if available
	if traceID := rm.tracer.GetTraceID(ctx); traceID != "" {
		resp.TraceID = traceID
	}

	return resp
}

// OK sends 200 JSON with data and trace_id
func (rm *ResponseManager) OK(c *gin.Context, data any) {
	resp := rm.NewOKResp(c.Request.Context(), data)
	c.JSON(http.StatusOK, resp)
}

// Unauthorized sends 401 response with trace_id
func (rm *ResponseManager) Unauthorized(c *gin.Context) {
	ctx := c.Request.Context()
	resp := Resp{
		ErrorCode: UnauthorizedErrorCode,
		Message:   "Authentication required",
	}

	if traceID := rm.tracer.GetTraceID(ctx); traceID != "" {
		resp.TraceID = traceID
	}

	c.JSON(http.StatusUnauthorized, resp)
}

// Forbidden sends 403 response with trace_id
func (rm *ResponseManager) Forbidden(c *gin.Context) {
	ctx := c.Request.Context()
	resp := Resp{
		ErrorCode: PermissionErrorCode,
		Message:   PermissionErrorMsg,
	}

	if traceID := rm.tracer.GetTraceID(ctx); traceID != "" {
		resp.TraceID = traceID
	}

	c.JSON(http.StatusForbidden, resp)
}

// Error sends error response with trace_id and optional external reporting
func (rm *ResponseManager) Error(c *gin.Context, err error) {
	statusCode, resp := rm.parseError(err, c)
	c.JSON(statusCode, resp)
}

// parseError converts error to HTTP response with trace integration
func (rm *ResponseManager) parseError(err error, c *gin.Context) (int, Resp) {
	ctx := c.Request.Context()
	traceID := rm.tracer.GetTraceID(ctx)

	// Handle different error types (simplified version)
	// In a real implementation, you'd import the errors package and handle specific types

	resp := Resp{
		ErrorCode: InternalServerErrorCode,
		Message:   DefaultErrorMessage,
	}

	// Add trace_id to all error responses
	if traceID != "" {
		resp.TraceID = traceID
	}

	// For internal server errors, report to external system and capture stack trace
	if rm.reporter != nil {
		stackTrace := captureStackTrace()
		message := rm.buildErrorReport(c, err.Error(), stackTrace)
		go func() {
			if reportErr := rm.reporter.ReportBug(context.Background(), message); reportErr != nil {
				// Log the reporting error (would use shared logger in real implementation)
				_ = reportErr
			}
		}()
	}

	return http.StatusInternalServerError, resp
}

// captureStackTrace captures the current stack trace
func captureStackTrace() []string {
	var pcs [DefaultStackTraceDepth]uintptr
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return nil
	}

	var stackTrace []string
	for _, pc := range pcs[:n] {
		f := runtime.FuncForPC(pc)
		if f != nil {
			file, line := f.FileLine(pc)
			stackTrace = append(stackTrace, fmt.Sprintf("%s:%d %s", file, line, f.Name()))
		}
	}

	return stackTrace
}

// buildErrorReport creates detailed error report with trace_id
func (rm *ResponseManager) buildErrorReport(c *gin.Context, errString string, backtrace []string) string {
	ctx := c.Request.Context()
	traceID := rm.tracer.GetTraceID(ctx)

	url := c.Request.URL.String()
	method := c.Request.Method
	params := c.Request.URL.Query().Encode()

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		bodyBytes = []byte("Failed to read request body")
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	body := string(bodyBytes)

	var sb strings.Builder
	sb.WriteString("================ SMAP SERVICE ERROR ================\n")
	if traceID != "" {
		sb.WriteString(fmt.Sprintf("Trace ID: %s\n", traceID))
	}
	sb.WriteString(fmt.Sprintf("Route   : %s\n", url))
	sb.WriteString(fmt.Sprintf("Method  : %s\n", method))
	sb.WriteString("----------------------------------------------------\n")

	if len(c.Request.Header) > 0 {
		sb.WriteString("Headers :\n")
		for key, values := range c.Request.Header {
			sb.WriteString(fmt.Sprintf("    %s: %s\n", key, strings.Join(values, ", ")))
		}
		sb.WriteString("----------------------------------------------------\n")
	}

	if params != "" {
		sb.WriteString(fmt.Sprintf("Params  : %s\n", params))
	}

	if body != "" {
		sb.WriteString("Body    :\n")
		var prettyBody bytes.Buffer
		if err := json.Indent(&prettyBody, bodyBytes, "    ", "  "); err == nil {
			sb.WriteString(prettyBody.String() + "\n")
		} else {
			sb.WriteString("    " + body + "\n")
		}
		sb.WriteString("----------------------------------------------------\n")
	}

	sb.WriteString(fmt.Sprintf("Error   : %s\n", errString))

	if len(backtrace) > 0 {
		sb.WriteString("\nBacktrace:\n")
		for i, line := range backtrace {
			sb.WriteString(fmt.Sprintf("[%d]: %s\n", i, line))
		}
	}

	sb.WriteString("====================================================\n")
	return sb.String()
}

// Convenience functions for backward compatibility

// OK sends 200 JSON response (uses default response manager)
func OK(c *gin.Context, data any) {
	defaultManager := NewResponseManager(nil, nil)
	defaultManager.OK(c, data)
}

// Unauthorized sends 401 response (uses default response manager)
func Unauthorized(c *gin.Context) {
	defaultManager := NewResponseManager(nil, nil)
	defaultManager.Unauthorized(c)
}

// Forbidden sends 403 response (uses default response manager)
func Forbidden(c *gin.Context) {
	defaultManager := NewResponseManager(nil, nil)
	defaultManager.Forbidden(c)
}

// Error sends error response (uses default response manager)
func Error(c *gin.Context, err error) {
	defaultManager := NewResponseManager(nil, nil)
	defaultManager.Error(c, err)
}

// NewOKResp creates OK response (uses default response manager)
func NewOKResp(data any) Resp {
	return Resp{
		ErrorCode: 0,
		Message:   MessageSuccess,
		Data:      data,
	}
}
