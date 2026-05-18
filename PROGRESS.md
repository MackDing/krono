# Krono — Progress

Handoff note for the long-running build/growth loop. Every session reads this
first and updates it before stopping. Conventions: `CLAUDE.md`.

## Done
- P-Plan due diligence; top-level plan (name Krono, MVP scope, milestones).
- Repo scaffolded; long-running harness installed (evaluator subagent, 5
  PowerShell hooks, default-FAIL contract). NOTE: .claude/settings.json (hook
  wiring) blocked by the auto-mode classifier — needs explicit user auth.
- Growth dashboard: tools/dashboard.ps1 — daily GitHub-API snapshots.
- Go 1.26.3 installed (portable, no admin) at ~/go-toolchain.
- MVP-1 mvp-scheduler-engine — cron + interval parser + firing loop. Evaluator
  PASS. Commit e6ffdff.
- MVP-2 mvp-persistence — internal/store: embedded SQLite. Evaluator PASS.
  Commit d571317.
- MVP-3 mvp-job-shell + mvp-job-http + mvp-run-observability — internal/executor.
  Evaluator PASS. Commit a01ade3.
- Wiring scheduler->executor->store — internal/runner loads enabled jobs onto
  the scheduler; each fire executes the job and records a Run. cmd/krono opens
  the store, seeds demo jobs, runs via the runner. Two correctness fixes:
  scheduler does not fire after ctx is done; executor records an interrupted
  run as "cancelled" not "failure". 17 tests; independent evaluator PASS
  (2026-05-18). Verified end-to-end — krono.exe loads jobs from SQLite,
  executes them, and run history accumulates across restarts.

## In progress
- MVP-4 mvp-web-dashboard — embedded web UI: list jobs, create/edit/delete,
  show run history and logs. Plan: new internal/web (HTTP handlers, httptest),
  an embedded HTML page (go:embed), served by cmd/krono on a port. Build in
  slices: read-only views first, then job CRUD.

## Next  (build order — one item per session)
1. mvp-single-binary — confirm the web UI is embedded; one static binary.
2. mvp-failure-notify — webhook on failed run.

## Notes
- Go: C:\Users\RD16019\go-toolchain\go\bin\go.exe (User PATH; full path in this session).
- go.mod requires modernc.org/sqlite v1.50.1; go directive is 1.25.0.
- Hooks are PowerShell (no python3). settings.json not wired.
- Module path placeholder: github.com/krono-sh/krono — replace with the real
  GitHub org/user once known.
- Evidence in evidence/ : *-test.txt.
- Build loop: TDD (test first -> red -> implement -> green) -> independent
  evaluator -> commit checkpoint.
- Known follow-ups (non-blocking): per-job execution timeout in executor; run
  duration is coarse (two time.Now calls); main.go's 200ms pre-summary drain is
  best-effort.
- Targets: 3-month base 5k stars / stretch 10k; 100k = North Star.
