// Package runner wires the store, scheduler, executor, and notifier together:
// it loads persisted jobs onto the scheduler so each fire executes the job,
// records a Run, and webhook-notifies on failure.
package runner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MackDing/krono/internal/executor"
	"github.com/MackDing/krono/internal/notify"
	"github.com/MackDing/krono/internal/schedule"
	"github.com/MackDing/krono/internal/scheduler"
	"github.com/MackDing/krono/internal/store"
)

// Load reads every enabled job from st, parses its schedule, and registers it
// with sc. When a registered job fires it is executed and the resulting Run is
// saved back to st; a failed run additionally POSTs a webhook to notifyURL
// (when notifyURL is non-empty). Load returns the number of jobs registered;
// it fails fast if any enabled job has an unparseable schedule.
func Load(sc *scheduler.Scheduler, st *store.Store, notifyURL string) (int, error) {
	jobs, err := st.ListJobs()
	if err != nil {
		return 0, fmt.Errorf("runner: list jobs: %w", err)
	}
	registered := 0
	for _, j := range jobs {
		if !j.Enabled {
			continue
		}
		sch, err := schedule.Parse(j.Schedule)
		if err != nil {
			return registered, fmt.Errorf("runner: job %q (id %d): %w", j.Name, j.ID, err)
		}
		job := j
		sc.Add(&scheduler.Job{
			Name:     job.Name,
			Schedule: sch,
			Run: func(ctx context.Context, fired time.Time) {
				if err := recordRun(ctx, st, job, notifyURL); err != nil {
					log.Printf("%v", err)
				}
			},
		})
		registered++
	}
	return registered, nil
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
