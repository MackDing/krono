// Package store persists Krono jobs and run history in an embedded SQLite
// database, so that scheduled jobs survive a process restart.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const tsFormat = time.RFC3339Nano

// Job is a persisted scheduled job.
type Job struct {
	ID        int64
	Name      string
	Schedule  string // a spec parseable by internal/schedule, e.g. "0 8 * * *"
	Type      string // "shell" or "http"
	Command   string // shell command line, or HTTP URL
	Enabled   bool
	CreatedAt time.Time
}

// Run is one persisted execution of a job.
type Run struct {
	ID        int64
	JobID     int64
	StartedAt time.Time
	EndedAt   time.Time // zero while still running
	Status    string    // "running", "success", "failure", "cancelled"
	ExitCode  int
	Output    string
}

// Store is a handle to the Krono SQLite database.
type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS jobs (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	name       TEXT NOT NULL,
	schedule   TEXT NOT NULL,
	type       TEXT NOT NULL DEFAULT 'shell',
	command    TEXT NOT NULL DEFAULT '',
	enabled    INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS runs (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	job_id     INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
	started_at TEXT NOT NULL,
	ended_at   TEXT NOT NULL DEFAULT '',
	status     TEXT NOT NULL,
	exit_code  INTEGER NOT NULL DEFAULT 0,
	output     TEXT NOT NULL DEFAULT ''
);`

// Open opens the SQLite database at path, creating the file, its parent
// directory, and the schema if they do not already exist.
func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("store: create data dir: %w", err)
		}
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	// A single connection keeps PRAGMAs effective and avoids SQLite lock contention.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: enable foreign keys: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

// CreateJob inserts j and sets j.ID and j.CreatedAt.
func (s *Store) CreateJob(j *Job) error {
	if j.Type == "" {
		j.Type = "shell"
	}
	j.CreatedAt = time.Now()
	res, err := s.db.Exec(
		`INSERT INTO jobs (name, schedule, type, command, enabled, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		j.Name, j.Schedule, j.Type, j.Command, boolToInt(j.Enabled), j.CreatedAt.Format(tsFormat))
	if err != nil {
		return fmt.Errorf("store: create job: %w", err)
	}
	if j.ID, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("store: create job: %w", err)
	}
	return nil
}

// GetJob returns the job with the given id.
func (s *Store) GetJob(id int64) (*Job, error) {
	row := s.db.QueryRow(
		`SELECT id, name, schedule, type, command, enabled, created_at FROM jobs WHERE id = ?`, id)
	return scanJob(row)
}

// ListJobs returns every job, oldest first.
func (s *Store) ListJobs() ([]*Job, error) {
	rows, err := s.db.Query(
		`SELECT id, name, schedule, type, command, enabled, created_at FROM jobs ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("store: list jobs: %w", err)
	}
	defer rows.Close()
	var jobs []*Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// UpdateJob writes j's mutable fields back to the database.
func (s *Store) UpdateJob(j *Job) error {
	res, err := s.db.Exec(
		`UPDATE jobs SET name = ?, schedule = ?, type = ?, command = ?, enabled = ? WHERE id = ?`,
		j.Name, j.Schedule, j.Type, j.Command, boolToInt(j.Enabled), j.ID)
	if err != nil {
		return fmt.Errorf("store: update job: %w", err)
	}
	return affectedOne(res, "update job", j.ID)
}

// DeleteJob removes a job and, via cascade, its run history.
func (s *Store) DeleteJob(id int64) error {
	res, err := s.db.Exec(`DELETE FROM jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("store: delete job: %w", err)
	}
	return affectedOne(res, "delete job", id)
}

// CreateRun inserts a run record and sets r.ID.
func (s *Store) CreateRun(r *Run) error {
	res, err := s.db.Exec(
		`INSERT INTO runs (job_id, started_at, ended_at, status, exit_code, output) VALUES (?, ?, ?, ?, ?, ?)`,
		r.JobID, r.StartedAt.Format(tsFormat), formatTime(r.EndedAt), r.Status, r.ExitCode, r.Output)
	if err != nil {
		return fmt.Errorf("store: create run: %w", err)
	}
	if r.ID, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("store: create run: %w", err)
	}
	return nil
}

// ListRuns returns a job's runs, most recent first.
func (s *Store) ListRuns(jobID int64) ([]*Run, error) {
	rows, err := s.db.Query(
		`SELECT id, job_id, started_at, ended_at, status, exit_code, output FROM runs WHERE job_id = ? ORDER BY id DESC`,
		jobID)
	if err != nil {
		return nil, fmt.Errorf("store: list runs: %w", err)
	}
	defer rows.Close()
	var runs []*Run
	for rows.Next() {
		var (
			r              Run
			started, ended string
		)
		if err := rows.Scan(&r.ID, &r.JobID, &started, &ended, &r.Status, &r.ExitCode, &r.Output); err != nil {
			return nil, fmt.Errorf("store: scan run: %w", err)
		}
		r.StartedAt, _ = parseTime(started)
		r.EndedAt, _ = parseTime(ended)
		runs = append(runs, &r)
	}
	return runs, rows.Err()
}

// --- helpers ---------------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(sc rowScanner) (*Job, error) {
	var (
		j       Job
		enabled int
		created string
	)
	if err := sc.Scan(&j.ID, &j.Name, &j.Schedule, &j.Type, &j.Command, &enabled, &created); err != nil {
		return nil, fmt.Errorf("store: get job: %w", err)
	}
	j.Enabled = enabled != 0
	j.CreatedAt, _ = parseTime(created)
	return &j, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(tsFormat)
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(tsFormat, s)
}

func affectedOne(res sql.Result, op string, id int64) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: %s: %w", op, err)
	}
	if n == 0 {
		return fmt.Errorf("store: %s: no job with id %d", op, id)
	}
	return nil
}
