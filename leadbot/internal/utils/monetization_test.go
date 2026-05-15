package utils

import (
	"testing"
)

func TestDetectPlatforms(t *testing.T) {
	tests := []struct {
		name  string
		links []string
		want  map[string]bool
	}{
		{
			name:  "patreon",
			links: []string{"https://patreon.com/creator"},
			want:  map[string]bool{"patreon": true},
		},
		{
			name:  "course platform",
			links: []string{"https://gumroad.com/course"},
			want:  map[string]bool{"course": true},
		},
		{
			name:  "merch",
			links: []string{"https://shopify.com/store"},
			want:  map[string]bool{"merch": true},
		},
		{
			name:  "amazon",
			links: []string{"https://amazon.com/shop"},
			want:  map[string]bool{"amazon": true},
		},
		{
			name:  "linktree",
			links: []string{"https://linktr.ee/creator"},
			want:  map[string]bool{"linktree": true},
		},
		{
			name:  "multiple platforms",
			links: []string{"https://patreon.com/c", "https://gumroad.com/c", "https://linktr.ee/c"},
			want:  map[string]bool{"patreon": true, "course": true, "linktree": true},
		},
		{
			name:  "no platforms",
			links: []string{"https://example.com"},
			want:  map[string]bool{},
		},
		{
			name:  "empty links",
			links: []string{},
			want:  map[string]bool{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectPlatforms(tc.links)
			for k, v := range tc.want {
				if got[k] != v {
					t.Errorf("DetectPlatforms: key %q = %v, want %v", k, got[k], v)
				}
			}
			for k, v := range got {
				if !tc.want[k] {
					t.Errorf("DetectPlatforms: unexpected key %q = %v", k, v)
				}
			}
		})
	}
}

func TestExtractLinksFromDescription(t *testing.T) {
	tests := []struct {
		name  string
		desc  string
		count int
	}{
		{
			name:  "extracts http links",
			desc:  "Check out my site: https://example.com and http://blog.example.com",
			count: 2,
		},
		{
			name:  "empty description",
			desc:  "",
			count: 0,
		},
		{
			name:  "no links",
			desc:  "Just some text without any URLs",
			count: 0,
		},
		{
			name:  "handles punctuation around URLs",
			desc:  "Visit https://example.com, and https://example2.com!",
			count: 2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractLinksFromDescription(tc.desc)
			if len(got) != tc.count {
				t.Errorf("ExtractLinksFromDescription returned %d links, want %d. Got: %v", len(got), tc.count, got)
			}
		})
	}
}

func TestCalculateScore_NoSignals(t *testing.T) {
	info := &YTDLPMetadata{}
	result := CalculateScore(info, nil, nil, false)
	if result.Score != 0 {
		t.Errorf("expected score 0 with no signals, got %d", result.Score)
	}
	if result.Tier != "standard" {
		t.Errorf("expected tier 'standard', got %q", result.Tier)
	}
	if len(result.Signals) != 0 {
		t.Errorf("expected 0 signals, got %d", len(result.Signals))
	}
}

func TestCalculateScore_WithEmail(t *testing.T) {
	info := &YTDLPMetadata{
		Email: "creator@example.com",
	}
	result := CalculateScore(info, nil, nil, false)
	if result.Score < 3 {
		t.Errorf("expected score >= 3 with email, got %d", result.Score)
	}
	if len(result.Signals) == 0 {
		t.Error("expected at least 1 signal with email")
	}
}

func TestCalculateScore_WithChannelEmail(t *testing.T) {
	info := &YTDLPMetadata{
		ChannelEmail: "business@example.com",
	}
	result := CalculateScore(info, nil, nil, false)
	if result.Score < 3 {
		t.Errorf("expected score >= 3 with channel email, got %d", result.Score)
	}
}

func TestCalculateScore_WithPlatforms(t *testing.T) {
	info := &YTDLPMetadata{}
	links := []string{
		"https://patreon.com/creator",
		"https://gumroad.com/course",
	}
	result := CalculateScore(info, nil, links, false)
	if result.Score < 4 {
		t.Errorf("expected score >= 4 from patreon(2)+course(2), got %d", result.Score)
	}
}

func TestCalculateScore_WithPricing(t *testing.T) {
	info := &YTDLPMetadata{}
	result := CalculateScore(info, nil, nil, true)
	if result.Score < 1 {
		t.Errorf("expected score >= 1 with pricing page, got %d", result.Score)
	}
}

func TestCalculateScore_WithMembership(t *testing.T) {
	info := &YTDLPMetadata{
		ChannelIsMembership: true,
	}
	result := CalculateScore(info, nil, nil, false)
	if result.Score < 1 {
		t.Errorf("expected score >= 1 with membership, got %d", result.Score)
	}
}

func TestCalculateScore_HighTier(t *testing.T) {
	info := &YTDLPMetadata{
		Email: "creator@example.com",
	}
	// email(3) + patreon(2) + course(2) + merch(1) = 8
	links := []string{
		"https://patreon.com/creator",
		"https://gumroad.com/course",
		"https://shopify.com/store",
	}
	result := CalculateScore(info, nil, links, false)
	if result.Tier != "high" {
		t.Errorf("expected tier 'high' for score >= 8, got %q with score %d", result.Tier, result.Score)
	}
}
