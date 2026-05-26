package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type MonetizationResult struct {
	Score   int
	Tier    string
	Signals []string
}

type sponsorBlockSegment struct {
	VideoID  string  `json:"videoID"`
	Category string  `json:"category"`
	Segment  []float64 `json:"segment"`
}

func CheckSponsorBlock(videoID string) (bool, error) {
	url := fmt.Sprintf("https://sponsor.ajay.app/api/searchSegments?videoID=%s&categories=[%%22sponsor%%22]", videoID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("create sponsorblock request: %w", err)
	}
	req.Header.Set("User-Agent", "YTLeadBot/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("sponsorblock request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	}
	if resp.StatusCode != 200 {
		return false, nil
	}

	var segments []sponsorBlockSegment
	if err := json.NewDecoder(resp.Body).Decode(&segments); err != nil {
		return false, nil
	}

	return len(segments) > 0, nil
}

func DetectPlatforms(links []string) map[string]bool {
	platforms := make(map[string]bool)
	for _, l := range links {
		switch {
		case IsPatreonURL(l):
			platforms["patreon"] = true
		case IsCourseURL(l):
			platforms["course"] = true
		case IsMerchURL(l):
			platforms["merch"] = true
		case IsAmazonURL(l):
			platforms["amazon"] = true
		case IsLinktreeURL(l):
			platforms["linktree"] = true
		}
	}
	return platforms
}

func CalculateScore(info *YTDLPMetadata, videos []YTDLPVideo, links []string, hasPricing bool) *MonetizationResult {
	result := &MonetizationResult{}
	var signals []string

	email := info.Email
	if email == "" {
		email = info.ChannelEmail
	}
	if email != "" {
		result.Score += 3
		signals = append(signals, "email_present")
	}

	for _, v := range videos {
		hasSponsors, err := CheckSponsorBlock(v.ID)
		if err == nil && hasSponsors {
			result.Score += 3
			signals = append(signals, "sponsorships")
			break
		}
	}

	platforms := DetectPlatforms(links)
	if platforms["patreon"] {
		result.Score += 2
		signals = append(signals, "patreon_link")
	}
	if platforms["course"] {
		result.Score += 2
		signals = append(signals, "course_link")
	}
	if platforms["merch"] {
		result.Score += 1
		signals = append(signals, "merch_link")
	}
	if platforms["amazon"] {
		result.Score += 1
		signals = append(signals, "amazon_link")
	}
	if platforms["linktree"] {
		result.Score += 1
		signals = append(signals, "linktree_link")
	}

	freq := CalculateUploadFrequency(videos)
	if freq >= 2.0 {
		result.Score += 1
		signals = append(signals, "active_uploader")
	}

	if info.ChannelIsMembership {
		result.Score += 1
		signals = append(signals, "channel_membership")
	}

	if hasPricing {
		result.Score += 1
		signals = append(signals, "pricing_page")
	}

	if result.Score >= 8 {
		result.Tier = "high"
	} else {
		result.Tier = "standard"
	}

	result.Signals = signals
	return result
}

func ExtractLinksFromDescription(desc string) []string {
	if desc == "" {
		return nil
	}
	var links []string
	words := strings.Fields(desc)
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?\"'()[]{}")
		if strings.HasPrefix(w, "http://") || strings.HasPrefix(w, "https://") {
			links = append(links, w)
		}
	}
	return links
}
