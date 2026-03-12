package cron

// NewCron creates a new cron scheduler with the specified configuration
func NewCron(cfg Config) (Cron, error) {
	switch cfg.Implementation {
	case ImplementationRobfig:
		robfigCfg := RobfigConfig{
			WithSeconds: true,
		}
		if cfg.Robfig != nil {
			robfigCfg = *cfg.Robfig
		}
		return NewRobfigCron(robfigCfg), nil
	default:
		return nil, ErrUnsupportedImplementation
	}
}
