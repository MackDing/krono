// Command krono is a self-hosted job scheduler with a web dashboard.
//
// Run with no arguments to start the scheduler with a couple of demo jobs.
// Storage and the web UI are still being built; see README.md for the roadmap.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/krono-sh/krono/internal/schedule"
	"github.com/krono-sh/krono/internal/scheduler"
)

// version is the build version, overridden at release time via -ldflags.
var version = "0.0.0-dev"

func main() {
	runFor := time.Duration(0)
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
		default:
			fmt.Fprintf(os.Stderr, "krono: unknown argument %q\n", os.Args[i])
			os.Exit(2)
		}
	}

	fmt.Printf("krono %s — self-hosted job scheduler\n", version)

	s := scheduler.New()
	demo := []struct{ name, spec, msg string }{
		{"heartbeat", "@every 2s", "still alive"},
		{"reminder", "@every 5s", "drink water"},
	}
	for _, d := range demo {
		sch, err := schedule.Parse(d.spec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "krono: bad demo schedule %q: %v\n", d.spec, err)
			os.Exit(1)
		}
		name, msg := d.name, d.msg
		s.Add(&scheduler.Job{
			Name:     name,
			Schedule: sch,
			Run: func(ctx context.Context, fired time.Time) {
				fmt.Printf("  %s  [%s] %s\n", fired.Format("15:04:05"), name, msg)
			},
		})
		fmt.Printf("registered job %-11q schedule %s\n", d.name, d.spec)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if runFor > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, runFor)
		defer cancel()
		fmt.Printf("running %d jobs for %s (Ctrl+C to stop early)\n\n", s.Jobs(), runFor)
	} else {
		fmt.Printf("running %d jobs (Ctrl+C to stop)\n\n", s.Jobs())
	}
	s.Run(ctx)
	fmt.Println("\nscheduler stopped")
}
