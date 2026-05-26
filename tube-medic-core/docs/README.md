# Tube Medic MVP

Scans a YouTube channel's recent videos, extracts all URLs from descriptions, and checks each one for broken links (4xx, 5xx, timeout, DNS failure).

## Setup

```bash
cp .env.example .env
# Edit .env and add your YouTube Data API v3 key:
#   YT_API_KEY=your_key_here

go build -o tube-medic ./cmd/tube-medic/
```

## Usage

```bash
./tube-medic --channel https://youtube.com/@mkbhd
./tube-medic --channel https://youtube.com/@mkbhd --max-videos 100
./tube-medic --channel https://youtube.com/channel/UCxxx
```

Flags:
| Flag | Default | Description |
|------|---------|-------------|
| `--channel` | (required) | Full YouTube channel URL |
| `--api-key` | `YT_API_KEY` env var | YouTube Data API key |
| `--max-videos` | `50` | Number of recent videos to scan |

## Testing

```bash
go test ./... -update   # regenerate golden files
go test ./... -v        # run all tests
```

Tests use only local HTTP servers — no external API calls.

## Project Layout

```
├── cmd/tube-medic/main.go     # entry point, pipeline orchestration
├── internal/
│   ├── config/                # .env + flag parsing
│   ├── youtube/               # YouTube Data API v3 client
│   ├── scraper/               # URL extraction from descriptions
│   ├── checker/               # concurrent link health checks
│   └── report/                # stdout output formatting
```

## Output

```
Tube Medic Report ─ "MKBHD"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Videos scanned    50
  Total links found 234
  Working           228
  Broken             6

 Broken Links
 ────────────────────────────────────────────
  1. 404        https://example.com/dead-link
               → "Best Gadgets 2024"
               https://youtube.com/watch?v=abc123

  2. DNS_FAIL   https://expired-domain.org
               → "My Setup Tour"
               https://youtube.com/watch?v=def456
```

Exits 0 if no broken links, 1 if any found.
