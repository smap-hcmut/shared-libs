package response

const (
	// Stack trace and error handling
	DefaultStackTraceDepth = 32
	DefaultErrorMessage    = "Something went wrong"

	// Success messages
	MessageSuccess = "Success"

	// Error codes and messages
	ValidationErrorCode     = 400
	ValidationErrorMsg      = "Validation error"
	UnauthorizedErrorCode   = 401
	UnauthorizedErrorMsg    = "Authentication required"
	PermissionErrorCode     = 403
	PermissionErrorMsg      = "You don't have permission to do this"
	InternalServerErrorCode = 500

	// Date and time formats
	DateFormat     = "2006-01-02"
	DateTimeFormat = "2006-01-02 15:04:05"

	// External service limits
	DiscordMaxMessageLen = 5000
)
