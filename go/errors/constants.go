package errors

import "net/http"

// HTTP status codes for predefined errors
const (
	StatusBadRequest          = http.StatusBadRequest          // 400
	StatusUnauthorized        = http.StatusUnauthorized        // 401
	StatusForbidden           = http.StatusForbidden           // 403
	StatusNotFound            = http.StatusNotFound            // 404
	StatusConflict            = http.StatusConflict            // 409
	StatusUnprocessableEntity = http.StatusUnprocessableEntity // 422
	StatusInternalServerError = http.StatusInternalServerError // 500
	StatusBadGateway          = http.StatusBadGateway          // 502
	StatusServiceUnavailable  = http.StatusServiceUnavailable  // 503
)

// Default error messages
const (
	MessageBadRequest          = "Bad Request"
	MessageUnauthorized        = "Unauthorized"
	MessageForbidden           = "Forbidden"
	MessageNotFound            = "Not Found"
	MessageConflict            = "Conflict"
	MessageUnprocessableEntity = "Unprocessable Entity"
	MessageInternalServerError = "Internal Server Error"
	MessageBadGateway          = "Bad Gateway"
	MessageServiceUnavailable  = "Service Unavailable"
)

// Business error codes
const (
	CodeValidationFailed = "VALIDATION_FAILED"
	CodePermissionDenied = "PERMISSION_DENIED"
	CodeResourceNotFound = "RESOURCE_NOT_FOUND"
	CodeResourceConflict = "RESOURCE_CONFLICT"
	CodeBusinessLogic    = "BUSINESS_LOGIC_ERROR"
	CodeExternalService  = "EXTERNAL_SERVICE_ERROR"
	CodeDatabaseError    = "DATABASE_ERROR"
	CodeNetworkError     = "NETWORK_ERROR"
)

// System component names
const (
	ComponentHTTP     = "http"
	ComponentDatabase = "database"
	ComponentCache    = "cache"
	ComponentQueue    = "queue"
	ComponentAuth     = "auth"
	ComponentExternal = "external"
)
