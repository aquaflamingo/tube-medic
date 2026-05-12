package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aquaflamingo/tubemedicmvp/internal/checker"
	"github.com/aquaflamingo/tubemedicmvp/internal/config"
	"github.com/aquaflamingo/tubemedicmvp/internal/report"
	"github.com/aquaflamingo/tubemedicmvp/internal/scraper"
	"github.com/aquaflamingo/tubemedicmvp/internal/youtube"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	// 1. Fetch channel info, videos, and descriptions
	fmt.Fprintf(os.Stderr, "Fetching channel...\n")
	ch, videos, quota, err := youtube.FetchChannel(cfg.APIKey, cfg.ChannelURL, cfg.MaxVideos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Found %q — scanning %d videos...\n", ch.Name, len(videos))
	fmt.Fprintf(os.Stderr, "API quota used: %d units", quota.Used)
	if quota.Remaining >= 0 {
		fmt.Fprintf(os.Stderr, " (%d remaining today)", quota.Remaining)
	}
	fmt.Fprintf(os.Stderr, "\n")

	// 2. Extract URLs from descriptions
	links := scraper.ExtractAll(videos)
	fmt.Fprintf(os.Stderr, "Extracted %d unique links\n", len(links))

	// 3. Check all links concurrently
	results := checker.CheckAll(links, 10)
	summary := checker.Summarize(results)

	// 4. Print report
	out := io.Writer(os.Stdout)
	if cfg.OutputFile != "" {
		f, err := os.Create(cfg.OutputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = io.MultiWriter(os.Stdout, f)
	}
	report.Print(out, ch, videos, summary, *quota)

	if summary.Broken > 0 {
		os.Exit(1)
	}
}
