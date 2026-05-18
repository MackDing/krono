// Package executor runs a job — a shell command or an HTTP request — and
// reports the result as a store.Run with timing, status, exit code, and
// captured output.
package executor

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/krono-sh/krono/internal/store"
)

// maxOutput caps how much job output is captured, keeping run records bounded.
const maxOutput = 64 * 1024

// Execute runs job and returns a completed store.Run. It does not return an
// error: any failure is reported inside the Run (Status "failure").
func Execute(ctx context.Context, job *store.Job) *store.Run {
	run := &store.Run{JobID: job.ID, StartedAt: time.Now()}
	switch job.Type {
	case "http":
		runHTTP(ctx, job.Command, run)
	default: // "shell" is the default job type
		runShell(ctx, job.Command, run)
	}
	run.EndedAt = time.Now()
	// A job interrupted by a cancelled context was not run to a verdict — it
	// was cut short (e.g. during shutdown). Record that as "cancelled" rather
	// than a misleading "failure"; a genuine success is left untouched.
	if ctx.Err() != nil && run.Status != "success" {
		run.Status = "cancelled"
	}
	return run
}

func runShell(ctx context.Context, command string, run *store.Run) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	out, err := cmd.CombinedOutput()
	run.Output = truncate(string(out))
	if err == nil {
		run.Status, run.ExitCode = "success", 0
		return
	}
	run.Status = "failure"
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		run.ExitCode = exitErr.ExitCode()
	} else {
		// the command could not start (shell missing, context cancelled, ...)
		run.ExitCode = -1
		if run.Output == "" {
			run.Output = err.Error()
		}
	}
}

func runHTTP(ctx context.Context, url string, run *store.Run) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		run.Status, run.ExitCode, run.Output = "failure", -1, err.Error()
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		run.Status, run.ExitCode, run.Output = "failure", -1, err.Error()
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxOutput))

	run.ExitCode = resp.StatusCode // the HTTP status is recorded as the exit code
	run.Output = truncate(string(body))
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		run.Status = "success"
	} else {
		run.Status = "failure"
	}
}

func truncate(s string) string {
	if len(s) > maxOutput {
		return s[:maxOutput] + "\n...[output truncated]"
	}
	return s
}
