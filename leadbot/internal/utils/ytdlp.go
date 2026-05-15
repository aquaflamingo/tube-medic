package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type YTDLPMetadata struct {
	ID                   string `json:"id"`
	ChannelID            string `json:"channel_id"`
	Channel              string `json:"channel"`
	ChannelURL           string `json:"channel_url"`
	ChannelFollowerCount int64  `json:"channel_follower_count"`
	Description          string `json:"description"`
	ViewCount            int64  `json:"view_count"`
	PlaylistCount        int64  `json:"playlist_count"`
	ChannelIsVerified    bool   `json:"channel_is_verified"`
	ChannelIsMembership  bool   `json:"channel_is_membership"`
	UploaderID           string `json:"uploader_id"`
	ChannelCountry       string `json:"channel_country"`
	Email                string `json:"email"`
	ChannelEmail         string `json:"channel_email"`
	UploadDate           string `json:"upload_date"`
	Extractor            string `json:"_type"`
	WebpageURL           string `json:"webpage_url"`
	RequestedFormats     []struct {
		URL string `json:"url"`
	} `json:"requested_formats"`
}

type YTDLPVideo struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	UploadDate string `json:"upload_date"`
	ViewCount  int64  `json:"view_count"`
}

type YTDLPSearchResult struct {
	ID         string `json:"id"`
	ChannelID  string `json:"channel_id"`
	Channel    string `json:"channel"`
	Uploader   string `json:"uploader"`
	Title      string `json:"title"`
	WebpageURL string `json:"webpage_url"`
}

func ytdlpPath() (string, error) {
	p, err := ensureYTDLPExtracted()
	if err == nil {
		return p, nil
	}
	return exec.LookPath("yt-dlp")
}

func ytdlp(args ...string) ([]byte, error) {
	path, err := ytdlpPath()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp not found: %w", err)
	}
	cmd := exec.Command(path, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp failed: %w\nstderr: %s", err, string(ee.Stderr))
		}
		return nil, fmt.Errorf("yt-dlp failed: %w", err)
	}
	return out, nil
}

func buildChannelURL(ident string) string {
	ident = strings.TrimSpace(ident)
	ident = strings.TrimPrefix(ident, "@")
	if strings.HasPrefix(ident, "UC") && len(ident) == 24 {
		return fmt.Sprintf("https://www.youtube.com/channel/%s/about", ident)
	}
	return fmt.Sprintf("https://www.youtube.com/@%s/about", ident)
}

func buildVideoURL(ident string) string {
	ident = strings.TrimSpace(ident)
	ident = strings.TrimPrefix(ident, "@")
	if strings.HasPrefix(ident, "UC") && len(ident) == 24 {
		return fmt.Sprintf("https://www.youtube.com/channel/%s/videos", ident)
	}
	return fmt.Sprintf("https://www.youtube.com/@%s/videos", ident)
}

func GetChannelInfo(ident string) (*YTDLPMetadata, error) {
	url := buildChannelURL(ident)
	// Use --playlist-end 1 to get just the channel info (first entry)
	// Without --flat-playlist to get full metadata including email
	out, err := ytdlp("--dump-json", "--playlist-end", "1", "--no-warnings", url)
	if err != nil {
		return nil, fmt.Errorf("get channel info: %w", err)
	}

	lines := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)
	if len(lines) == 0 || lines[0] == "" {
		return nil, fmt.Errorf("empty output from yt-dlp for %s", url)
	}

	var info YTDLPMetadata
	if err := json.Unmarshal([]byte(lines[0]), &info); err != nil {
		return nil, fmt.Errorf("parse yt-dlp output: %w", err)
	}

	if info.ChannelID == "" {
		info.ChannelID = info.ID
	}

	return &info, nil
}

func SearchKeywords(query string) ([]string, error) {
	searchStr := fmt.Sprintf("ytsearch20:%s", query)
	out, err := ytdlp("--dump-json", "--flat-playlist", "--no-warnings", "--no-playlist", searchStr)
	if err != nil {
		return nil, fmt.Errorf("search keywords: %w", err)
	}

	var results []string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var r YTDLPSearchResult
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			continue
		}
		if r.ChannelID != "" {
			results = append(results, r.ChannelID)
		} else if r.ID != "" {
			results = append(results, r.ID)
		}
	}
	return results, nil
}

func GetRecentVideos(ident string, maxResults int) ([]YTDLPVideo, error) {
	url := buildVideoURL(ident)
	// Without --flat-playlist to get upload_date
	limit := fmt.Sprintf("--playlist-end=%d", maxResults)
	out, err := ytdlp("--dump-json", "--no-warnings", limit, url)
	if err != nil {
		return nil, fmt.Errorf("get recent videos: %w", err)
	}

	var videos []YTDLPVideo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var v YTDLPVideo
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			continue
		}
		videos = append(videos, v)
	}
	return videos, nil
}

func CalculateUploadFrequency(videos []YTDLPVideo) float64 {
	if len(videos) < 2 {
		return 0
	}

	var dates []time.Time
	for _, v := range videos {
		if v.UploadDate == "" {
			continue
		}
		t, err := time.Parse("20060102", v.UploadDate)
		if err != nil {
			continue
		}
		dates = append(dates, t)
	}
	if len(dates) < 2 {
		return 0
	}

	minTime := dates[0]
	maxTime := dates[0]
	for _, d := range dates {
		if d.Before(minTime) {
			minTime = d
		}
		if d.After(maxTime) {
			maxTime = d
		}
	}

	months := maxTime.Sub(minTime).Hours() / (24.0 * 30.0)
	if months < 0.5 {
		months = 0.5
	}

	return float64(len(dates)) / months
}
