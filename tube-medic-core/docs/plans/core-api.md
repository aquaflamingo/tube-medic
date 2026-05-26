# Core API Refactor: Core + Satellites

## Goal

Refactor from a CLI-only binary to a "core + satellite" pattern:
- **Core:** root-level public Go package (`github.com/aqfl/tmcore`) — exports `Config`, `Report`, `RunScan`
- **Satellite 1:** CLI (`cmd/tube-medic/`) — thin wrapper around core + terminal output
- **Satellite 2 (future):** Web module — imports core package for programmatic access

## Why

- The web module (`github.com/aqfl/tubemedic-web`) needs to invoke the scan pipeline programmatically without shelling out
- CLI-specific concerns (flag parsing, terminal colors, file output) should not leak into the public API
- Clear dependency direction: packages import core, not the other way

## Changes

### 1. New: `tmcore.go` (module root)

```go
package tmcore

type Config struct {
    APIKey     string
    ChannelURL string
    MaxVideos  int
}

type Report struct {
    Channel  *youtube.Channel
    Videos   []youtube.Video
    Results  []checker.CheckResult
    Summary  checker.Summary
    Quota    youtube.Quota
}

func RunScan(cfg Config) (*Report, error)
```

- Orchestrates: FetchChannel → ExtractAll → CheckAll → Summarize
- Returns structured data, performs no I/O beyond the scan itself
- Only export what an API consumer needs (no CLI flags, no terminal output)

### 2. Refactor: `internal/config/`

- `Config` struct moves to root package (public)
- `Load()` returns `(*tmcore.Config, outputFile string, error)`
- `OutputFile` stays CLI-only (not in public `Config`)
- Still uses `flag` package and `.env` loading

### 3. Thin: `cmd/tube-medic/main.go`

```
config.Load() → flags + env
tmcore.RunScan(cfg) → *Report
report.Print(w, ...) → terminal output
```

- No longer imports internal/ packages directly (except `config` and `report`)
- Reporting/formatting stays in `internal/report` (CLI-only)

### 4. Unchanged

- `internal/checker/`, `internal/classifier/`, `internal/scraper/`, `internal/youtube/` — no changes
- `internal/report/` — unchanged (terminal formatting, CLI-specific)

## Dependency Graph

```
tmcore (root)                     ← public API
  ├─ internal/checker
  ├─ internal/scraper
  └─ internal/youtube

cmd/tube-medic                          ← CLI satellite
  ├─ tmcore (root)
  ├─ internal/config                    ← CLI flag/env parsing
  └─ internal/report                    ← terminal formatting

tubemedic-web (external)                ← web satellite
  └─ tmcore (root)
```
