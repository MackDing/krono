# 3. Live job reload via periodic resync

Status: accepted

## Context

Jobs are created, edited, and deleted in the web dashboard, which writes to the
store. The running scheduler holds its own in-memory set of jobs. Without a way
to reconcile the two, dashboard changes would only take effect after restarting
krono — a poor first experience.

Two options: (a) the web layer pushes each change directly to the scheduler, or
(b) the scheduler periodically re-reads the store and reconciles.

## Decision

Periodic resync — option (b). `runner.Sync` rebuilds the scheduler's job set
from the store; `cmd/krono` calls it every 10 seconds. `Scheduler.Sync`
reconciles by job ID and preserves the next fire time of any job whose schedule
is unchanged.

## Consequences

- The web layer stays decoupled from the scheduler — it only touches the store.
- A dashboard change takes up to ~10s to take effect, not instantly. Acceptable
  for a scheduler; revisit with an explicit push if instant reload is needed.
- The reconcile MUST preserve next-fire times for unchanged jobs. Otherwise each
  resync would push every job's next fire further out and nothing would ever
  run — see `TestSyncKeepsNextFireForUnchangedJobs`.
