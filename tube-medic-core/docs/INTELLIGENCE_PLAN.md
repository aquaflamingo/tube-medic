# Link Intelligence Plan

## Goal
When checking links in video descriptions, detect which links drive revenue (products, courses, memberships, affiliates, donations) and surface broken ones prominently in the report.

## Implementation

### 1. Add `Context` to `ScrapedLink` (`youtube/types.go`)
- New field: `Context string` — surrounding text from the description

### 2. Update scraper to capture context (`scraper/scraper.go`)
- Replace naive URL extraction with context-aware extraction
- Capture ~60 chars before and after each URL
- When deduplicating across videos, merge context strings
- Modified `ExtractAll` in-place

### 3. New package: `internal/classifier/`

```go
type Priority int
const (
    PriorityNormal  Priority = 0
    PriorityRevenue Priority = 1
)

func Classify(link youtube.ScrapedLink) Priority
```

Heuristics:
- **URL patterns**: `/product/`, `/course/`, `/checkout`, `/cart`, `/purchase`, `/pricing`, `/membership`, `/subscribe`, `/donate`, `/support`, `/order`, `/enroll`, `/download`, `/shop`, `/store`, `/bundle`, `/template`, `/ebook` — and domains like gumroad.com, teachable.com, patreon.com, ko-fi.com, shopify.com, etsy.com, amazon.*/dp/, plus affiliate markers (`ref=`, `tag=`, `affiliate`, `referral`)
- **Context keywords** (case-insensitive): buy, purchase, checkout, cart, course, enroll, product, digital, download, bundle, shop, store, membership, premium, affiliate, referral, patreon, ko-fi, donation, donate, support, tip, template, ebook, subscribe

### 4. Add `Priority` to `CheckResult` (`checker/result.go`)
```go
type CheckResult struct {
    // ... existing fields
    Priority Priority
}
```

### 5. Wire classification into `CheckAll` + `Summarize` (`checker/checker.go`)
- Classify links during `CheckAll`

### 6. `Summary` gets `CriticalBroken` + `CriticalLinks` (`checker/result.go`)
```go
type Summary struct {
    // ... existing fields
    CriticalBroken  int
    CriticalLinks   []CheckResult
}
```

### 7. Update report (`report/report.go`)
- Summary gets `Revenue-critical  N` line
- Broken links split into **two sections** (both only shown when > 0):
  1. **Revenue-Critical Broken Links** (red ANSI, shown first)
  2. **Other Broken Links** (standard color, shown below)

### 8. Update golden test files & existing tests

### Files changed
| File | Change |
|------|--------|
| `youtube/types.go` | `Context` on `ScrapedLink` |
| `scraper/scraper.go` | Context-aware extraction |
| `scraper/scraper_test.go` | Update tests |
| `classifier/classifier.go` | **New** |
| `classifier/classifier_test.go` | **New** |
| `checker/result.go` | `Priority`, `CriticalBroken`, `CriticalLinks` |
| `checker/checker.go` | Classify in `CheckAll` |
| `checker/checker_test.go` | Update tests |
| `report/report.go` | Two sections + summary line |
| `report/report_test.go` | Update golden files |
