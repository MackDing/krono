// Package scheduler fires registered jobs at their scheduled times.
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/krono-sh/krono/internal/schedule"
)

// Job is a unit of scheduled work.
type Job struct {
	Name     string
	Schedule schedule.Schedule
	// Run is invoked in its own goroutine each time the job fires.
	Run func(ctx context.Context, fired time.Time)
}

type entry struct {
	job  *Job
	next time.Time
}

// Scheduler fires registered jobs at their scheduled times.
type Scheduler struct {
	mu      sync.Mutex
	entries []*entry
}

// New returns an empty Scheduler.
func New() *Scheduler { return &Scheduler{} }

// Add registers a job; its first fire time is computed from now.
func (s *Scheduler) Add(j *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, &entry{job: j, next: j.Schedule.Next(time.Now())})
}

// Jobs returns the number of registered jobs.
func (s *Scheduler) Jobs() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

// Run blocks, firing jobs until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	for {
		soonest := s.soonest()
		wait := time.Hour // nothing scheduled: wake periodically to re-check
		if !soonest.IsZero() {
			wait = time.Until(soonest)
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		// The timer and ctx.Done can become ready together; do not start jobs
		// once the context is done, so a fired job never runs against a dead
		// context (which would record a misleading failure during shutdown).
		if ctx.Err() != nil {
			return
		}
		s.fireDue(ctx, time.Now())
	}
}

func (s *Scheduler) soonest() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	var soonest time.Time
	for _, e := range s.entries {
		if e.next.IsZero() {
			continue
		}
		if soonest.IsZero() || e.next.Before(soonest) {
			soonest = e.next
		}
	}
	return soonest
}

func (s *Scheduler) fireDue(ctx context.Context, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.entries {
		if e.next.IsZero() || e.next.After(now) {
			continue
		}
		job, firedAt := e.job, e.next
		go job.Run(ctx, firedAt)
		e.next = job.Schedule.Next(now)
	}
}
