# YTLeadBot

YouTube creator lead finder. Discovers channels (40K–500K subs), enriches with monetization scoring and emails, and exports to JSON.

## Quick Start

```bash
cp .env.example .env
make build                        # auto-downloads & embeds yt-dlp
./bin/ytleadbot --job discover    # keyword search → enrich
./bin/ytleadbot --job export      # dump leads to JSON
```

## Requirements

- Go 1.22+ (to build)
- YouTube Data API key (optional, for category search — set `YT_API_KEY` in `.env`)

No external dependencies at runtime — yt-dlp is embedded in the binary.

## Usage

```
ytleadbot --job discover                    Search keywords, queue & enrich channels
ytleadbot --job enrich [--channel @handle]  Process enrichment queue or single channel
ytleadbot --job export                      Export enriched leads to JSON
ytleadbot --job agency_scan                 Scan agency rosters + detect agencies
ytleadbot --job reenrich_stale              Re-enrich channels last enriched >30d ago
```

## Make Targets

| Target | Description |
|---|---|
| `make build` | Download yt-dlp (if missing) + build native binary |
| `make test` | Run all tests |
| `make clean` | Remove binary, DB, export files |
| `make distclean` | Above + remove embedded yt-dlp + Go build cache |
| `make download-yt-dlp` | Download/refresh embedded yt-dlp for current OS |
| `make run-discover` | Build + run discover job |

```
  discover ──┬── keyword search ──┐
             └── category search ─┤
                                  ├──→ enrich_channel queue ──→ enrichment ──→ channels table ──→ v_export_ready ──→ export
                                  │
  agency_scan ─┬── roster scrape ─┤
               ├── domain cluster ┤
               └── desc regex ────┘

  enrichment also crawls related channels from the channels table and re-queues them.
```

Agency scan and discover are safe to run in parallel — both just queue `enrich_channel` jobs and the `channels` table uses idempotent upserts. SQLite handles concurrent access with busy retries.

## Cross-Compile

Build for macOS ARM64 from Linux:
```bash
curl -sL -o internal/utils/ytdlp_static/yt-dlp \
  https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos && chmod +x $_
GOOS=darwin GOARCH=arm64 go build -o bin/ytleadbot-darwin-arm64 ./cmd/ytleadbot
```

## Data

Leads land in `data/ytleadbot.db` (SQLite). JSON exports go to `exports/`. Query directly:

```bash
sqlite3 data/ytleadbot.db "SELECT name, subscriber_count, email, monetization_score FROM v_export_ready;"
```
