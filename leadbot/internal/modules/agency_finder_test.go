package modules

import (
	"testing"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		rawURL string
		want   string
	}{
		{"https://example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"https://www.example.com", "example.com"},
		{"https://example.com/path", "example.com"},
		{"example.com", "example.com"},
		{"", ""},
	}
	for _, tc := range tests {
		got := extractDomain(tc.rawURL)
		if got != tc.want {
			t.Errorf("extractDomain(%q) = %q, want %q", tc.rawURL, got, tc.want)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Simple Name", "simple-name"},
		{"Name & Co", "name-and-co"},
		{"  Spaces  ", "spaces"},
		{"Dots.And,Commas", "dotsandcommas"},
		{"UPPERCASE", "uppercase"},
		{"", ""},
	}
	for _, tc := range tests {
		got := slugify(tc.input)
		if got != tc.want {
			t.Errorf("slugify(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
