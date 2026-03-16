package errors

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
