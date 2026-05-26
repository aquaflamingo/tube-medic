# YTLeadBot — Project Guidance

A Go-based lead-finding pipeline for YouTube creator outreach. Discovers channels (40K–500K subs),
enriches with monetization scoring, and exports to JSON/SQLite.

## Quick Start

```bash
cp .env.example .env      # edit if needed (YT_API_KEY recommended)
ytleadbot --job discover  # keyword search → queue → enrich
ytleadbot --job export    # dump enriched leads to JSON
```

## Build

```bash
make build   # native build (auto-downloads & embeds yt-dlp)
```

- No external dependencies — yt-dlp is embedded in the binary (all platforms)
- No API keys needed (SponsorBlock is free, keyword-only mode)
- YouTube Data API key (optional, enables category-based discovery)
- Cross-compile: replace the embedded yt-dlp with the target platform's binary, then set `GOOS`/`GOARCH` (see README)

## Make Targets

| Target | Description |
|---|---|
| `make build` | Download yt-dlp (if missing) + build binary |
| `make test` | Run all tests |
| `make clean` | Remove binary, DB, exports |
| `make distclean` | Above + remove embed + Go cache |
| `make download-yt-dlp` | Force refresh embedded yt-dlp |
| `make run-discover` | Build + run discover |
| `make run-enrich` | Build + run enrich |
| `make run-export` | Build + run export |
| `make run-agency-scan` | Build + run agency scan |
| `make run-reenrich` | Build + run re-enrich stale |

## Directory Structure

```
cmd/ytleadbot/main.go         # CLI entrypoint
internal/config/config.go     # .env loading, JSON seed files
internal/core/db.go           # SQLite schema + DAL (channels, jobs, runs, rate_limit_budget)
internal/modules/
  enrichment.go               # Single-channel enrich pipeline
  discovery.go                # Keyword + category + related channel discovery
  orchestrator.go             # Job dispatch loop with retry
  scheduler.go                # discover/reenrich_stale pipelines
  agency_finder.go            # Agency roster scrape + detect → queue
  exporter.go                 # JSON export from v_export_ready
  runlog.go                   # Run tracking helpers
internal/utils/
  ytdlp.go                    # yt-dlp subprocess wrapper
  ytdlp_embed.go              # Embedded static binary extraction
  ytdlp_static/               # Pre-downloaded yt-dlp binary (for //go:embed)
  youtubeapi.go               # YouTube Data API v3 client (category search, channel details)
  scraper.go                  # HTTP + goquery helpers
  email_extractor.go          # Regex email + page scraping
  monetization.go             # Scoring + SponsorBlock
config/
  keywords_seed.json           # Keyword array + niche_categories map
  agencies_seed.json           # Agency seed list
docs/
  DESIGN.md                    # Architecture design doc
  DEPLOYMENT.md                # Deployment + cron + monitoring guide
```

## Job Types

| CLI Flag | What It Does |
|---|---|
| `--job enrich [--channel @h]` | Process enrich queue or single channel |
| `--job discover` | Keyword search + category search → queue → enrich → related channels → enrich |
| `--job export` | Export `v_export_ready` to JSON |
| `--job agency_scan` | Agency roster scrape + email domain clustering + description regex → queue → enrich |
| `--job reenrich_stale` | Re-enrich channels last done >30d ago |

## Architecture

Work-queue model. Every stage writes jobs to SQLite; workers pick them up independently. Both `discover` and `agency_scan` feed into the same queue and can safely run in parallel (idempotent upserts).

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

## Key Conventions

- Go std `log/slog` for logging (structured JSON)
- `modernc.org/sqlite` (pure Go, no CGO)
- Config via `.env` + JSON seed files with built-in fallbacks
- Random delay 3-8s between yt-dlp calls
- Failed jobs retry up to 3×; stale `running` jobs reset on startup
- Every pipeline run logged to `runs` table with counters (discovered, enriched, emails, errors)
- Daily budget tracking in `rate_limit_budget` table (YouTube API enforced, yt-dlp monitored)

## Progress

### Slice 1 — Core Pipeline ✅
- [x] SQLite schema + DAL (channels, jobs, runs, agencies, agency_channels)
- [x] Config loader (.env + JSON seed files with fallbacks)
- [x] yt-dlp subprocess wrapper (channel info, recent videos, keyword search)
- [x] Scraper (YouTube IDs from page, pricing, social links)
- [x] Email extraction (regex, website scraping, Linktree)
- [x] Monetization scoring + SponsorBlock checks
- [x] Single-channel enrichment pipeline
- [x] JSON exporter from `v_export_ready`
- [x] CLI entrypoint with all job types

### Slice 2 — Discovery at Scale ✅
- [x] Keyword search via yt-dlp → queue enrich jobs
- [x] YouTube API category search (`SearchByCategory`, `GetChannelDetails`)
- [x] Quota tracking (in-memory + persistent `rate_limit_budget`)
- [x] Related channel discovery from recently enriched channels (scraper)
- [x] `RunDiscoverPipeline` chains all 4 discovery strategies
- [x] Niche category IDs in `config/keywords_seed.json`

### Slice 3 — Agency Finder ✅
- [x] Agency roster scraping (seed list → roster page → YouTube IDs)
- [x] Agency channel search via yt-dlp keywords
- [x] Email domain clustering (≥3 channels with same domain → agency)
- [x] Description regex agency detection
- [x] Agency upsert + channel linking + evidence tracking

### Slice 4 — Hardening ✅
- [x] `rate_limit_budget` table + `ConsumeBudget`/`BudgetRemaining` methods
- [x] YouTube API budget enforcement (100 units/call, 10K/day)
- [x] yt-dlp budget tracking (1 per enrichment, limit from `MaxEnrichments`)
- [x] Run logging wired into discover, agency_scan, enrich, export pipelines
- [x] `docs/DEPLOYMENT.md` (cron, monitoring, troubleshooting)
- [x] `BudgetRemaining` returns -1 for untracked budgets (no row, no limit)

### Slice 5 — Embed + Cross-Compile ✅
- [x] yt-dlp embedded via `//go:embed` (freshness check: re-extracts if cached size differs)
- [x] `ensureYTDLPExtracted()` runs on all platforms (not just Linux) — `~/.cache/ytleadbot/yt-dlp`
- [x] Native `make build` downloads + embeds yt-dlp for current platform
- [x] Cross-compile: replace `internal/utils/ytdlp_static/yt-dlp` + set `GOOS`/`GOARCH`

### Slice 6 — Logging & Makefile Cleanup ✅
- [x] Per-channel logging in discovery pipeline (keyword results count, queued channel IDs)
- [x] DB write logging ("channel saved to database" after upsert)
- [x] Job processing logged at Info level (was Debug)
- [x] Simplified Makefile (48 lines, no stamp complexity)
- [x] ASCII architecture diagram in README
- [x] Fixed DB path references (`data/leads.db` → `data/ytleadbot.db`)
