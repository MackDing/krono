// Package runner wires the store, scheduler, executor, and notifier together:
// it loads persisted jobs onto the scheduler so each fire executes the job,
// records a Run, and webhook-notifies on failure.
package runner

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/MackDing/krono/internal/executor"
	"github.com/MackDing/krono/internal/notify"
	"github.com/MackDing/krono/internal/schedule"
	"github.com/MackDing/krono/internal/scheduler"
	"github.com/MackDing/krono/internal/store"
)

// Sync reconciles the scheduler with the enabled jobs currently in the store:
// jobs created, edited, or deleted (e.g. via the web dashboard) take effect on
// the next call — no restart needed. When a registered job fires it is
// executed, the Run is recorded, and a failed run POSTs a webhook to notifyURL.
// Sync is safe to call repeatedly and returns the number of jobs registered. A
// job with an unparseable schedule is skipped and logged, not fatal.
func Sync(sc *scheduler.Scheduler, st *store.Store, notifyURL string) (int, error) {
	jobs, err := st.ListJobs()
	if err != nil {
		return 0, fmt.Errorf("runner: list jobs: %w", err)
	}
	registered := make([]*scheduler.Job, 0, len(jobs))
	for _, j := range jobs {
		if !j.Enabled {
			continue
		}
		sch, err := schedule.Parse(j.Schedule)
		if err != nil {
			log.Printf("runner: skipping job %q (id %d): %v", j.Name, j.ID, err)
			continue
		}
		job := j
		registered = append(registered, &scheduler.Job{
			Key:      strconv.FormatInt(job.ID, 10),
			Name:     job.Name,
			Schedule: sch,
			Run: func(ctx context.Context, fired time.Time) {
				if err := recordRun(ctx, st, job, notifyURL); err != nil {
					log.Printf("%v", err)
				}
			},
		})
	}
	sc.Sync(registered)
	return len(registered), nil
}

// recordRun executes job, persists the resulting Run, and — if the run failed —
// sends a webhook notification. It is the unit the scheduler invokes per fire.
func recordRun(ctx context.Context, st *store.Store, job *store.Job, notifyURL string) error {
	run := executor.Execute(ctx, job)
	if err := st.CreateRun(run); err != nil {
		return fmt.Errorf("runner: record run for job %q: %w", job.Name, err)
	}
	if run.Status == "failure" {
		if err := notify.Webhook(ctx, notifyURL, job, run); err != nil {
			log.Printf("runner: notify for job %q: %v", job.Name, err)
		}
	}
	return nil
}
