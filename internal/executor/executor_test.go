package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krono-sh/krono/internal/store"
)

func TestExecuteShellSuccess(t *testing.T) {
	job := &store.Job{Type: "shell", Command: "echo hello-krono"}
	run := Execute(context.Background(), job)

	if run.Status != "success" {
		t.Fatalf("Status = %q, want success (output: %q)", run.Status, run.Output)
	}
	if run.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", run.ExitCode)
	}
	if !strings.Contains(run.Output, "hello-krono") {
		t.Fatalf("Output = %q, want it to contain hello-krono", run.Output)
	}
	if run.StartedAt.IsZero() || run.EndedAt.IsZero() || run.EndedAt.Before(run.StartedAt) {
		t.Fatalf("timing not captured: start=%v end=%v", run.StartedAt, run.EndedAt)
	}
}

func TestExecuteShellFailure(t *testing.T) {
	job := &store.Job{Type: "shell", Command: "exit 3"}
	run := Execute(context.Background(), job)

	if run.Status != "failure" {
		t.Fatalf("Status = %q, want failure", run.Status)
	}
	if run.ExitCode != 3 {
		t.Fatalf("ExitCode = %d, want 3", run.ExitCode)
	}
}

func TestExecuteHTTPSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))
	defer srv.Close()

	job := &store.Job{Type: "http", Command: srv.URL}
	run := Execute(context.Background(), job)

	if run.Status != "success" {
		t.Fatalf("Status = %q, want success", run.Status)
	}
	if run.ExitCode != 200 {
		t.Fatalf("ExitCode (HTTP status) = %d, want 200", run.ExitCode)
	}
	if !strings.Contains(run.Output, "pong") {
		t.Fatalf("Output = %q, want it to contain pong", run.Output)
	}
}

func TestExecuteHTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	job := &store.Job{Type: "http", Command: srv.URL}
	run := Execute(context.Background(), job)

	if run.Status != "failure" {
		t.Fatalf("Status = %q, want failure", run.Status)
	}
	if run.ExitCode != 500 {
		t.Fatalf("ExitCode (HTTP status) = %d, want 500", run.ExitCode)
	}
}

// TestRunIsObservableViaStore proves the end-to-end observability path:
// execute a job, persist the run, and read every recorded field back.
func TestRunIsObservableViaStore(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "krono.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer s.Close()

	job := &store.Job{Name: "echo-job", Schedule: "@every 1m", Type: "shell", Command: "echo observable", Enabled: true}
	if err := s.CreateJob(job); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	run := Execute(context.Background(), job)
	if err := s.CreateRun(run); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	runs, err := s.ListRuns(job.ID)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("ListRuns = %d runs, want 1", len(runs))
	}
	got := runs[0]
	if got.Status != "success" || got.ExitCode != 0 {
		t.Fatalf("run not observable: status=%q exit=%d", got.Status, got.ExitCode)
	}
	if got.StartedAt.IsZero() || got.EndedAt.IsZero() {
		t.Fatal("run start/end times not recorded")
	}
	if !strings.Contains(got.Output, "observable") {
		t.Fatalf("run output not recorded: %q", got.Output)
	}
}

// TestExecuteCancelledContext verifies that a job interrupted by a cancelled
// context is recorded as "cancelled", not as a misleading "failure".
func TestExecuteCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before the job runs

	run := Execute(ctx, &store.Job{Type: "shell", Command: "echo hi"})
	if run.Status != "cancelled" {
		t.Fatalf("Status = %q, want cancelled when the context is already cancelled", run.Status)
	}
}
