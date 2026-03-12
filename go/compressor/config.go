package compressor

// Config defines the configuration for a compressor
type Config struct {
	Implementation Implementation `json:"implementation" yaml:"implementation"`
	Zstd           *ZstdConfig    `json:"zstd,omitempty" yaml:"zstd,omitempty"`
}

// ZstdConfig defines Zstd specific configuration
type ZstdConfig struct {
	DefaultLevel Level `json:"default_level" yaml:"default_level"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Implementation: ImplementationZstd,
		Zstd: &ZstdConfig{
			DefaultLevel: LevelDefault,
		},
	}
}
