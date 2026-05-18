package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/MackDing/krono/internal/schedule"
)

func mustParse(t *testing.T, spec string) schedule.Schedule {
	t.Helper()
	s, err := schedule.Parse(spec)
	if err != nil {
		t.Fatalf("parse %q: %v", spec, err)
	}
	return s
}

func TestSchedulerFiresOnInterval(t *testing.T) {
	s := New()
	var mu sync.Mutex
	fires := 0
	s.Sync([]*Job{{
		Key:      "tick",
		Name:     "tick",
		Schedule: mustParse(t, "@every 50ms"),
		Run: func(ctx context.Context, fired time.Time) {
			mu.Lock()
			fires++
			mu.Unlock()
		},
	}})
	if got := s.Jobs(); got != 1 {
		t.Fatalf("Jobs() = %d, want 1", got)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 280*time.Millisecond)
	defer cancel()
	s.Run(ctx)

	mu.Lock()
	got := fires
	mu.Unlock()
	if got < 3 || got > 6 {
		t.Fatalf("got %d fires in 280ms at a 50ms interval, want 3-6", got)
	}
}

// TestSyncReconciles checks that Sync adds new jobs, drops missing ones, and
// keeps jobs present in both sets.
func TestSyncReconciles(t *testing.T) {
	s := New()
	noop := func(context.Context, time.Time) {}
	job := func(key string) *Job {
		return &Job{Key: key, Name: key, Schedule: mustParse(t, "@every 1m"), Run: noop}
	}

	s.Sync([]*Job{job("1"), job("2")})
	if s.Jobs() != 2 {
		t.Fatalf("after Sync(1,2): Jobs() = %d, want 2", s.Jobs())
	}

	s.Sync([]*Job{job("2"), job("3")}) // 1 dropped, 3 added, 2 kept
	if s.Jobs() != 2 {
		t.Fatalf("after Sync(2,3): Jobs() = %d, want 2", s.Jobs())
	}
	if _, ok := s.nextFor("1"); ok {
		t.Error("job 1 should have been dropped")
	}
	if _, ok := s.nextFor("3"); !ok {
		t.Error("job 3 should have been added")
	}

	s.Sync(nil)
	if s.Jobs() != 0 {
		t.Fatalf("after Sync(nil): Jobs() = %d, want 0", s.Jobs())
	}
}

// TestSyncKeepsNextFireForUnchangedJobs is critical: a job present in both the
// old and new sets must keep its next fire time, otherwise a periodic re-Sync
// would push every job out forever and nothing would ever fire.
func TestSyncKeepsNextFireForUnchangedJobs(t *testing.T) {
	s := New()
	noop := func(context.Context, time.Time) {}
	mk := func() *Job { return &Job{Key: "1", Name: "j", Schedule: mustParse(t, "@every 1h"), Run: noop} }

	s.Sync([]*Job{mk()})
	first, ok := s.nextFor("1")
	if !ok {
		t.Fatal("job 1 missing after the first Sync")
	}

	s.Sync([]*Job{mk()}) // same key, same schedule
	second, _ := s.nextFor("1")
	if !second.Equal(first) {
		t.Fatalf("re-Sync changed the next fire time: %v -> %v", first, second)
	}
}
