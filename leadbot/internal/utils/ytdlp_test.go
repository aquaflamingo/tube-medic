package utils

import (
	"testing"
	"time"
)

func TestBuildChannelURL(t *testing.T) {
	tests := []struct {
		ident string
		want  string
	}{
		{"UCabcdefghijklmnopqrstuv", "https://www.youtube.com/channel/UCabcdefghijklmnopqrstuv/about"},
		{"@handle", "https://www.youtube.com/@handle/about"},
		{"handle", "https://www.youtube.com/@handle/about"},
		{"  @spaced  ", "https://www.youtube.com/@spaced/about"},
	}
	for _, tc := range tests {
		got := buildChannelURL(tc.ident)
		if got != tc.want {
			t.Errorf("buildChannelURL(%q) = %q, want %q", tc.ident, got, tc.want)
		}
	}
}

func TestBuildVideoURL(t *testing.T) {
	tests := []struct {
		ident string
		want  string
	}{
		{"UCabcdefghijklmnopqrstuv", "https://www.youtube.com/channel/UCabcdefghijklmnopqrstuv/videos"},
		{"@handle", "https://www.youtube.com/@handle/videos"},
		{"handle", "https://www.youtube.com/@handle/videos"},
	}
	for _, tc := range tests {
		got := buildVideoURL(tc.ident)
		if got != tc.want {
			t.Errorf("buildVideoURL(%q) = %q, want %q", tc.ident, got, tc.want)
		}
	}
}

func TestCalculateUploadFrequency(t *testing.T) {
	now := time.Now()
	day := func(d int) string {
		return now.AddDate(0, 0, d).Format("20060102")
	}

	tests := []struct {
		name   string
		videos []YTDLPVideo
		want   float64
	}{
		{
			name:   "less than 2 videos",
			videos: []YTDLPVideo{{ID: "v1", UploadDate: day(-1)}},
			want:   0,
		},
		{
			name:   "empty list",
			videos: []YTDLPVideo{},
			want:   0,
		},
		{
			name: "no valid dates",
			videos: []YTDLPVideo{
				{ID: "v1", UploadDate: ""},
				{ID: "v2", UploadDate: ""},
			},
			want: 0,
		},
		{
			name: "2 videos over ~1 month",
			videos: []YTDLPVideo{
				{ID: "v1", UploadDate: day(-30)},
				{ID: "v2", UploadDate: day(0)},
			},
			want: 2.0,
		},
		{
			name: "5 videos over 1 month",
			videos: []YTDLPVideo{
				{ID: "v1", UploadDate: day(-30)},
				{ID: "v2", UploadDate: day(-23)},
				{ID: "v3", UploadDate: day(-16)},
				{ID: "v4", UploadDate: day(-8)},
				{ID: "v5", UploadDate: day(0)},
			},
			want: 5.0,
		},
		{
			name: "clamps minimum span to 0.5 months",
			videos: []YTDLPVideo{
				{ID: "v1", UploadDate: day(-1)},
				{ID: "v2", UploadDate: day(0)},
			},
			want: 2.0 / 0.5,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateUploadFrequency(tc.videos)
			if got != tc.want {
				t.Errorf("CalculateUploadFrequency = %v, want %v", got, tc.want)
			}
		})
	}
}
