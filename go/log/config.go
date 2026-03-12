package log

// ZapConfig holds configuration for the Zap logger
type ZapConfig struct {
	Level        string // debug, info, warn, error, fatal, panic, dpanic
	Mode         string // production, development
	Encoding     string // console, json
	ColorEnabled bool   // enable colored output for console encoding
}

// Default configurations
var (
	// DefaultDevelopmentConfig provides sensible defaults for development
	DefaultDevelopmentConfig = ZapConfig{
		Level:        LevelInfo,
		Mode:         ModeDevelopment,
		Encoding:     EncodingConsole,
		ColorEnabled: true,
	}

	// DefaultProductionConfig provides sensible defaults for production
	DefaultProductionConfig = ZapConfig{
		Level:        LevelInfo,
		Mode:         ModeProduction,
		Encoding:     EncodingJSON,
		ColorEnabled: false,
	}
)
