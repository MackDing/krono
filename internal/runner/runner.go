// Package runner wires the store, scheduler, and executor together: it loads
// persisted jobs onto the scheduler so that each fire executes the job and
// records a Run back to the store.
package runner

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/krono-sh/krono/internal/executor"
	"github.com/krono-sh/krono/internal/schedule"
	"github.com/krono-sh/krono/internal/scheduler"
	"github.com/krono-sh/krono/internal/store"
)

// Load reads every enabled job from st, parses its schedule, and registers it
// with sc. When a registered job fires it is executed and the resulting Run is
// saved back to st. Load returns the number of jobs registered; it fails fast
// if any enabled job has an unparseable schedule.
func Load(sc *scheduler.Scheduler, st *store.Store) (int, error) {
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
				if err := recordRun(ctx, st, job); err != nil {
					log.Printf("%v", err)
				}
			},
		})
		registered++
	}
	return registered, nil
}

// recordRun executes job and persists the resulting Run to st. It is the unit
// the scheduler invokes on every fire.
func recordRun(ctx context.Context, st *store.Store, job *store.Job) error {
	run := executor.Execute(ctx, job)
	if err := st.CreateRun(run); err != nil {
		return fmt.Errorf("runner: record run for job %q: %w", job.Name, err)
	}
	return nil
}
