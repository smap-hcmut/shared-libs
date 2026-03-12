package cron

// HandleFunc defines the function signature for cron jobs
type HandleFunc func()

// JobInfo contains information about a cron job
type JobInfo struct {
	Schedule string     `json:"schedule" yaml:"schedule"`
	Handler  HandleFunc `json:"-" yaml:"-"`
	Name     string     `json:"name,omitempty" yaml:"name,omitempty"`
}

// Cron defines the interface for cron job management
type Cron interface {
	// AddJob adds a new job to the cron scheduler
	AddJob(info JobInfo) error

	// Start starts the cron scheduler
	Start()

	// Stop stops the cron scheduler
	Stop()

	// SetFuncWrapper sets a wrapper for job functions (useful for tracing/logging)
	SetFuncWrapper(f func(HandleFunc))
}
