# Krono — Progress

Handoff note for the long-running build/growth loop. Every session reads this
first and updates it before stopping. Conventions: `CLAUDE.md`.

## Done — MVP COMPLETE (all 8 contract rows green)
- P-Plan due diligence; top-level plan; repo scaffold; long-running harness;
  growth dashboard (tools/dashboard.ps1); Go 1.26.3 (portable, ~/go-toolchain).
- MVP-1 mvp-scheduler-engine — cron + interval parser + firing loop. e6ffdff
- MVP-2 mvp-persistence — internal/store, embedded SQLite. d571317
- MVP-3 mvp-job-shell + mvp-job-http + mvp-run-observability — internal/executor. a01ade3
- Wiring scheduler->executor->store — internal/runner. d3fc654
- MVP-4 mvp-web-dashboard — internal/web: JSON API + embedded HTML UI. 3e54fbb, 81db137
- MVP-5 mvp-failure-notify (internal/notify) + mvp-single-binary (clean-room verified).
- Every milestone built test-first (red->green) and passed an independent
  evaluator. 32 tests, go vet clean. krono.exe is a self-contained single binary.

## State
Krono is a working self-hosted job scheduler: cron/interval schedules, SQLite
persistence, shell + HTTP job execution, run observability, an embedded web
dashboard (job CRUD + run history), and failure webhooks — all in one Go binary.

## In progress
- (none — the MVP build is complete)

## Next — launch phase (needs user decisions)
- Confirm the project name (Krono is a working name) and the real GitHub
  org/user, then fix the module path (github.com/krono-sh/krono is a placeholder).
- Push to GitHub as a public repo (a user-authorized action).
- Pre-launch assets: README polish + a demo GIF/screenshot.
- Launch: Show HN + Product Hunt + r/selfhosted, timed per the GEO windows.
- Start the daily PDCA loop: tools/dashboard.ps1 on a schedule once the repo is
  public; iterate SEO/GEO/product/ops per the dashboard.
- Known limitation to address early: the running scheduler loads jobs only at
  startup — jobs created via the web UI take effect on the next restart.

## Notes
- Go: C:\Users\RD16019\go-toolchain\go\bin\go.exe
- go.mod: modernc.org/sqlite v1.50.1; go directive 1.25.0.
- Hooks are PowerShell; .claude/settings.json not wired (classifier-blocked; needs user auth).
- Build loop: TDD -> independent evaluator -> flip contract -> commit checkpoint.
- Known follow-ups (non-blocking, from evaluators): per-job execution timeout
  (executor; notify.Webhook also has no timeout, relies on ctx); run duration
  is coarse; dashboard JS fillLast/showRuns lack try/catch on the /runs fetch;
  a shared status constant instead of "failure"/"success" string literals.
- Targets: 3-month base 5k stars / stretch 10k; 100k = North Star. Star
  criteria in test-results.json are evidence-gated (GitHub API snapshot).
