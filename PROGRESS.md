# Krono — Progress

Handoff note for the long-running build/growth loop. Every session reads this
first and updates it before stopping. Conventions: `CLAUDE.md`.

## Done
- P-Plan due diligence; top-level plan (name Krono, MVP scope, milestones).
- Repo scaffolded; long-running harness installed (evaluator subagent, 5
  PowerShell hooks, default-FAIL contract). NOTE: .claude/settings.json (hook
  wiring) blocked by the auto-mode classifier — needs explicit user auth; the
  harness primitives still work as conventions.
- Growth dashboard: tools/dashboard.ps1 — daily GitHub-API snapshots.
- Go 1.26.3 installed (portable zip, no admin) at ~/go-toolchain.
- MVP-1 mvp-scheduler-engine — cron + interval parser + firing loop
  (internal/schedule, internal/scheduler). Evaluator PASS. Commit e6ffdff.
- MVP-2 mvp-persistence — internal/store: embedded SQLite (modernc.org/sqlite).
  Evaluator PASS. Commit d571317.
- MVP-3 mvp-job-shell + mvp-job-http + mvp-run-observability — internal/executor:
  runs shell and HTTP jobs, captures exit code / duration / output as a
  store.Run. TDD red->green; go test 13/13; independent evaluator PASS
  (2026-05-18). Three contract rows flipped.

## In progress
- Wiring scheduler -> executor -> store (new internal/runner): load enabled
  jobs from the store, fire them on schedule, execute, and record each Run.
  Then update cmd/krono to open the store and run via the runner.

## Next  (build order — one item per session)
1. mvp-web-dashboard — embedded web UI: job CRUD + run history + logs.
2. mvp-single-binary — embed the UI; one static binary.
3. mvp-failure-notify — webhook on failed run.

## Notes
- Go: C:\Users\RD16019\go-toolchain\go\bin\go.exe (User PATH; use the full path
  in this session — new shells pick it up after a restart).
- go.mod requires modernc.org/sqlite v1.50.1; go directive is 1.25.0.
- Hooks are PowerShell (no python3). settings.json not wired (see Done).
- Module path placeholder: github.com/krono-sh/krono — replace with the real
  GitHub org/user once known.
- Evidence in evidence/ : *-test.txt.
- Build loop: TDD (test first -> red -> implement -> green) -> independent
  evaluator -> flip contract -> commit checkpoint.
- Known follow-ups (non-blocking, raised by evaluators): add a per-job
  execution timeout in internal/executor; run duration is coarse (two
  time.Now calls so an instant command can show duration 0).
- Targets: 3-month base 5k stars / stretch 10k; 100k = North Star.
