# Krono

> A modern, self-hosted job scheduler — cron with a web dashboard.
> Schedule shell commands, HTTP requests, and AI agents from a single Go binary.

Krono is an open-source, self-hosted scheduler for recurring jobs — the cron you
always wanted. Classic `cron` has no memory, no visibility, and no UI: a job
fails silently or the box reboots, and you are debugging blind. Krono fixes that.

- **Persistent** — jobs and run history live in an embedded SQLite database and
  survive restarts. No database to operate.
- **Observable** — every run records its exit code, duration, and full output.
  No more silent failures.
- **Visual** — a built-in web dashboard to create, edit, and delete jobs and
  browse run history. No more editing crontabs over SSH.
- **Runs anything** — shell commands and HTTP requests; point a job at an
  AI-agent run just as easily.
- **One binary** — a single static Go binary with the web UI embedded. No
  runtime, no external dependencies. Download and run.

## Schedule your AI agents

AI agents are good at doing work and bad at remembering to. Agent frameworks
(CrewAI, LangGraph, and friends) ship no scheduler of their own. Krono is the
missing piece — point a job at an agent run, give it a schedule, and get run
history and failure webhooks for free:

```
name:      overnight-research
schedule:  0 8 * * 1-5           # every weekday at 08:00
command:   python research_agent.py --topic "AI news"
```

## Quick start

Krono builds from source today (prebuilt release binaries are coming):

```bash
git clone https://github.com/krono-sh/krono.git
cd krono
go build -o krono ./cmd/krono
./krono
```

Then open the dashboard at <http://localhost:8400>. On first run Krono seeds a
couple of demo jobs so you can see it working right away.

```bash
./krono --addr :8400 --db krono.db --notify-url https://hooks.example.com/krono
```

## Features

- Cron-expression and interval schedules — `0 8 * * 1-5`, `@every 30s`, `@daily`
- Persistent jobs and run history (embedded SQLite)
- Shell-command and HTTP-request jobs — exit code, duration, and output captured
- A web dashboard: create / edit / delete jobs and browse per-job run history
- Webhook notification when a run fails
- Ships as one self-contained static binary

## Roadmap

- Live job reload — apply dashboard changes without restarting
- Distributed mode — no double-firing across multiple instances
- Container job type; job dependencies / DAGs
- A native AI-agent job type with token and cost tracking
- More notification channels — Slack, Discord, email
- Multi-user access and authentication
- Prebuilt release binaries and a one-file `docker-compose`

## Status

Krono is an early but working MVP: the scheduler, persistence, job execution,
web dashboard, and failure webhooks are all built and tested. Expect rough
edges. Issues and pull requests are welcome — and ⭐ stars help.

## License

[MIT](LICENSE) © Krono Authors
