package store

import (
	"path/filepath"
	"testing"
	"time"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "krono.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpenAndClose(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "krono.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestJobCRUD(t *testing.T) {
	s := openTest(t)

	j := &Job{Name: "backup", Schedule: "0 3 * * *", Type: "shell", Command: "backup.sh", Enabled: true}
	if err := s.CreateJob(j); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if j.ID == 0 {
		t.Fatal("CreateJob did not set ID")
	}
	if j.CreatedAt.IsZero() {
		t.Fatal("CreateJob did not set CreatedAt")
	}

	got, err := s.GetJob(j.ID)
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.Name != "backup" || got.Schedule != "0 3 * * *" || got.Command != "backup.sh" || !got.Enabled {
		t.Fatalf("GetJob mismatch: %+v", got)
	}

	jobs, err := s.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("ListJobs = %d jobs, want 1", len(jobs))
	}

	j.Name = "nightly-backup"
	j.Enabled = false
	if err := s.UpdateJob(j); err != nil {
		t.Fatalf("UpdateJob: %v", err)
	}
	got, _ = s.GetJob(j.ID)
	if got.Name != "nightly-backup" || got.Enabled {
		t.Fatalf("UpdateJob not persisted: %+v", got)
	}

	if err := s.DeleteJob(j.ID); err != nil {
		t.Fatalf("DeleteJob: %v", err)
	}
	if _, err := s.GetJob(j.ID); err == nil {
		t.Fatal("GetJob after delete: want error, got nil")
	}
}

func TestRunHistory(t *testing.T) {
	s := openTest(t)
	j := &Job{Name: "ping", Schedule: "@every 1m", Type: "http", Command: "https://example.com", Enabled: true}
	if err := s.CreateJob(j); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	r := &Run{JobID: j.ID, StartedAt: time.Now(), EndedAt: time.Now(), Status: "success", ExitCode: 0, Output: "ok"}
	if err := s.CreateRun(r); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if r.ID == 0 {
		t.Fatal("CreateRun did not set ID")
	}

	runs, err := s.ListRuns(j.ID)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("ListRuns = %d runs, want 1", len(runs))
	}
	if runs[0].Status != "success" || runs[0].Output != "ok" {
		t.Fatalf("ListRuns mismatch: %+v", runs[0])
	}
}

func TestPersistenceSurvivesRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "krono.db")

	// Session 1: write a job and a run, then close (simulating shutdown).
	s1, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open #1: %v", err)
	}
	j := &Job{Name: "survivor", Schedule: "@daily", Type: "shell", Command: "echo hi", Enabled: true}
	if err := s1.CreateJob(j); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if err := s1.CreateRun(&Run{JobID: j.ID, StartedAt: time.Now(), EndedAt: time.Now(), Status: "success"}); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if err := s1.Close(); err != nil {
		t.Fatalf("Close #1: %v", err)
	}

	// Session 2: reopen the same file — the data must still be there.
	s2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open #2: %v", err)
	}
	defer s2.Close()

	jobs, err := s2.ListJobs()
	if err != nil {
		t.Fatalf("ListJobs after restart: %v", err)
	}
	if len(jobs) != 1 || jobs[0].Name != "survivor" || jobs[0].Schedule != "@daily" {
		t.Fatalf("job did not survive restart: %+v", jobs)
	}
	runs, err := s2.ListRuns(jobs[0].ID)
	if err != nil {
		t.Fatalf("ListRuns after restart: %v", err)
	}
	if len(runs) != 1 || runs[0].Status != "success" {
		t.Fatalf("run did not survive restart: %+v", runs)
	}
}
