package cron

// Implementation defines the type of cron implementation
type Implementation string

const (
	// ImplementationRobfig uses robfig/cron/v3
	ImplementationRobfig Implementation = "robfig"
)
