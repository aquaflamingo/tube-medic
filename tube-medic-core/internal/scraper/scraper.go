package scraper

import (
	"regexp"
	"strings"

	"github.com/aquaflamingo/tubemedicmvp/internal/youtube"
)

var urlRe = regexp.MustCompile(`https?://[^\s<>"']+`)

// trailingPunct matches punctuation commonly found after URLs in descriptions
// that is not part of the actual URL.
var trailingPunct = regexp.MustCompile(`[.,!?;:)\]>]+$`)

// ExtractAll extracts all URLs from the descriptions of the given videos.
// Each URL returned is paired with its source video.
func ExtractAll(videos []youtube.Video) []youtube.ScrapedLink {
	var links []youtube.ScrapedLink
	seen := make(map[string]bool)

	for _, v := range videos {
		matches := urlRe.FindAllString(v.Description, -1)
		for _, raw := range matches {
			clean := cleanURL(raw)
			if clean == "" || seen[clean] {
				continue
			}
			seen[clean] = true
			links = append(links, youtube.ScrapedLink{
				URL:        clean,
				VideoID:    v.ID,
				VideoTitle: v.Title,
			})
		}
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

func cleanURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = trailingPunct.ReplaceAllString(raw, "")
	return raw
}
