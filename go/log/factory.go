package log

import "os"

// NewLogger creates a new logger with the specified configuration
func NewLogger(cfg ZapConfig) Logger {
	return NewZapLogger(cfg)
}

// NewDevelopmentLogger creates a logger suitable for development
func NewDevelopmentLogger() Logger {
	return NewZapLogger(DefaultDevelopmentConfig)
}

// NewProductionLogger creates a logger suitable for production
func NewProductionLogger() Logger {
	return NewZapLogger(DefaultProductionConfig)
}

// NewLoggerFromEnv creates a logger based on environment variables
func NewLoggerFromEnv() Logger {
	cfg := ZapConfig{
		Level:        getEnvOrDefault("LOG_LEVEL", LevelInfo),
		Mode:         getEnvOrDefault("LOG_MODE", ModeDevelopment),
		Encoding:     getEnvOrDefault("LOG_ENCODING", EncodingConsole),
		ColorEnabled: getEnvOrDefault("LOG_COLOR", "true") == "true",
	}

	// Auto-detect production mode
	if cfg.Mode == ModeProduction {
		cfg.Encoding = EncodingJSON
		cfg.ColorEnabled = false
	}

	return NewZapLogger(cfg)
}

// getEnvOrDefault returns the environment variable value or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
