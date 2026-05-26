package scraper

import (
	"regexp"
	"strings"

	"github.com/aquaflamingo/tmcore/internal/youtube"
)

var urlRe = regexp.MustCompile(`https?://[^\s<>"']+`)

// trailingPunct matches punctuation commonly found after URLs in descriptions
// that is not part of the actual URL.
var trailingPunct = regexp.MustCompile(`[.,!?;:)\]>]+$`)

const contextWindow = 60

// ExtractAll extracts all URLs from the descriptions of the given videos.
// Each URL returned is paired with its source video and surrounding context.
func ExtractAll(videos []youtube.Video) []youtube.ScrapedLink {
	type entry struct {
		link  youtube.ScrapedLink
		ctxs  []string
	}

	seen := make(map[string]*entry)

	for _, v := range videos {
		matches := urlRe.FindAllStringIndex(v.Description, -1)
		for _, m := range matches {
			raw := v.Description[m[0]:m[1]]
			clean := cleanURL(raw)
			if clean == "" {
				continue
			}
			ctx := extractContext(v.Description, m[0], m[1])
			if e, ok := seen[clean]; ok {
				e.ctxs = append(e.ctxs, ctx)
			} else {
				seen[clean] = &entry{
					link: youtube.ScrapedLink{
						URL:        clean,
						VideoID:    v.ID,
						VideoTitle: v.Title,
					},
					ctxs: []string{ctx},
				}
			}
		}
	}

	links := make([]youtube.ScrapedLink, 0, len(seen))
	for _, e := range seen {
		e.link.Context = mergeContexts(e.ctxs)
		links = append(links, e.link)
	}
	return links
}

// ExtractFromString extracts URLs from a single description string.
// Exported for testing.
func ExtractFromString(desc string) []string {
	matches := urlRe.FindAllString(desc, -1)
	var out []string
	seen := make(map[string]bool)
	for _, raw := range matches {
		clean := cleanURL(raw)
		if clean == "" || seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	return out
}

func extractContext(desc string, matchStart, matchEnd int) string {
	start := matchStart - contextWindow
	if start < 0 {
		start = 0
	}
	end := matchEnd + contextWindow
	if end > len(desc) {
		end = len(desc)
	}
	return desc[start:end]
}

func mergeContexts(ctxs []string) string {
	seen := make(map[string]bool)
	var merged []string
	for _, c := range ctxs {
		if !seen[c] {
			seen[c] = true
			merged = append(merged, c)
		}
	}
	return strings.Join(merged, " | ")
}

func cleanURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = trailingPunct.ReplaceAllString(raw, "")
	return raw
}
