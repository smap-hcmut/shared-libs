package cron

import (
	"sync"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	c := New()
	
	var wg sync.WaitGroup
	wg.Add(1)

	var jobDone bool
	err := c.AddJob(JobInfo{
		CronTime: "* * * * * *", // Every second
		Handler: func() {
			jobDone = true
			wg.Done()
		},
	})
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	c.Start()
	defer c.Stop()

	// Wait for job or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if !jobDone {
			t.Error("Job was not executed")
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for cron job")
	}
}

func TestCronWrapper(t *testing.T) {
	c := New()
	
	var wrapCount int
	c.SetFuncWrapper(func(f HandleFunc) {
		wrapCount++
		f()
	})

	var wg sync.WaitGroup
	wg.Add(1)

	err := c.AddJob(JobInfo{
		CronTime: "* * * * * *",
		Handler: func() {
			wg.Done()
		},
	})
	if err != nil {
		t.Fatalf("Failed to add job: %v", err)
	}

	c.Start()
	defer c.Stop()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if wrapCount == 0 {
			t.Error("Wrapper was not called")
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for cron job with wrapper")
	}
}
