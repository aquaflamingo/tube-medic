# Tube Medic — Project Guidance

## Modules

This is a Go multi-module repo using a root `go.work` workspace.

| Module | Path | Purpose |
|--------|------|---------|
| **core** (`github.com/aquaflamingo/tmcore`) | `./tube-medic-core/` | Public API: `Config`, `Report`, `RunScan`. Scans YouTube channel descriptions for broken/suspicious links. |
| **web** (`github.com/aquaflamingo/tubemedic-web`) | `./web/` | Web UI (HTMX + Tailwind). Importes `tmcore` and calls `RunScan`. |
| **leadbot** (`github.com/aquaflamingo/ytleadbot`) | `./leadbot/` | Lead-finding pipeline for YouTube creator outreach. |

## Architecture

```
core (tmcore.go)               ← public API
  ├── cmd/tube-medic/          ← CLI satellite
  └── web/                     ← web satellite (imports root package via go.work)

go.work (root)                 ← workspace tying all three modules together
```

## Recent Changes

### Web now uses core's public API

The `web/internal/handlers/handlers.go` was rewired to import `github.com/aquaflamingo/tmcore`
instead of the old `github.com/aquaflamingo/tubemedicmvp/pkg/youtube`.

**Before:** `HandleScan` only did a channel resolve (partial).
**After:** `HandleScan` builds `tmcore.Config`, calls `tmcore.RunScan()`, and renders the full
report (channel info, stat cards, broken links, revenue-critical links, quota usage).

Key changes:
- Created root `go.work` replacing stale `web/go.work` (which referenced non-existent paths)
- `web/go.mod` now depends on `github.com/aquaflamingo/tmcore` (resolved locally via workspace)
- Added optional `max_videos` form field to the web UI (default 50, range 1-200)
- No API changes needed to core — `Config`, `Report`, `RunScan` already covered web's needs

## Commands

```sh
# build, test, vet all modules (workspace-wide)
go build ./...
go test ./...
go vet ./...

# or per-module
cd web && go build ./cmd/server/
cd tube-medic-core && go test ./...
```
