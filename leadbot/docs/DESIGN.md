# YTLeadBot — Design Document

**YouTube "Middle Class" Creator & Agency Lead Finder**
Version 1.0 | May 2026

---

## 1. Purpose & Problem Statement

The creator economy has a largely untapped "middle class" — YouTube channels with 40K–500K subscribers. They're large enough to have a real audience and be actively monetizing, but small enough that outreach is personal and conversion is high. Finding them at scale, with contact info, is currently either expensive (paid tools like Modash, Heepsy) or manual.

YTLeadBot solves this by running an autonomous agentic pipeline — no LLM required — that discovers, enriches, scores, and stores these leads into a local SQLite database, ready for downstream outreach pipelines.

**Secondary objective:** Identify creator talent agencies and map their rosters, since winning an agency relationship unlocks many creators at once.

---

## 2. Goals

| Goal                                                              | Priority |
| ----------------------------------------------------------------- | -------- |
| Find YouTube channels with 40K–500K subscribers                   | P0       |
| Extract: name, channel URL, sub count, email, niche               | P0       |
| Score channels by monetization activity                           | P0       |
| Store leads in SQLite with dedup                                  | P0       |
| Run free (no paid APIs)                                           | P0       |
| Run on cron unattended                                            | P0       |
| Identify creator agencies + their talent rosters                  | P1       |
| Export leads in a format compatible with downstream outreach tool | P1       |

---

## 3. Non-Goals

- Real-time lead delivery (batch is fine)
- LLM-based content analysis (heuristic scoring is sufficient)
- Paid API integrations (Hunter.io, Modash, etc.)
- UI or dashboard (SQLite is the interface)

---

## 4. System Architecture

```
┌────────────────────────────────────────────────────────────┐
│                      CRON SCHEDULER                         │
│          (system cron calling compiled binary)               │
└───────────────────────────┬────────────────────────────────┘
                            │
┌───────────────────────────▼────────────────────────────────┐
│                    ORCHESTRATOR                              │
│  - Reads/writes job queue in SQLite                         │
│  - Dispatches tasks to workers                              │
│  - Tracks state, retries, run logs                          │
└───┬───────────────┬───────────────┬──────────────┬──────────┘
    │               │               │              │
    ▼               ▼               ▼              ▼
DISCOVERY      ENRICHMENT      AGENCY          EXPORT
MODULE         MODULE          FINDER          MODULE
(find IDs)     (get details)   MODULE          (to pipeline)
```

The system is a **work queue + worker** model. Every stage writes its outputs as jobs into SQLite. Workers pick up jobs independently. This makes the system resumable, idempotent, and easy to debug.

---

## 5. Data Sources

All sources are free.

### 5.1 `yt-dlp` (Primary Workhorse)

`yt-dlp` is called as a subprocess via Go's `os/exec`. Originally a YouTube downloader, it has evolved into a full YouTube data extraction engine with no API key required.

Key capabilities used:

- Search YouTube by keyword, returning channel results
- Extract full channel metadata: name, description, subscriber count, view count, upload schedule, country
- **Extract business inquiry email from the About tab** — YouTube rot13-obfuscates these, yt-dlp decodes them transparently
- Extract all linked URLs from the channel About page (Patreon, Linktree, Gumroad, etc.)

```bash
# Example: extract channel metadata + email
yt-dlp --dump-json --flat-playlist "https://www.youtube.com/@mkbhd"

# Example: search and get channel IDs
yt-dlp "ytsearch20:fitness creator" --get-id --no-playlist
```

### 5.2 YouTube Data API v3 (Supplemental)

Free tier: **10,000 units/day**.

| Operation                            | Cost      | Daily budget (~units)        |
| ------------------------------------ | --------- | ---------------------------- |
| `search.list` (keyword search)       | 100 units | ~30 searches = 3,000 units   |
| `channels.list` (channel details)    | 1 unit    | ~5,000 lookups = 5,000 units |
| `videos.list` (upload cadence check) | 1 unit    | ~2,000 checks = 2,000 units  |

Accessed via `google.golang.org/api/youtube/v3`. Used for structured search by `categoryId` and bulk channel stats lookups.

### 5.3 SponsorBlock API (Monetization Signal)

Open, free, no key required. Returns whether a video has sponsor segments — a direct signal that a creator is taking paid sponsorships.

```
GET https://sponsor.ajay.app/api/searchSegments?videoID={id}&categories=["sponsor"]
```

### 5.4 Agency Website Scraping

A curated seed list of creator agency websites is scraped for their talent rosters, extracting linked YouTube channels via HTTP + `goquery` HTML parsing.

Initial seed list includes: Whalar, Viral Nation, Gleam Futures, Studio71, Fullscreen, Night Media, Select Management, and ~20 others.

---

## 6. Modules

### 6.1 Discovery Module

**Job:** Produce a list of candidate YouTube channel IDs to enrich.

**Inputs:** Keyword seed lists + category IDs
**Output:** Rows inserted into `jobs` table with type `enrich_channel`

**Strategy:**

1. **Keyword search via yt-dlp** — iterates over a seed keyword matrix (niche × intent):
   - Niches: fitness, personal finance, business, tech, cooking, travel, gaming, self-improvement, parenting, beauty
   - Intents: "course", "coaching", "how I make money", "my business", "sponsor", "collab"

2. **YouTube API category search** — queries `search.list` by `type=channel` and `videoCategoryId`

3. **"Related channels" crawl** — yt-dlp surfaces related channels from sidebar data

4. **Agency roster back-fill** — channels found by the Agency Finder Module

Discovered IDs are deduplicated against the `channels` table before being queued.

### 6.2 Enrichment Module

**Job:** For each queued channel ID, fetch full metadata and compute a monetization score.

**Inputs:** `jobs` rows of type `enrich_channel`
**Output:** Upserted rows in `channels` table

**Steps per channel:**

1. **Fetch via yt-dlp** — subscriber count, description, about links, email, country, upload history
2. **Apply sub count filter** — discard if outside 40K–500K window
3. **Compute monetization score** — see Section 7
4. **Discard if score < threshold** (configurable, default: 3)
5. **Enrich email** — if no email on About page, scrape linked website's `/contact` page
6. **Write to `channels` table**

**Rate limiting:** Random delay 3–8 seconds between yt-dlp calls. Exponential backoff on errors. Daily cap: ~800 channel enrichments.

### 6.3 Agency Finder Module

**Job:** Discover creator talent agencies and map their rosters back to the `channels` table.

**Strategy A — Known Agency Roster Scraping:**

- Iterate seed agency website list
- Use `net/http` + `goquery` to parse `/talent`, `/roster`, `/creators` pages
- Extract all YouTube channel links → queue for enrichment

**Strategy B — YouTube Search for Agencies:**

- Search terms: `"talent management"`, `"creator agency"`, `"we represent"`, `"MCN"`
- Filter by channel description keywords: "management", "agency", "roster", "talent"

**Strategy C — Shared Email Domain Detection:**

- Cluster channels sharing the same contact email domain
- Clusters of 3+ channels = likely agency-managed group

**Strategy D — Description Parsing:**

- Regex scan channel descriptions for patterns: `"managed by"`, `"represented by"`, `"booking via"`, `"talent:"`

### 6.4 Export Module

**Job:** Surface "ready" leads to the downstream outreach pipeline.

**Output formats:**

- SQLite view `v_export_ready` (primary) — downstream pipeline queries this directly
- JSON file dump — written to `exports/leads_YYYYMMDD.json`

**Export schema (maps to downstream pipeline's expected input):**

```json
{
  "channel_id": "UCxxxxxx",
  "name": "Graham Stephan",
  "handle": "@GrahamStephan",
  "email": "graham@example.com",
  "subscriber_count": 480000,
  "niche": "personal_finance",
  "monetization_score": 8,
  "monetization_signals": ["sponsorships", "course_link", "email_present"],
  "website": "https://grahamstephan.com",
  "agency": null,
  "agency_id": null,
  "country": "US",
  "channel_url": "https://youtube.com/@GrahamStephan",
  "status": "new"
}
```

Leads are marked `status = 'exported'` after export to avoid re-processing.

---

## 7. Monetization Scoring

Each channel receives a score. Only channels with score ≥ 3 are stored as leads.

| Signal                                                          | Score | Detection Method                                             |
| --------------------------------------------------------------- | ----- | ------------------------------------------------------------ |
| Business email present on About                                 | +3    | yt-dlp extracts directly                                     |
| SponsorBlock: has sponsor segments                              | +3    | SponsorBlock API                                             |
| Patreon / Ko-fi / Buy Me a Coffee link                          | +2    | About page link regex                                        |
| Course platform link (Teachable, Kajabi, Gumroad, Podia, Skool) | +2    | About page link regex                                        |
| Shopify / Teespring / merch link                                | +1    | About page link regex                                        |
| Amazon storefront link                                          | +1    | About page link regex                                        |
| Linktree/bio link (proxied monetization)                        | +1    | About page, then scrape the Linktree                         |
| Upload frequency ≥ 2 videos/month (last 6 months)               | +1    | yt-dlp video list                                            |
| Channel membership detected                                     | +1    | HTML scrape of channel page                                  |
| Website with checkout/pricing page                              | +1    | Scrape linked website, check for `/pricing`, `/shop`, `/buy` |

Maximum possible score: 16. Threshold for inclusion: **3**.

Channels scoring 8+ are flagged `monetization_tier = 'high'` for prioritized outreach.

---

## 8. SQLite Schema

```sql
-- Core leads table
CREATE TABLE channels (
  id                    TEXT PRIMARY KEY,   -- YouTube channel ID
  handle                TEXT,
  name                  TEXT,
  subscriber_count      INTEGER,
  view_count            INTEGER,
  video_count           INTEGER,
  upload_frequency      REAL,               -- avg videos/month, last 6mo
  niche                 TEXT,               -- inferred from discovery keyword
  description           TEXT,
  country               TEXT,
  email                 TEXT,
  website               TEXT,
  social_links          TEXT,               -- JSON array of {platform, url}
  monetization_score    INTEGER DEFAULT 0,
  monetization_tier     TEXT,               -- 'high' | 'standard' | null
  monetization_signals  TEXT,               -- JSON array of signal names
  agency_id             TEXT REFERENCES agencies(id),
  status                TEXT DEFAULT 'new', -- new|enriched|exported|contacted|dead
  discovery_keyword     TEXT,               -- what search found this channel
  first_seen_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  enriched_at           TIMESTAMP,
  exported_at           TIMESTAMP,
  raw_yt_metadata       TEXT                -- full JSON blob from yt-dlp
);

-- Agency table
CREATE TABLE agencies (
  id            TEXT PRIMARY KEY,           -- domain or slug
  name          TEXT,
  website       TEXT,
  email         TEXT,
  roster_count  INTEGER DEFAULT 0,
  detection_method TEXT,                    -- 'seed'|'shared_email'|'description'|'search'
  created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Many-to-many: agencies <-> channels
CREATE TABLE agency_channels (
  agency_id     TEXT REFERENCES agencies(id),
  channel_id    TEXT REFERENCES channels(id),
  confidence    REAL DEFAULT 1.0,           -- 0.0–1.0
  evidence      TEXT,                       -- how we linked them
  PRIMARY KEY (agency_id, channel_id)
);

-- Job queue (orchestrator state)
CREATE TABLE jobs (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  type        TEXT NOT NULL,                -- 'discover'|'enrich_channel'|'agency_scan'|'export'
  payload     TEXT,                         -- JSON
  status      TEXT DEFAULT 'pending',       -- pending|running|done|failed
  retries     INTEGER DEFAULT 0,
  error       TEXT,
  created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMP
);

-- Run log for observability
CREATE TABLE runs (
  id                  INTEGER PRIMARY KEY AUTOINCREMENT,
  run_type            TEXT,
  started_at          TIMESTAMP,
  completed_at        TIMESTAMP,
  channels_discovered INTEGER DEFAULT 0,
  channels_enriched   INTEGER DEFAULT 0,
  emails_found        INTEGER DEFAULT 0,
  agencies_found      INTEGER DEFAULT 0,
  leads_exported      INTEGER DEFAULT 0,
  errors              INTEGER DEFAULT 0
);

-- Useful views
CREATE VIEW v_export_ready AS
  SELECT * FROM channels
  WHERE status = 'enriched'
    AND monetization_score >= 3
    AND email IS NOT NULL
  ORDER BY monetization_score DESC;

CREATE VIEW v_agency_rosters AS
  SELECT a.name AS agency_name, a.website, c.name AS creator_name,
         c.subscriber_count, c.email, c.monetization_score
  FROM agencies a
  JOIN agency_channels ac ON ac.agency_id = a.id
  JOIN channels c ON c.id = ac.channel_id
  ORDER BY a.name, c.subscriber_count DESC;
```

---

## 9. Tech Stack

| Component     | Tool                                          | Notes                              |
| ------------- | --------------------------------------------- | ---------------------------------- |
| Language      | Go 1.22+                                      | Compiled single binary             |
| Module mgmt   | Go modules (`go.mod`)                         | Standard                           |
| YouTube data  | `yt-dlp` via `os/exec`                        | Subprocess, no API key needed      |
| YouTube API   | `google.golang.org/api/youtube/v3`            | Free 10K units/day                 |
| Web scraping  | `net/http` + `github.com/PuerkitoBio/goquery` | jQuery-style HTML parser           |
| HTML parsing  | `golang.org/x/net/html`                       | Standard library                   |
| Database      | `modernc.org/sqlite`                          | Pure Go SQLite, no CGO             |
| Scheduling    | System cron calling compiled binary           | `/usr/local/bin/ytleadbot --job X` |
| Retry/backoff | `github.com/cenkalti/backoff/v4`              | Exponential backoff                |
| Config        | `.env` via `github.com/joho/godotenv`         | API keys, thresholds               |
| Logging       | `log/slog` (standard library)                 | Structured JSON logs               |
| CLI           | `flag` (standard library)                     | No external CLI framework          |

**Zero recurring cost. Single binary deployment.** No runtime dependencies beyond `yt-dlp` on `$PATH`.

---

## 10. Cron Schedule

```
# system crontab
# Discovery: find new channel IDs (daily, 1am)
0 1 * * * /usr/local/bin/ytleadbot --job discover

# Enrichment: process queued channels (daily, 2am)
0 2 * * * /usr/local/bin/ytleadbot --job enrich

# Agency scan (weekly, Sunday 3am)
0 3 * * 0 /usr/local/bin/ytleadbot --job agency_scan

# Export new leads to outreach pipeline (daily, 4am)
0 4 * * * /usr/local/bin/ytleadbot --job export

# Re-enrich stale leads (monthly, 1st of month)
0 5 1 * * /usr/local/bin/ytleadbot --job reenrich_stale
```

---

## 11. Agent Loop (Orchestrator)

The orchestrator is a lightweight "agent" that runs without an LLM — it follows a deterministic task graph based on queue state.

```
on_startup:
  check for stale 'running' jobs → reset to 'pending'

main_loop:
  while jobs_pending():
    job = dequeue_oldest_pending()
    mark job 'running'

    try:
      result = dispatch(job.type, job.payload)
      mark job 'done'
      enqueue_downstream_jobs(result)
    except RateLimitError:
      sleep_with_backoff()
      requeue(job)
    except PermanentError:
      mark job 'failed', log error
    except Exception:
      if job.retries < MAX_RETRIES:
        requeue(job, retries+1)
      else:
        mark job 'failed'
```

This makes the system **resumable** — if it crashes mid-run, restarting picks up exactly where it left off. Jobs in `running` state at startup are evidence of a crash and are reset.

---

## 12. Email Extraction Strategy

Email finding is prioritized in order of reliability:

1. **yt-dlp About tab** — highest quality. yt-dlp handles YouTube's rot13 email obfuscation and surfaces the "For business inquiries" email directly in metadata. ~40–60% of monetizing creators have this set.

2. **Linked website contact scraping** — for channels with a personal website, scrape `/contact`, `/about`, `/work-with-me` pages for email addresses using regex: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`. Filter out noreply@, info@ patterns where possible.

3. **Linktree / bio link scraping** — many creators use Linktree, Beacons, or Koji. These pages often surface an email contact button. Scrape the rendered page for `mailto:` links.

4. **No email found** — channel is still stored with `email = NULL` and `status = 'enriched_no_email'`. These can be re-enriched later or passed to the downstream pipeline for manual review.

---

## 13. Agency Finder — Detailed Flow

```
1. Load seed agency list from config/agencies_seed.json
2. For each agency:
   a. Fetch their /talent or /roster page
   b. Extract all youtube.com/ and youtube.com/c/ URLs
   c. Queue each discovered channel for enrichment
   d. Upsert agency record

3. Run YouTube search for agency-type channels:
   Search terms:
   - "creator talent agency"
   - "YouTube MCN management"
   - "influencer management company"
   Filter: channel description contains ["management", "agency", "roster", "booking"]

4. After enrichment batch completes:
   a. Group channels by email domain
   b. For groups of 3+: create agency record, link channels

5. Regex scan all enriched channel descriptions:
   Patterns: "managed by", "represented by", "booking via", "talent: "
   Extract agency name → search web for their website → scrape roster
```

---

## 14. Handling Rate Limits & Politeness

| Source               | Rate Limit Strategy                                                    |
| -------------------- | ---------------------------------------------------------------------- |
| yt-dlp / YouTube     | 3–8s random delay between requests; max 800 enrichments/day            |
| YouTube Data API     | Track daily unit spend in SQLite; pause when within 500 units of limit |
| SponsorBlock API     | 1s delay; lightweight endpoint, generous limits                        |
| Agency websites      | 5–10s delay; one session per domain per run                            |
| General web scraping | Backoff retry with exponential backoff; max 3 retries                  |

A `rate_limit_budget` table in SQLite tracks daily API spend, reset at midnight.

---

## 15. File & Directory Structure

```
ytleadbot/
├── cmd/
│   └── ytleadbot/
│       └── main.go              # CLI entrypoint, arg parsing
├── internal/
│   ├── config/
│   │   ├── config.go            # .env + JSON config loading
│   │   ├── keywords_seed.json   # Discovery keyword matrix
│   │   └── agencies_seed.json   # Known agency websites
│   ├── core/
│   │   ├── db.go                # SQLite connection, schema, DAL
│   │   ├── orchestrator.go      # Job queue loop
│   │   └── scheduler.go         # Cron-compatible dispatch
│   ├── modules/
│   │   ├── discovery.go         # Channel ID discovery
│   │   ├── enrichment.go        # Channel metadata + scoring
│   │   ├── agency_finder.go     # Agency detection
│   │   └── exporter.go          # Export to JSON/CSV
│   └── utils/
│       ├── ytdlp.go             # yt-dlp subprocess wrapper
│       ├── scraper.go           # HTTP + goquery helpers
│       ├── email_extractor.go   # Email finding logic
│       └── monetization.go      # Scoring functions
├── data/
│   └── ytleadbot.db                 # SQLite database
├── exports/                     # JSON/CSV exports land here
├── logs/                        # Log output
├── .env.example                 # Template for secrets
├── .gitignore
├── go.mod
└── README.md
```

---

## 16. Downstream Pipeline Integration

The outreach pipeline reads from the SQLite DB directly via the `v_export_ready` view, or consumes the nightly JSON export. The contract:

**Input schema (what this system produces):**

- `channel_id`, `name`, `handle`, `email`, `subscriber_count`
- `niche`, `monetization_score`, `monetization_signals` (array)
- `website`, `social_links`, `country`
- `agency_id`, `agency_name` (if applicable)

**Handoff:** The downstream pipeline's `valueFunc` receives each lead and determines whether/how to personalize outreach. The `monetization_signals` array is the richest input for this — it tells the valueFunc _how_ this creator makes money, enabling highly targeted messaging (e.g. "We help Teachable creators..." for a creator with a course link).

After the downstream pipeline processes a lead, it should call back to update `channels.status = 'contacted'` to prevent re-export.

---

## 17. Risks & Mitigations

| Risk                                               | Likelihood | Mitigation                                                                        |
| -------------------------------------------------- | ---------- | --------------------------------------------------------------------------------- |
| YouTube rate limiting / IP block                   | Medium     | Slow polling, user-agent rotation, backoff; consider residential proxy if blocked |
| yt-dlp breaking on YouTube changes                 | Medium     | Pin yt-dlp version, monitor GitHub for updates                                    |
| Low email coverage (<40%)                          | High       | Accept it; email on About is optional. Use downstream enrichment pass             |
| False positives (low-quality channels scored high) | Low        | Tune threshold; add recency check (last upload < 60 days)                         |
| Agency websites restructuring                      | Low        | Manual seed list maintenance; quarterly review                                    |
| YouTube API quota exhaustion                       | Low        | Budget tracking in SQLite; yt-dlp fallback for most operations                    |

---

## 18. Build Phases (Go-specific)

### Slice 1 — End-to-End Thin Slice (~1 session)

- [ ] Initialize Go module, add dependencies
- [ ] `internal/core/db.go` — SQLite schema + DAL (upsert, job queue ops, views)
- [ ] `internal/config/config.go` — `.env` loading via godotenv
- [ ] `internal/utils/ytdlp.go` — `exec.Command("yt-dlp", "--dump-json", ...)` wrapper with retry
- [ ] `internal/utils/scraper.go` — `goquery`-based HTML scraping helpers
- [ ] `internal/utils/email_extractor.go` — regex email extraction from text/HTML
- [ ] `internal/utils/monetization.go` — SponsorBlock API call + link pattern scoring
- [ ] `internal/modules/enrichment.go` — full enrich pipeline for one channel
- [ ] `internal/modules/exporter.go` — JSON dump with `v_export_ready`
- [ ] `cmd/ytleadbot/main.go` — `--job enrich "UCxxx"` and `--job export` CLI commands
- [ ] Build: `go build -o bin/ytleadbot ./cmd/ytleadbot`

### Slice 2 — Discovery at Scale (~1 session)

- [ ] `internal/modules/discovery.go` — keyword search, YouTube API search, related channels, dedup
- [ ] `internal/core/orchestrator.go` — job loop with dequeue/dispatch/retry
- [ ] `internal/core/scheduler.go` — `--job discover` and `--job reenrich_stale` modes

### Slice 3 — Agency Finder (~1 session)

- [ ] `internal/modules/agency_finder.go` — roster scraping, email domain clustering, description regex
- [ ] Wire into orchestrator as `agency_scan` job type

### Slice 4 — Hardening (~1 session)

- [ ] Rate-limit budget tracking table + enforcement
- [ ] Stale lead re-enrichment
- [ ] Run logging stats in `runs` table
- [ ] Error alerting stub
- [ ] Documentation + deployment notes

---

## 19. Success Metrics

| Metric               | Target                                           |
| -------------------- | ------------------------------------------------ |
| New leads per week   | 500–1,000                                        |
| Email capture rate   | ≥ 35% of enriched leads                          |
| Agency rosters found | 50+ agencies, 500+ linked channels               |
| False positive rate  | < 10% (non-monetizing channels slipping through) |
| Cost to operate      | $0/month                                         |
| Uptime / reliability | Runs unattended for 30+ days                     |

---

_End of document._
