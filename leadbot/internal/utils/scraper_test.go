package utils

import (
	"testing"
)

func TestIsLinktreeURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://linktr.ee/creator", true},
		{"https://beacons.ai/creator", true},
		{"https://koji.to/creator", true},
		{"https://bio.link/creator", true},
		{"https://bento.me/creator", true},
		{"https://example.com", false},
		{"", false},
	}
	for _, tc := range tests {
		got := IsLinktreeURL(tc.url)
		if got != tc.want {
			t.Errorf("IsLinktreeURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestIsPatreonURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://patreon.com/creator", true},
		{"https://ko-fi.com/creator", true},
		{"https://buymeacoffee.com/creator", true},
		{"https://example.com", false},
	}
	for _, tc := range tests {
		got := IsPatreonURL(tc.url)
		if got != tc.want {
			t.Errorf("IsPatreonURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestIsCourseURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://teachable.com/course", true},
		{"https://kajabi.com/course", true},
		{"https://gumroad.com/course", true},
		{"https://podia.com/course", true},
		{"https://skool.com/course", true},
		{"https://thinkific.com/course", true},
		{"https://udemy.com/course", true},
		{"https://example.com", false},
	}
	for _, tc := range tests {
		got := IsCourseURL(tc.url)
		if got != tc.want {
			t.Errorf("IsCourseURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestIsMerchURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://shopify.com/store", true},
		{"https://teespring.com/store", true},
		{"https://spring.com/store", true},
		{"https://example.com/merch", true},
		{"https://example.com", false},
	}
	for _, tc := range tests {
		got := IsMerchURL(tc.url)
		if got != tc.want {
			t.Errorf("IsMerchURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestIsAmazonURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://amazon.com/shop/creator", true},
		{"https://amazon.co.uk/shop/creator", true},
		{"https://amzn.to/xyz", true},
		{"https://example.com", false},
	}
	for _, tc := range tests {
		got := IsAmazonURL(tc.url)
		if got != tc.want {
			t.Errorf("IsAmazonURL(%q) = %v, want %v", tc.url, got, tc.want)
		}
	}
}

func TestExtractYouTubeChannelIdent(t *testing.T) {
	tests := []struct {
		rawURL string
		want   string
	}{
		{"https://www.youtube.com/channel/UCabc123", "UCabc123"},
		{"https://www.youtube.com/c/MyChannel", "@MyChannel"},
		{"https://www.youtube.com/@handle", "@handle"},
		{"https://www.youtube.com/user/username", "@username"},
		{"https://example.com", ""},
		{"", ""},
	}
	for _, tc := range tests {
		got := ExtractYouTubeChannelIdent(tc.rawURL)
		if got != tc.want {
			t.Errorf("ExtractYouTubeChannelIdent(%q) = %q, want %q", tc.rawURL, got, tc.want)
		}
	}
}

func TestExtractYouTubeIDsFromPage_NoNetwork(t *testing.T) {
	// Should return error for unreachable URLs (not a network test)
	_, err := ExtractYouTubeIDsFromPage("https://nonexistent.invalid/page")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}
