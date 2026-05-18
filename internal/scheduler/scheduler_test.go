package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/krono-sh/krono/internal/schedule"
)

func TestSchedulerFiresOnInterval(t *testing.T) {
	s := New()
	sch, err := schedule.Parse("@every 50ms")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var mu sync.Mutex
	fires := 0
	s.Add(&Job{
		Name:     "tick",
		Schedule: sch,
		Run: func(ctx context.Context, fired time.Time) {
			mu.Lock()
			fires++
			mu.Unlock()
		},
	})
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
