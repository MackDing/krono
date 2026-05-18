// Package scheduler fires registered jobs at their scheduled times.
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/MackDing/krono/internal/schedule"
)

// Job is a unit of scheduled work. Key identifies the job across Sync calls.
type Job struct {
	Key      string
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
	wake    chan struct{}
}

// New returns an empty Scheduler.
func New() *Scheduler {
	return &Scheduler{wake: make(chan struct{}, 1)}
}

// Sync reconciles the registered jobs with jobs, matched by Job.Key: new keys
// are added, missing keys are dropped, and a key present in both keeps its
// next fire time unless its schedule changed. It is safe to call repeatedly
// while Run is executing — Run picks up the change promptly.
func (s *Scheduler) Sync(jobs []*Job) {
	now := time.Now()
	s.mu.Lock()
	prev := make(map[string]*entry, len(s.entries))
	for _, e := range s.entries {
		prev[e.job.Key] = e
	}
	next := make([]*entry, 0, len(jobs))
	for _, j := range jobs {
		if e, ok := prev[j.Key]; ok && e.job.Schedule.String() == j.Schedule.String() {
			next = append(next, &entry{job: j, next: e.next}) // keep timing, refresh job
		} else {
			next = append(next, &entry{job: j, next: j.Schedule.Next(now)})
		}
	}
	s.entries = next
	s.mu.Unlock()
	s.wakeup()
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
		case <-s.wake:
			timer.Stop()
			continue // the job set changed; recompute the soonest fire time
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

// wakeup nudges Run to recompute its timer after the job set changes.
func (s *Scheduler) wakeup() {
	select {
	case s.wake <- struct{}{}:
	default:
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

// nextFor returns the next fire time of the job with the given key.
func (s *Scheduler) nextFor(key string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.entries {
		if e.job.Key == key {
			return e.next, true
		}
	}
	return time.Time{}, false
}
