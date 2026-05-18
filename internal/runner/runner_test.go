package runner

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krono-sh/krono/internal/scheduler"
	"github.com/krono-sh/krono/internal/store"
)

func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "krono.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

// TestRecordRunExecutesAndPersists is the core wiring check: executing a stored
// job produces a Run that lands in the store and is queryable — deterministic,
// no scheduler timer involved.
func TestRecordRunExecutesAndPersists(t *testing.T) {
	st := openStore(t)
	job := &store.Job{Name: "echo", Schedule: "@every 1m", Type: "shell", Command: "echo wired", Enabled: true}
	if err := st.CreateJob(job); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	if err := recordRun(context.Background(), st, job); err != nil {
		t.Fatalf("recordRun: %v", err)
	}

	runs, err := st.ListRuns(job.ID)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("ListRuns = %d runs, want 1", len(runs))
	}
	got := runs[0]
	if got.Status != "success" || got.ExitCode != 0 {
		t.Fatalf("recorded run: status=%q exit=%d, want success/0", got.Status, got.ExitCode)
	}
	if !strings.Contains(got.Output, "wired") {
		t.Fatalf("recorded run output = %q, want it to contain wired", got.Output)
	}
	if got.StartedAt.IsZero() || got.EndedAt.IsZero() {
		t.Fatal("recorded run start/end times not set")
	}
}

// TestLoadRegistersEnabledJobsOnly checks that Load registers every enabled
// job onto the scheduler and skips disabled ones.
func TestLoadRegistersEnabledJobsOnly(t *testing.T) {
	st := openStore(t)
	for _, j := range []*store.Job{
		{Name: "on-interval", Schedule: "@every 1m", Type: "shell", Command: "echo a", Enabled: true},
		{Name: "on-cron", Schedule: "0 8 * * *", Type: "shell", Command: "echo b", Enabled: true},
		{Name: "disabled", Schedule: "@every 1m", Type: "shell", Command: "echo c", Enabled: false},
	} {
		if err := st.CreateJob(j); err != nil {
			t.Fatalf("CreateJob %q: %v", j.Name, err)
		}
	}

	sc := scheduler.New()
	n, err := Load(sc, st)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if n != 2 {
		t.Fatalf("Load registered %d jobs, want 2 (the disabled job must be skipped)", n)
	}
	if sc.Jobs() != 2 {
		t.Fatalf("scheduler has %d jobs, want 2", sc.Jobs())
	}
}

// TestLoadRejectsBadSchedule checks that an unparseable schedule fails Load.
func TestLoadRejectsBadSchedule(t *testing.T) {
	st := openStore(t)
	if err := st.CreateJob(&store.Job{
		Name: "bad", Schedule: "not-a-cron", Type: "shell", Command: "echo x", Enabled: true,
	}); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if _, err := Load(scheduler.New(), st); err == nil {
		t.Fatal("Load with an unparseable schedule: want error, got nil")
	}
}
