# Krono — Progress

Handoff note for the long-running build/growth loop. Every session reads this
first and updates it before stopping. Conventions: `CLAUDE.md`.

## Done
- P-Plan due diligence: competitor data (30 repos), realistic targets, SEO/GEO
  matrices, direction locked (B+A: self-hosted scheduler + AI-agent spearhead).
- Top-level plan: name (Krono, working), MVP scope, milestones.
- Repo scaffolded; long-running harness installed (evaluator subagent, 5
  PowerShell hooks, default-FAIL contract). NOTE: .claude/settings.json (hook
  wiring) was blocked by the auto-mode classifier — needs explicit user
  authorization; the harness primitives still work as conventions.
- Growth dashboard: tools/dashboard.ps1 — daily GitHub-API snapshots.
- Go 1.26.3 installed (portable zip, no admin) at ~/go-toolchain.
- MVP-1 mvp-scheduler-engine — DONE. cron + interval parser + firing loop
  (internal/schedule, internal/scheduler). Evaluator PASS. Commit e6ffdff.
- MVP-2 mvp-persistence — DONE. internal/store: embedded SQLite (pure-Go
  modernc.org/sqlite) for jobs + run history; built test-first (TDD red→green);
  TestPersistenceSurvivesRestart proves restart survival. go test 8/8, go vet
  clean, independent evaluator PASS (2026-05-18). Contract flipped.

## In progress
- MVP-3 — job execution + run observability. New package internal/executor:
  run shell-command and HTTP jobs, capture exit code / duration / output as a
  store.Run. Satisfies mvp-job-shell, mvp-job-http, mvp-run-observability.

## Next  (build order — one item per session)
1. Wire scheduler -> executor -> store: a fired job runs and its Run is recorded.
2. mvp-web-dashboard — embedded web UI: job CRUD + run history + logs.
3. mvp-single-binary — embed the UI; one static binary.
4. mvp-failure-notify — webhook on failed run.

## Notes
- Go: C:\Users\RD16019\go-toolchain\go\bin\go.exe (on User PATH; new shells get
  it after a restart, otherwise call by full path).
- go.mod requires modernc.org/sqlite v1.50.1; go directive is 1.25.0.
- Hooks are PowerShell (no python3). settings.json not wired (see Done).
- Module path placeholder: github.com/krono-sh/krono — replace with the real
  GitHub org/user once known.
- Evidence in evidence/ : *-build.txt, *-test.txt, *.png.
- Targets: 3-month base 5k stars / stretch 10k; 100k = North Star. Star
  criteria in test-results.json are evidence-gated (GitHub API snapshot).
- Build loop: TDD (test first → red → implement → green) → independent
  evaluator → flip contract → commit checkpoint.
