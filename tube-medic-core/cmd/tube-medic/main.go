package main

import (
	"fmt"
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
	ch, videos, err := youtube.FetchChannel(cfg.APIKey, cfg.ChannelURL, cfg.MaxVideos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching channel: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Found %q — scanning %d videos...\n", ch.Name, len(videos))

	// 2. Extract URLs from descriptions
	links := scraper.ExtractAll(videos)
	fmt.Fprintf(os.Stderr, "Extracted %d unique links\n", len(links))

	// 3. Check all links concurrently
	results := checker.CheckAll(links, 10)
	summary := checker.Summarize(results)

	// 4. Print report to stdout
	report.Print(ch, videos, summary)

	// 5. Exit with code 1 if any broken links found
	if summary.Broken > 0 {
		os.Exit(1)
	}
}
