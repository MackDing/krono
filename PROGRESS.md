# Krono — Progress

Handoff note for the long-running build/growth loop. Every session reads this
first and updates it before stopping. Conventions: `CLAUDE.md`.

## Done
- P-Plan due diligence: competitor data (30 repos, live GitHub API), realistic
  targets, SEO/GEO matrices, direction locked (B+A: self-hosted scheduler with
  an AI-agent spearhead). Tech stack: Go + embedded SQLite + embedded web UI.
- Top-level plan: name (Krono, working), MVP scope, milestones.
- Repo scaffolded: README, LICENSE (MIT), go.mod, .gitignore.
- Long-running harness installed: evaluator subagent, 5 PowerShell hooks,
  default-FAIL contract (test-results.json), this file. NOTE: .claude/settings.json
  (the hook wiring) was NOT installed — the auto-mode classifier blocked it; it
  needs explicit user authorization. The harness primitives still work as
  conventions until then.
- Growth dashboard: tools/dashboard.ps1 — daily GitHub-API snapshots into
  dashboard/history.csv + DASHBOARD.md.
- Go 1.26.3 installed (portable zip, no admin) at ~/go-toolchain.
- MVP-1 mvp-scheduler-engine — DONE. cron + interval parser (internal/schedule)
  and the firing loop (internal/scheduler). go build/vet clean, go test 4/4,
  independent evaluator PASS (2026-05-18). Contract row flipped to true.

## In progress
- MVP-2 mvp-persistence — next to build.

## Next  (build order — one item per session)
1. mvp-persistence — SQLite store; jobs + run history survive restart.
2. mvp-job-shell — shell-command execution with output capture.
3. mvp-job-http — HTTP-request job execution.
4. mvp-run-observability — per-run records (exit code, duration, output).
5. mvp-web-dashboard — embedded web UI: job CRUD + run history + logs.
6. mvp-single-binary — embed the UI; one static binary.
7. mvp-failure-notify — webhook on failed run.

## Notes
- Go: C:\Users\RD16019\go-toolchain\go\bin\go.exe (added to the User PATH; new
  shells pick it up after a restart, otherwise call it by full path).
- Hooks are PowerShell (no python3 on this machine). settings.json not yet
  wired (see Done) — the harness primitives still work as conventions.
- Module path placeholder: github.com/krono-sh/krono — replace with the real
  GitHub org/user once known (touches go.mod + imports + README URLs).
- Evidence goes in evidence/ : *-build.txt, *-test.txt, *.png.
- Targets: 3-month base 5k stars / stretch 10k; 100k = North Star. All star
  criteria in test-results.json are evidence-gated (GitHub API snapshot).
- MVP-2: use a pure-Go SQLite driver (modernc.org/sqlite, no cgo) to keep the
  build single-binary and cross-compile friendly.
