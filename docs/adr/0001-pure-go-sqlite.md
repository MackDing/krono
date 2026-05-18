# 1. Pure-Go SQLite for persistence

Status: accepted

## Context

Krono needs durable storage for jobs and run history, and its core promise is
"one self-contained binary — download and run". The common Go SQLite driver,
`mattn/go-sqlite3`, requires cgo: a C toolchain to build, and platform-specific
binaries that are harder to cross-compile and distribute.

## Decision

Use `modernc.org/sqlite`, a pure-Go SQLite implementation. No cgo.

## Consequences

- `go build` produces a single static binary on any platform with no C
  toolchain; cross-compilation is trivial.
- Slightly larger binary and marginally slower than the cgo driver — acceptable
  for a scheduler's low query volume.
- The `database/sql` API is unchanged, so the rest of the code stays
  driver-agnostic.
- The race detector (`go test -race`) needs cgo, so it runs in CI on a Linux
  runner rather than on every developer machine.
