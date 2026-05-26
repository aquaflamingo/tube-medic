package scraper_test

import (
	"testing"

	"github.com/aqfl/tmcore/internal/scraper"
	"github.com/aqfl/tmcore/internal/youtube"
)

func TestExtractFromString(t *testing.T) {
	tests := []struct {
		name  string
		desc  string
		want  []string
		count int
	}{
		{
			name:  "no urls",
			desc:  "this is a description with no links",
			want:  nil,
			count: 0,
		},
		{
			name:  "single url",
			desc:  "check this out: https://example.com/product",
			want:  []string{"https://example.com/product"},
			count: 1,
		},
		{
			name:  "multiple urls",
			desc:  "link1: https://example.com/a and link2: https://example.com/b",
			want:  []string{"https://example.com/a", "https://example.com/b"},
			count: 2,
		},
		{
			name:  "url with trailing period",
			desc:  "visit https://example.com/page for more.",
			want:  []string{"https://example.com/page"},
			count: 1,
		},
		{
			name:  "url with trailing comma",
			desc:  "check https://example.com/a, https://example.com/b",
			want:  []string{"https://example.com/a", "https://example.com/b"},
			count: 2,
		},
		{
			name:  "url with trailing parens",
			desc:  "see (https://example.com/page) for details",
			want:  []string{"https://example.com/page"},
			count: 1,
		},
		{
			name:  "https only",
			desc:  "http://insecure.com and https://secure.com",
			want:  []string{"http://insecure.com", "https://secure.com"},
			count: 2,
		},
		{
			name:  "url with path and query",
			desc:  "shop at https://shop.example.com/products?id=123&ref=yt",
			want:  []string{"https://shop.example.com/products?id=123&ref=yt"},
			count: 1,
		},
		{
			name:  "duplicate urls in same description",
			desc:  "https://example.com and again https://example.com",
			want:  []string{"https://example.com"},
			count: 1,
		},
		{
			name:  "url with trailing exclamation",
			desc:  "must click: https://example.com/deal!",
			want:  []string{"https://example.com/deal"},
			count: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scraper.ExtractFromString(tt.desc)
			if len(got) != tt.count {
				t.Errorf("got %d urls, want %d: %v", len(got), tt.count, got)
			}
			if tt.want != nil {
				for i, w := range tt.want {
					if i >= len(got) {
						t.Errorf("missing expected url %q", w)
						continue
					}
					if got[i] != w {
						t.Errorf("url[%d] = %q, want %q", i, got[i], w)
					}
				}
			}
		})
	}
}

func TestExtractAll_deduplicatesAcrossVideos(t *testing.T) {
	videos := []youtube.Video{
		{ID: "vid1", Title: "First Video", Description: "link: https://example.com/abc"},
		{ID: "vid2", Title: "Second Video", Description: "same link: https://example.com/abc"},
	}

	links := scraper.ExtractAll(videos)
	if len(links) != 1 {
		t.Fatalf("expected 1 unique link, got %d", len(links))
	}
	if links[0].URL != "https://example.com/abc" {
		t.Errorf("link url = %q, want %q", links[0].URL, "https://example.com/abc")
	}
}

func TestExtractAll_ignoresNoURLs(t *testing.T) {
	videos := []youtube.Video{
		{ID: "vid1", Title: "No Links", Description: "just text"},
	}
	links := scraper.ExtractAll(videos)
	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}
