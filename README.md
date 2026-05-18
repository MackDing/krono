# Krono

> A modern, self-hosted job scheduler — cron with a web dashboard.
> Schedule shell commands, HTTP requests, and AI agents from a single Go binary.

**🚧 Status: early development.** The MVP is being built in the open — Krono is **not usable yet**. ⭐ Star and 👀 watch to follow the first release.

## What is Krono?

Krono is an open-source, self-hosted scheduler for recurring jobs — the cron you always wanted.

Classic `cron` has no memory, no visibility, and no UI: a job fails silently or the box reboots, and you are debugging blind. Krono fixes that:

- **Persistent** — jobs and run history survive restarts (embedded SQLite, no database to operate).
- **Observable** — every run captures exit code, duration, and full output. No more silent failures.
- **Visual** — manage everything from a clean web dashboard. No more editing crontabs over SSH.
- **Simple to self-host** — one static binary, zero external dependencies.
- **Runs anything** — shell commands, HTTP requests, and **AI agents**, on any schedule.

## Schedule your AI agents

AI agents are good at doing work and bad at remembering to. Agent frameworks (CrewAI, LangGraph, and friends) ship no scheduler of their own. Krono is the missing piece — point it at an agent run, give it a schedule, and get retries, run history, and failure alerts for free:

```
Cron:  0 8 * * 1-5                               # every weekday at 08:00
Run:   python research_agent.py --topic "AI news"
```

## Features

**MVP — in progress**
- [ ] Cron-expression and interval schedules
- [ ] Persistent jobs & run history (embedded SQLite)
- [ ] Job types: shell command, HTTP request
- [ ] Web dashboard — create/edit jobs, run history, live logs
- [ ] Run observability — exit code, duration, captured output
- [ ] Single static binary — download and run
- [ ] Failure notifications (webhook)

**Roadmap**
- [ ] Distributed mode — no double-firing across multiple instances
- [ ] Container job type
- [ ] Job dependencies / DAGs
- [ ] Native AI-agent job type with token & cost tracking
- [ ] More notification channels — Slack, Discord, email
- [ ] Multi-user access & authentication
- [ ] One-file `docker-compose` deployment

## Quick start

> ⚠️ Pre-release. Building from source today gives you the dev scaffold only — the scheduler and web UI land with the MVP. ⭐ Watch the repo for the release.

```bash
git clone https://github.com/krono-sh/krono.git
cd krono
go build -o krono ./cmd/krono
./krono --version
```

## Contributing

Krono is in early development — ideas, issues, and pull requests are welcome. The single most useful thing right now: ⭐ star the repo, and open an issue describing your worst scheduling pain. That feedback shapes the MVP.

## License

[MIT](LICENSE) © Krono Authors
