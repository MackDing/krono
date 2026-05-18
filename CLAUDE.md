# Krono — project context & long-running conventions

Krono is a self-hosted job scheduler with a web dashboard: a single Go binary
(embedded SQLite + embedded web UI) that runs shell commands, HTTP requests,
and AI-agent jobs on a schedule, with full run observability. See `README.md`.

Krono is built and grown by a long-running, self-iterating agent loop. These
conventions (adapted from anthropics/cwc-long-running-agents) keep that loop
honest across many fresh sessions.

## Always start here
Before anything else, read `PROGRESS.md` — the handoff note from the previous
session. Then `git log --oneline -10` to see what was just committed, and
`go build ./...` once to confirm the tree builds (not a broken handoff).

## One item per session
Work exactly one item from `PROGRESS.md`. Finish it — code written, `go build
./...` clean, evidence captured, evaluator returns PASS — before starting
another. New tasks mid-session go into `PROGRESS.md`; finish the current one first.

## Proof before passing
A `test-results.json` row flips to `"passes": true` only after you have:
1. produced real evidence into `evidence/` (a `go build` log, `go test` output,
   or a UI screenshot),
2. opened that evidence with the Read tool,
3. confirmed it shows what it should, and
4. had the `evaluator` subagent review the work and return `PASS`.
The `verify-gate` hook denies writes to `test-results.json` until evidence has
been Read. Do not work around it.

## Targets are evidence-gated
`test-results.json` also carries growth criteria (stars, launch). They start
`false` and flip only on a verified GitHub-API snapshot. Aim as high as you
want — the contract guarantees the reported number is always the real one.

## Keep PROGRESS.md current
After each completed item: check off what is done, record what you learned,
note what is next. Future sessions read this file cold.

## Commit often
Commit at meaningful checkpoints with descriptive messages; `git add` new files
yourself. The `Stop` hook commits whatever tracked work is left at session end.

## Operator control
`OPERATOR STEERING:` messages come from the human via `STEER.md` — higher
priority than your current plan. If `AGENT_STOP` exists, you are halted.
