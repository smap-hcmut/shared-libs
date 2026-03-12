package cron

import (
	"fmt"

	robfigcron "github.com/robfig/cron/v3"
)

type robfigCron struct {
	cron        *robfigcron.Cron
	funcWrapper func(HandleFunc)
}

// NewRobfigCron creates a new robfig/cron implementation
func NewRobfigCron(cfg RobfigConfig) Cron {
	var opts []robfigcron.Option

	if cfg.WithSeconds {
		parser := robfigcron.NewParser(
			robfigcron.SecondOptional |
				robfigcron.Minute |
				robfigcron.Hour |
				robfigcron.Dom |
				robfigcron.Month |
				robfigcron.Dow |
				robfigcron.Descriptor,
		)
		opts = append(opts, robfigcron.WithParser(parser))
	}

	return &robfigCron{
		cron: robfigcron.New(opts...),
	}
}

func (c *robfigCron) SetFuncWrapper(f func(HandleFunc)) {
	c.funcWrapper = f
}

func (c *robfigCron) getFuncWrapper() func(HandleFunc) {
	if c.funcWrapper == nil {
		return func(f HandleFunc) {
			f()
		}
	}
	return c.funcWrapper
}

func (c *robfigCron) AddJob(info JobInfo) error {
	if info.CronTime == "" {
		return ErrInvalidCronTime
	}
	if info.Handler == nil {
		return ErrJobHandlerRequired
	}

	fw := c.getFuncWrapper()

	_, err := c.cron.AddFunc(info.CronTime, func() {
		fw(info.Handler)
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	return nil
}

func (c *robfigCron) Start() {
	c.cron.Start()
}

func (c *robfigCron) Stop() {
	c.cron.Stop()
}
