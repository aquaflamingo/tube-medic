## Session: TLS Fingerprint Spoofing → Simpler Approach

**Goal:** Bypass 403 bot protection when checking external links in YouTube video descriptions (primary target: Udemy).

### Approach 1: uTLS (abandoned)

- Added `github.com/refraction-networking/utls v1.8.2` and `golang.org/x/net v0.54.0`
- Created `internal/checker/transport.go` — uTLS transport with Chrome 133 fingerprint (`HelloChrome_Auto`), browser-grade headers, cookie jar
- Replaced `userAgentTransport` wrapper with `newBrowserClient()`

**Problem:** Chrome's ALPN advertises `["h2", "http/1.1"]`. Go's `http.Transport.dialConn` type-asserts the raw connection to `*tls.Conn` when ALPN says `h2`, but `*utls.UConn` fails that assertion — causing HTTP/2 to silently skip. Server then expects HTTP/2 frames while client sends HTTP/1.1 → protocol mismatch.

**Fix attempted:** Before uTLS handshake, find `*utls.ALPNExtension` in `tlsConn.Extensions` and remove `"h2"` from `AlpnProtocols`. Set `ForceAttemptHTTP2: false`.

**Outcome:** Still got `NET_ERR` on sites where even standard `crypto/tls.Dial` fails (IP-level CDN blocks). The complexity of uTLS + ALPN surgery was not worth it.

### Approach 2: Standard Go TLS (current)

- Removed `github.com/refraction-networking/utls` (and 4 transitive deps)
- `internal/checker/transport.go` now uses plain `http.Transport` with browser headers wrapper and cookie jar — no custom TLS stack
- HTTP/2 works natively

### Key Files

| File                       | Purpose                                                               |
| -------------------------- | --------------------------------------------------------------------- |
| `internal/checker/transport.go` | HTTP client: standard Go TLS, browser headers, cookie jar             |
| `internal/checker/checker.go`   | Worker pool for concurrent link checking                              |
| `internal/checker/result.go`    | Status classification, error categorization (NET_ERR, DNS_FAIL, etc.) |
| `internal/youtube/client.go`    | YouTube API client                                                    |

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
