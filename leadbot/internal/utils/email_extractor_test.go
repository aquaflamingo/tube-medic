package utils

import (
	"testing"
)

func TestExtractEmails(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		want  int
		first string
	}{
		{
			name:  "simple email",
			text:  "Contact me at creator@example.com",
			want:  1,
			first: "creator@example.com",
		},
		{
			name:  "multiple emails",
			text:  "Email: a@b.com or c@d.com",
			want:  2,
			first: "a@b.com",
		},
		{
			name:  "filters noreply",
			text:  "noreply@example.com",
			want:  0,
		},
		{
			name:  "filters info",
			text:  "info@example.com",
			want:  0,
		},
		{
			name:  "filters image extension",
			text:  "image@example.png",
			want:  0,
		},
		{
			name:  "duplicates removed",
			text:  "a@b.com a@b.com",
			want:  1,
			first: "a@b.com",
		},
		{
			name:  "no email",
			text:  "this has no email address",
			want:  0,
		},
		{
			name:  "empty string",
			text:  "",
			want:  0,
		},
		{
			name:  "preserves valid business email",
			text:  "reach out to business@startup.io",
			want:  1,
			first: "business@startup.io",
		},
		{
			name:  "filters donotreply",
			text:  "donotreply@example.com",
			want:  0,
		},
		{
			name:  "handles dots and trailing punctuation",
			text:  "Email me at user@domain.com.",
			want:  1,
			first: "user@domain.com",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractEmails(tc.text)
			if len(got) != tc.want {
				t.Errorf("ExtractEmails(%q) returned %d emails, want %d. Got: %v", tc.text, len(got), tc.want, got)
			}
			if tc.want > 0 && tc.first != "" && got[0] != tc.first {
				t.Errorf("ExtractEmails(%q) first = %q, want %q", tc.text, got[0], tc.first)
			}
		})
	}
}
