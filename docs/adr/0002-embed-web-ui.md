# 2. Embed the web UI in the binary

Status: accepted

## Context

Krono has a web dashboard (HTML, CSS, JS). Shipping it as separate files would
break the "one binary — download and run" promise: the binary would need its
assets installed alongside it at a known path.

## Decision

Embed the dashboard (`internal/web/index.html`) into the binary with `go:embed`
and serve it from memory.

## Consequences

- The binary is fully self-contained — no asset files to deploy.
- The UI is a single hand-written HTML file with vanilla JS: no build step, no
  npm, no bundler. The whole build stays a plain `go build`.
- This caps UI complexity, which is acceptable for an MVP dashboard. A richer UI
  later would mean introducing a frontend build — revisit then.
- Updating the UI requires recompiling, which is fine: the UI ships with the binary.
