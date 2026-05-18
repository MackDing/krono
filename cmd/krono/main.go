// Command krono is a self-hosted job scheduler.
//
// On start it opens its SQLite store, loads every enabled job, and runs them
// on schedule — executing each job and recording the run. The web UI is still
// being built; see README.md for the roadmap.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/krono-sh/krono/internal/runner"
	"github.com/krono-sh/krono/internal/scheduler"
	"github.com/krono-sh/krono/internal/store"
)

// version is the build version, overridden at release time via -ldflags.
var version = "0.0.0-dev"

func main() {
	runFor := time.Duration(0)
	dbPath := "krono.db"
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-v", "--version", "version":
			fmt.Printf("krono %s\n", version)
			return
		case "--for":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "krono: --for needs a duration, e.g. --for 30s")
				os.Exit(2)
			}
			d, err := time.ParseDuration(os.Args[i+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "krono: bad --for duration: %v\n", err)
				os.Exit(2)
			}
			runFor = d
			i++
		case "--db":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "krono: --db needs a path")
				os.Exit(2)
			}
			dbPath = os.Args[i+1]
			i++
		default:
			fmt.Fprintf(os.Stderr, "krono: unknown argument %q\n", os.Args[i])
			os.Exit(2)
		}
	}

	fmt.Printf("krono %s — self-hosted job scheduler\n", version)

	st, err := store.Open(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krono:", err)
		os.Exit(1)
	}
	defer st.Close()
	fmt.Printf("store: %s\n", dbPath)

	if err := seedDemoJobs(st); err != nil {
		fmt.Fprintln(os.Stderr, "krono: seed demo jobs:", err)
		os.Exit(1)
	}

	sc := scheduler.New()
	n, err := runner.Load(sc, st)
	if err != nil {
		fmt.Fprintln(os.Stderr, "krono:", err)
		os.Exit(1)
	}
	fmt.Printf("loaded %d enabled job(s)\n", n)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if runFor > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, runFor)
		defer cancel()
		fmt.Printf("running for %s (Ctrl+C to stop early)\n", runFor)
	} else {
		fmt.Println("running (Ctrl+C to stop)")
	}
	sc.Run(ctx)

	time.Sleep(200 * time.Millisecond) // let any in-flight runs finish before the summary
	fmt.Println("\nscheduler stopped — run history:")
	printRunSummary(st)
}

// seedDemoJobs inserts a couple of demo jobs the first time krono runs against
// a fresh, empty store, so there is something to schedule out of the box.
func seedDemoJobs(st *store.Store) error {
	jobs, err := st.ListJobs()
	if err != nil {
		return err
	}
	if len(jobs) > 0 {
		return nil
	}
	demo := []*store.Job{
		{Name: "heartbeat", Schedule: "@every 3s", Type: "shell", Command: "echo krono is alive", Enabled: true},
		{Name: "checkup", Schedule: "@every 5s", Type: "shell", Command: "echo running scheduled checkup", Enabled: true},
	}
	for _, j := range demo {
		if err := st.CreateJob(j); err != nil {
			return err
		}
	}
	fmt.Printf("seeded %d demo job(s) into a fresh store\n", len(demo))
	return nil
}

// printRunSummary shows, per job, how many runs are recorded and the latest.
func printRunSummary(st *store.Store) {
	jobs, err := st.ListJobs()
	if err != nil {
		fmt.Fprintln(os.Stderr, "krono:", err)
		return
	}
	for _, j := range jobs {
		runs, err := st.ListRuns(j.ID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "krono:", err)
			return
		}
		last := "never run"
		if len(runs) > 0 {
			last = fmt.Sprintf("last %s at %s", runs[0].Status, runs[0].StartedAt.Format("15:04:05"))
		}
		fmt.Printf("  %-11s %3d run(s)  %s\n", j.Name, len(runs), last)
	}
}
