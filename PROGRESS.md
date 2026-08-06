# Krono — Progress

Handoff note for the long-running build/growth loop. Every session reads this
first and updates it before stopping. Conventions: `CLAUDE.md`.

## Done — MVP complete, repo LIVE, pre-launch hardening underway
- MVP-1..5 + wiring — a working self-hosted job scheduler in one Go binary.
  All 8 MVP contract rows green; every milestone TDD'd and evaluator-passed.
- Repo: github.com/MackDing/krono — PUBLIC, pushed, 13 SEO topics.
- Live job reload — jobs created / edited / deleted in the web dashboard take
  effect without restarting krono (Scheduler.Sync reconciles by key + a 10s
  resync loop; a wake channel fires new jobs promptly). TDD; 35 tests;
  independent evaluator PASS (e2e: a job added to the running instance fired
  without a restart, exact cadence held across resync cycles).
- Growth dashboard (tools/dashboard.ps1) tracks the live repo.

## State
Krono is live and working, and dashboard edits apply without a restart.
0 stars — not yet announced anywhere.

## Next — finish pre-launch, then announce (the announce needs the user)
- Demo screenshot / GIF of the dashboard for the README.
- A GitHub Actions CI workflow (go build / vet / test -race) — `-race` cannot
  run locally (no C compiler); CI on GitHub's runners closes that.
- Draft the announce posts: Show HN, r/selfhosted, Product Hunt.
- Announce — the agent cannot post; the user posts, timed to the GEO windows.
- Run tools/dashboard.ps1 daily (Task Scheduler) — the PDCA C-Check.

## Notes
- Toolchain: Go 1.25.0 (see `go.mod`), with `modernc.org/sqlite`.
- git remote origin = https://github.com/MackDing/krono.git (HTTPS).
- Hooks are PowerShell; .claude/settings.json not wired (classifier-blocked).
- Build loop: TDD -> independent evaluator -> commit checkpoint.
- Known follow-ups (non-blocking): per-job execution timeout; dashboard JS
  try/catch on the /runs fetch; a shared run-status constant; the dashboard
  "1d change" column is really "change since last snapshot"; resync latency is
  up to ~10s (acceptable). `go test -race` should run in CI.
- Targets: 3-month base 5k stars / stretch 10k; 100k = North Star.
