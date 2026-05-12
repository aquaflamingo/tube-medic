package report_test

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/aquaflamingo/tubemedicmvp/internal/checker"
	"github.com/aquaflamingo/tubemedicmvp/internal/report"
	"github.com/aquaflamingo/tubemedicmvp/internal/youtube"
)

var update = flag.Bool("update", false, "update golden files")

func TestPrint_noBrokenLinks(t *testing.T) {
	ch := &youtube.Channel{ID: "UC_test", Name: "Clean Channel"}
	videos := []youtube.Video{
		{ID: "v1", Title: "Only Working Links"},
	}
	summary := checker.Summary{
		Total:       3,
		Broken:      0,
		Live:        3,
		BrokenLinks: nil,
	}

	got := captureOutput(func() {
		report.Print(ch, videos, summary)
	})

	goldenPath := filepath.Join("testdata", "no_broken_links.golden")
	if *update {
		os.WriteFile(goldenPath, []byte(got), 0644)
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}
	if got != string(want) {
		t.Errorf("output mismatch\n--- got:\n%s\n--- want:\n%s", got, string(want))
	}
}

func TestPrint_withBrokenLinks(t *testing.T) {
	ch := &youtube.Channel{ID: "UC_test", Name: "Broken Channel"}
	videos := []youtube.Video{
		{ID: "v1", Title: "First Video"},
		{ID: "v2", Title: "Second Video"},
	}
	summary := checker.Summary{
		Total:  4,
		Broken: 2,
		Live:   2,
		BrokenLinks: []checker.CheckResult{
			{URL: "https://example.com/404", VideoID: "v1", VideoTitle: "First Video", Status: checker.StatusBroken, Error: "404"},
			{URL: "https://example.com/timeout", VideoID: "v2", VideoTitle: "Second Video", Status: checker.StatusBroken, Error: "TIMEOUT"},
		},
	}

	got := captureOutput(func() {
		report.Print(ch, videos, summary)
	})

	goldenPath := filepath.Join("testdata", "with_broken_links.golden")
	if *update {
		os.WriteFile(goldenPath, []byte(got), 0644)
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}
	if got != string(want) {
		t.Errorf("output mismatch\n--- got:\n%s\n--- want:\n%s", got, string(want))
	}
}

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
