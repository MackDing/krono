# Krono — Context

The shared vocabulary for Krono. Contributors and tools should use these terms
consistently — in code, commit messages, issues, and docs.

## Domain language

- **Job** — a unit of recurring work: a name, a schedule, a type, a command,
  and an enabled flag. Persisted in the store (`store.Job`).
- **Run** — one execution of a job: start and end times, status, exit code, and
  captured output (`store.Run`).
- **Schedule** — when a job fires: a 5-field cron expression (`0 8 * * 1-5`), an
  interval (`@every 30s`), or a shortcut (`@daily`). Parsed by `internal/schedule`
  into a value that computes the next fire time.
- **Status** — a run's verdict: `success`, `failure`, `cancelled` (interrupted
  by shutdown), or `running`.
- **Type** — how a job runs: `shell` (a shell command) or `http` (an HTTP GET).

## Components

- **store** (`internal/store`) — embedded SQLite persistence for jobs and runs.
- **schedule** (`internal/schedule`) — parses schedule specs; computes next fire times.
- **scheduler** (`internal/scheduler`) — holds the live job set; fires jobs on time.
- **executor** (`internal/executor`) — runs one job and produces a `Run`.
- **notify** (`internal/notify`) — POSTs a webhook when a run fails.
- **runner** (`internal/runner`) — wires the above together; `Sync` reconciles
  the scheduler with the store.
- **web** (`internal/web`) — the dashboard: a JSON API and an embedded HTML UI.

## Key behaviours

- **Sync / resync** — reconciling the scheduler's live job set with the store,
  matched by job ID. `cmd/krono` re-syncs every 10s, so dashboard edits take
  effect without a restart. A job whose schedule is unchanged keeps its next
  fire time across a resync.
- **Single binary** — `krono` is one static Go binary: SQLite is pure-Go (no
  cgo) and the web UI is embedded with `go:embed`.

The reasoning behind the larger architecture decisions lives in `docs/adr/`.
