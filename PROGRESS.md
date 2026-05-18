# Krono — Progress

Handoff note for the long-running build/growth loop. Every session reads this
first and updates it before stopping. Conventions: `CLAUDE.md`.

## Done — MVP COMPLETE and the repo is LIVE
- MVP-1..5 plus the scheduler/executor/store wiring — a working self-hosted job
  scheduler in one Go binary. All 8 MVP contract rows green. Every milestone was
  built test-first and passed an independent evaluator. 32 tests, go vet clean.
- Repo: **github.com/MackDing/krono** — PUBLIC, pushed (9 commits on `main`).
  13 SEO topics set (cron, scheduler, job-scheduler, self-hosted, ai-agents, …).
  Module path is github.com/MackDing/krono.
- Growth dashboard (tools/dashboard.ps1) now tracks the live repo — day-1
  baseline 2026-05-18: Krono 0 stars; 10 competitors snapshotted.

## State
Krono is live and working: cron/interval schedules, SQLite persistence,
shell + HTTP job execution, run observability, an embedded web dashboard,
and failure webhooks — all in one Go binary. 0 stars (just published, not yet
announced anywhere).

## Next — the launch (needs the user; cannot be automated)
- Announce — the agent cannot post for you:
  - Show HN: "Show HN: Krono – a self-hosted job scheduler with a web dashboard"
  - r/selfhosted, Product Hunt — timed to the GEO windows (NA 14:00-17:00 UTC,
    Tue-Thu is the highest-leverage slot).
- Pre-announce polish: a demo GIF / screenshot in the README; a short docs page.
- Run tools/dashboard.ps1 daily (Windows Task Scheduler) — the PDCA C-Check.
- Iterate per the dashboard: SEO (topics/keywords), GEO (channels), product
  (the roadmap), ops (which channels actually convert).

## Notes
- Go: C:\Users\RD16019\go-toolchain\go\bin\go.exe ; go.mod: go 1.25.0, modernc.org/sqlite.
- git remote origin = https://github.com/MackDing/krono.git (HTTPS, gh credential helper).
- Hooks are PowerShell; .claude/settings.json not wired (classifier-blocked).
- Build loop: TDD -> independent evaluator -> flip contract -> commit checkpoint.
- Known follow-ups (non-blocking): live job reload (UI-created jobs need a
  restart); per-job execution timeout; the dashboard "1d change" column is
  really "change since last snapshot"; dashboard JS try/catch; a shared status
  constant instead of "failure"/"success" string literals.
- Targets: 3-month base 5k stars / stretch 10k; 100k = North Star. Star
  criteria in test-results.json are evidence-gated (GitHub API snapshot).
