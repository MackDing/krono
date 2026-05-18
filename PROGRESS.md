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
- MVP-1 mvp-scheduler-engine — cron + interval parser + firing loop. Commit e6ffdff.
- MVP-2 mvp-persistence — internal/store, embedded SQLite. Commit d571317.
- MVP-3 mvp-job-shell + mvp-job-http + mvp-run-observability — internal/executor.
  Commit a01ade3.
- Wiring scheduler->executor->store — internal/runner; cmd/krono runs persisted
  jobs end-to-end. Commit d3fc654.
- MVP-4 mvp-web-dashboard — internal/web: JSON API (slice 4a, commit 3e54fbb) +
  embedded single-page HTML dashboard (slice 4b, go:embed). cmd/krono serves it
  on :8400 alongside the scheduler. 28 tests; independent evaluator PASS on both
  slices; verified end-to-end over HTTP.
  All MVP-1..4 contract rows green; only MVP-5 remains.

## In progress
- MVP-5 — two remaining contract rows:
  - mvp-single-binary: verify `go build` yields one self-contained binary with
    the web UI embedded (go:embed + pure-Go SQLite already deliver this; needs
    a clean-room run to confirm).
  - mvp-failure-notify: new internal/notify — POST a webhook when a run fails;
    wired into internal/runner (recordRun), with a --notify-url flag on cmd/krono.

## Next  (after MVP-5, the MVP scope is complete)
- Pre-launch: README polish + demo GIF, pick the real GitHub org, push, launch
  prep (Show HN / Product Hunt / r/selfhosted), set up the daily dashboard cron.

## Notes
- Go: C:\Users\RD16019\go-toolchain\go\bin\go.exe (User PATH; full path in this session).
- go.mod requires modernc.org/sqlite v1.50.1; go directive is 1.25.0.
- Hooks are PowerShell (no python3). settings.json not wired.
- Module path placeholder: github.com/krono-sh/krono — replace with the real
  GitHub org/user once known.
- Evidence in evidence/ : *-test.txt.
- Build loop: TDD (test first -> red -> implement -> green) -> independent
  evaluator -> flip contract -> commit checkpoint.
- Known follow-ups (non-blocking, from evaluators): per-job execution timeout
  in executor; run duration coarse (two time.Now calls); dashboard JS
  fillLast/showRuns lack try/catch on the /runs fetch (cosmetic).
- Targets: 3-month base 5k stars / stretch 10k; 100k = North Star.
