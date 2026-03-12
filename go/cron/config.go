package cron

// Config defines the configuration for the cron scheduler
type Config struct {
	Implementation Implementation `json:"implementation" yaml:"implementation"`
	Robfig         *RobfigConfig  `json:"robfig,omitempty" yaml:"robfig,omitempty"`
}

// RobfigConfig defines specific configuration for robfig/cron
type RobfigConfig struct {
	// WithSeconds indicates if the parser should support seconds (optional)
	WithSeconds bool `json:"with_seconds" yaml:"with_seconds"`
}

// DefaultConfig returns the default cron configuration
func DefaultConfig() Config {
	return Config{
		Implementation: ImplementationRobfig,
		Robfig: &RobfigConfig{
			WithSeconds: true,
		},
	}
}
