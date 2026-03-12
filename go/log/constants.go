package log

const (
	// Mode constants
	ModeProduction  = "production"
	ModeDevelopment = "development"

	// Encoding constants
	EncodingConsole = "console"
	EncodingJSON    = "json"

	// Level constants
	LevelDebug  = "debug"
	LevelInfo   = "info"
	LevelWarn   = "warn"
	LevelError  = "error"
	LevelFatal  = "fatal"
	LevelPanic  = "panic"
	LevelDPanic = "dpanic"

	// TraceIDKey is the key for trace_id in context
	TraceIDKey = "trace_id"
)
