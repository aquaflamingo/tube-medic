## Architecture: Core + Satellites

The module root (`tmcore.go`) is the **public API** — it exports `Config`, `Report`, and `RunScan` for programmatic consumers (CLI, web, etc.).

```
tmcore (root)                    ← public API, imports internal/*
cmd/tube-medic/                  ← CLI satellite (config.Load + tmcore.RunScan + report.Print)
tubemedic-web (external)         ← web satellite (imports root package)
```

### Key Files

| File                                | Purpose                                                             |
| ----------------------------------- | ------------------------------------------------------------------- |
| `tmcore.go`                         | Public API: `Config`, `Report`, `RunScan`                           |
| `internal/checker/transport.go`     | HTTP client: standard Go TLS, browser headers, cookie jar           |
| `internal/checker/checker.go`       | Worker pool for concurrent link checking                            |
| `internal/checker/result.go`        | Status classification, error categorization (NET_ERR, DNS_FAIL etc) |
| `internal/youtube/client.go`        | YouTube API client                                                  |
| `internal/config/config.go`         | CLI flag/env parsing (returns public `Config`, not in API surface)  |
| `internal/report/report.go`         | Terminal-formatted output (CLI-only)                                |
| `internal/scraper/scraper.go`       | URL extraction from video descriptions                              |
| `internal/classifier/classifier.go` | Revenue-critical link classification                                |

### Error Classification (`internal/checker/result.go:errorKind`)

- `DNS_FAIL` — no such host / DNS errors
- `TIMEOUT` — timeout errors
- `REFUSED` — connection refused
- `TLS_ERR` — TLS/certificate errors
- `NET_ERR` — everything else (EOF, unexpected EOF, etc.)

### Commands

```sh
go test ./...
go vet ./...
go build ./...
```
