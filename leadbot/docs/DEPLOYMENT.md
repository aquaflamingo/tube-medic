# Deployment Guide

## Prerequisites

- Go 1.22+
- `yt-dlp` installed and on `$PATH`
- (Optional) YouTube Data API v3 key for category-based discovery

## Build

```bash
go build -o bin/ytleadbot ./cmd/ytleadbot
```

## Setup

```bash
cp .env.example .env
# Edit .env with your settings (see Configuration below)
mkdir -p data exports logs
```

## Configuration

| Variable          | Default                     | Description                                              |
| ----------------- | --------------------------- | -------------------------------------------------------- |
| `YT_API_KEY`      | —                           | YouTube Data API key (optional, enables category search) |
| `MAX_ENRICHMENTS` | 800                         | Max channels to enrich per run                           |
| `MIN_SUBS`        | 40000                       | Min subscriber count                                     |
| `MAX_SUBS`        | 500000                      | Max subscriber count                                     |
| `MIN_SCORE`       | 3                           | Min monetization score for export                        |
| `REQUEST_DELAY`   | 5                           | Seconds between yt-dlp requests                          |
| `DATA_DIR`        | `data`                      | Working data directory                                   |
| `DB_PATH`         | `data/ytleadbot.db`         | SQLite database path                                     |
| `EXPORT_DIR`      | `exports`                   | JSON export directory                                    |
| `KEYWORDS_PATH`   | `config/keywords_seed.json` | Keyword seed file                                        |
| `AGENCIES_PATH`   | `config/agencies_seed.json` | Agency seed file                                         |

## First Run

```bash
# Discover channels from keywords + categories, enrich, and export
ytleadbot --job discover
ytleadbot --job export
```

## Cron Setup

Run daily for continuous lead generation:

```cron
# Daily discover + enrich (runs ~2-4 hours depending on queue)
0 6 * * * cd /home/ytleadbot && ./bin/ytleadbot --job discover >> logs/discover.log 2>&1

# Export enriched leads to JSON
0 12 * * * cd /home/ytleadbot && ./bin/ytleadbot --job export >> logs/export.log 2>&1

# Weekly agency scan (adds 1-2 hours)
0 6 * * 1 cd /home/ytleadbot && ./bin/ytleadbot --job agency_scan >> logs/agency_scan.log 2>&1

# Monthly re-enrich stale channels
0 6 1 * * cd /home/ytleadbot && ./bin/ytleadbot --job reenrich_stale >> logs/reenrich.log 2>&1
```

## Monitoring

### Run History

```bash
sqlite3 data/ytleadbot.db "SELECT id, run_type, started_at, completed_at, \
  channels_discovered, channels_enriched, emails_found, errors \
  FROM runs ORDER BY id DESC LIMIT 10;"
```

### Budget Tracking

```bash
sqlite3 data/ytleadbot.db "SELECT * FROM rate_limit_budget ORDER BY date DESC, id;"
```

### Enriched Channels

```bash
# Count channels by status
sqlite3 data/ytleadbot.db "SELECT status, COUNT(*) FROM channels GROUP BY status;"

# Top leads by monetization score
sqlite3 data/ytleadbot.db "SELECT name, subscriber_count, email, monetization_score \
  FROM v_export_ready LIMIT 20;"
```

### Failed Jobs

```bash
sqlite3 data/ytleadbot.db "SELECT id, type, payload, retries, error, created_at \
  FROM jobs WHERE status = 'failed' ORDER BY created_at DESC LIMIT 10;"
```

## Logging

All logs are structured JSON via `log/slog`. For production deployments:

```bash
# Pipe to a log rotator
ytleadbot --job discover 2>&1 | logger -t ytleadbot
```

Or use systemd journal:

```ini
# /etc/systemd/system/ytleadbot-discover.service
[Unit]
Description=YTLeadBot - Daily Discover
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/home/ytleadbot/bin/ytleadbot --job discover
WorkingDirectory=/home/ytleadbot
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

## Troubleshooting

| Symptom                             | Likely Cause                | Fix                                                                        |
| ----------------------------------- | --------------------------- | -------------------------------------------------------------------------- |
| `yt-dlp: executable file not found` | yt-dlp not installed        | `pip install yt-dlp`                                                       |
| All categories return 0 results     | YouTube API quota exhausted | Check `rate_limit_budget` table; run without API key for keyword-only mode |
| `database is locked`                | Concurrent access           | Only one `ytleadbot` instance at a time (single-process design)            |
| Stale `running` jobs                | Previous run crashed        | Automatically reset on next startup (`ResetStaleJobs`)                     |

## Backup

```bash
# Backup the database
cp data/ytleadbot.db backups/leads-$(date +%Y%m%d).db

# Backup exports
tar czf backups/exports-$(date +%Y%m%d).tar.gz exports/
```
