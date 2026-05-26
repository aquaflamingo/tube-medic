// Package tmcore is the public API for scanning YouTube video descriptions
// for broken or suspicious links.
package tmcore

import (
	"fmt"

	"github.com/aquaflamingo/tmcore/internal/checker"
	"github.com/aquaflamingo/tmcore/internal/scraper"
	"github.com/aquaflamingo/tmcore/internal/youtube"
)

// Config holds scan parameters.
type Config struct {
	APIKey     string
	ChannelURL string
	MaxVideos  int
}

// Report holds structured scan results.
type Report struct {
	Channel *youtube.Channel
	Videos  []youtube.Video
	Results []checker.CheckResult
	Summary checker.Summary
	Quota   youtube.Quota
}

// RunScan executes the full scan pipeline: fetches channel info and videos,
// extracts links from descriptions, checks each link, and returns a summary report.
func RunScan(cfg Config) (*Report, error) {
	ch, videos, quota, err := youtube.FetchChannel(cfg.APIKey, cfg.ChannelURL, cfg.MaxVideos)
	if err != nil {
		return nil, fmt.Errorf("fetching channel: %w", err)
	}

	links := scraper.ExtractAll(videos)
	results := checker.CheckAll(links, 10)
	summary := checker.Summarize(results)

	return &Report{
		Channel: ch,
		Videos:  videos,
		Results: results,
		Summary: summary,
		Quota:   *quota,
	}, nil
}
