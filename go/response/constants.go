package response

const (
	// Stack trace and error handling
	DefaultStackTraceDepth = 32
	DefaultErrorMessage    = "Something went wrong"
	MessageSuccess         = "Success"
	MessageCreated         = "Created successfully"
	MessageUpdated         = "Updated successfully"
	MessageDeleted         = "Deleted successfully"

	// HTTP status error codes
	ValidationErrorCode     = 400
	UnauthorizedErrorCode   = 401
	PermissionErrorCode     = 403
	InternalServerErrorCode = 500

	// Error messages
	ValidationErrorMsg = "Validation error"
	PermissionErrorMsg = "You don't have permission to do this"

	// Date formats
	DateFormat     = "2006-01-02"
	DateTimeFormat = "2006-01-02 15:04:05"

	// Health status constants
	StatusHealthy   = "healthy"
	StatusDegraded  = "degraded"
	StatusUnhealthy = "unhealthy"

	// Pagination constants
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)
