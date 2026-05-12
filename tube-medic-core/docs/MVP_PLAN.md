
# Tube Medic MVP — Implementation Plan

## Project Layout
```
tube-medic-mvp/
├── .env.example                 # YT_API_KEY=your_key_here
├── .gitignore
├── go.mod                       # module github.com/aquaflamingo/tube-medic-mvp
├── cmd/
│   └── tube-medic/
│       └── main.go              # Entry point — flags, wiring, pipeline orchestration
├── internal/
│   ├── config/
│   │   └── config.go            # .env loading + CLI flag parsing → Config struct
│   ├── youtube/
│   │   ├── client.go            # YouTube Data API v3 calls (channels, search, videos)
│   │   └── types.go             # Channel, Video, API response structs
│   ├── scraper/
│   │   └── scraper.go           # regex URL extraction from description text
│   ├── checker/
│   │   ├── checker.go           # Worker pool (10 goroutines), GET checks with 5s timeout
│   │   └── result.go            # CheckResult type + classification
│   └── report/
│       └── report.go            # stdout summary + broken link listing
```

## Dependencies

**None beyond Go stdlib.** No external Go modules. Parse `.env` manually (key=value lines). `net/http` handles everything (YouTube API + link checking).

## Data Flow
```
.env (YT_API_KEY) + CLI flags
     ↓
┌─ config.Load() ────────────────────────┐
│  APIKey, ChannelURL, MaxVideos         │
└──────┬─────────────────────────────────┘
       ↓
┌─ youtube.Client ──────────────────────────────────────────────┐
│  1. Parse channel URL → extract handle/ID                    │
│  2. GET /channels?forHandle=@handle → channel ID + name      │
│  3. GET /search?channelId=X&order=date → []{videoID, title}  │
│  4. GET /videos?id=id1,id2,... → map[videoID]description     │
└──────┬────────────────────────────────────────────────────────┘
       ↓
┌─ scraper.ExtractAll(descriptions) ───────────────────────────┐
│  → []ScrapedLink{URL, VideoID, VideoTitle}                   │
└──────┬────────────────────────────────────────────────────────┘
       ↓
┌─ checker.CheckAll(links, concurrency=10) ────────────────────┐
│  Worker pool — GET requests, 5s timeout, follows redirects   │
│  → []CheckResult{URL, StatusCode, Error, Status}             │
└──────┬────────────────────────────────────────────────────────┘
       ↓
┌─ report.Print(results) ──────────────────────────────────────┐
│  Summary table + broken link details                          │
│  Exit code 1 if any broken links found                        │
└───────────────────────────────────────────────────────────────┘
```

## CLI Interface
```
tube-medic --channel https://youtube.com/@mkbhd
tube-medic --channel https://youtube.com/@mkbhd --api-key X
tube-medic --channel https://youtube.com/channel/UCxxx
tube-medic --channel https://youtube.com/@mkbhd --max-videos 100
```

Flags:
- `--channel` (required) — full YouTube channel URL
- `--api-key` (optional, falls back to `YT_API_KEY` in `.env`)
- `--max-videos` (optional, default 50)

## URL Parsing (channel input)
- `https://youtube.com/@handle` → extract handle, resolve via `channels.list?forHandle=@handle`
- `https://youtube.com/channel/UCxxx` → extract channel ID directly
- `https://youtube.com/c/CustomName` → extract custom name, fallback resolve

## Link Checking (GET-only)
All link checks use **GET** requests. `http.Client` with 5s timeout, follows redirects.
- **Working** — 2xx/3xx (any redirect that resolves)
- **Broken** — 4xx/5xx, timeout, DNS failure, connection refused, TLS error
- **Skipped** — mailto:, non-http schemes, obviously malformed

## Output Format
```
Tube Medic Report ─ "Channel Name"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Videos scanned     50
  Total links found  234
  Working            228
  Broken              6

 Broken Links
 ────────────────────────────────────────────
  1. 404       https://example.com/dead-link
               → "Best Gadgets 2024"
               https://youtube.com/watch?v=abc123

  2. TIMEOUT   https://expired-domain.org
               → "My Studio Tour"
               https://youtube.com/watch?v=def456

  3. DNS_FAIL  https://nonexistent.shop
               → "Budget Tech Picks"
               https://youtube.com/watch?v=ghi789
```

Exit code 0 if no broken links, 1 if any found.

## Testing

### Table Tests
Each internal package has a `*_test.go` with table-driven tests covering:
- **scraper**: URL extraction edge cases (trailing punctuation, duplicates, query params)
- **checker**: HTTP status classification, timeout/DNS error simulation
- **youtube**: URL parsing for channel resolution
- **report**: Render output format fixtures

### Golden File Tests
Full pipeline integration tests in `internal/scraper/testdata/` and `internal/report/testdata/`:
- Scraper output for a known set of descriptions → golden `.txt` of expected `ScrapedLink` slice
- Report output for a known set of `CheckResult` → golden `.txt` of expected terminal output

Run with:
```
go test ./... -update   # update golden files (if flag added)
go test ./...
```

## Execution
All Go commands (go mod init, go build, etc.) must be run by the user. Code is written to disk only.
