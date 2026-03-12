package cron

import (
	"fmt"
)

// Default global cron instance
var defaultCron Cron

func init() {
	var err error
	defaultCron, err = NewCron(DefaultConfig())
	if err != nil {
		panic(fmt.Sprintf("failed to initialize default cron: %v", err))
	}
}

// SetDefaultCron sets the global default cron scheduler
func SetDefaultCron(c Cron) {
	defaultCron = c
}

// GetDefaultCron returns the global default cron scheduler
func GetDefaultCron() Cron {
	return defaultCron
}

// AddJob adds a new job using the default cron scheduler
func AddJob(info JobInfo) error {
	return defaultCron.AddJob(info)
}

// Start starts the default cron scheduler
func Start() {
	defaultCron.Start()
}

// Stop stops the default cron scheduler
func Stop() {
	defaultCron.Stop()
}

// SetFuncWrapper sets a function wrapper for the default cron scheduler
func SetFuncWrapper(f func(HandleFunc)) {
	defaultCron.SetFuncWrapper(f)
}

// Backwards compatibility constructor (returns the struct instead of interface to match old behavior if needed, 
// but here we keep it simple as the old code used values, not pointers in New())
func New() Cron {
	c, _ := NewCron(DefaultConfig())
	return c
}
