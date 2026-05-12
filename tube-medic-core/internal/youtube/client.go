package youtube

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://www.googleapis.com/youtube/v3"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// YouTube Data API v3 estimated costs per request type.
const (
	costChannels     = 1
	costSearch       = 100
	costVideos       = 1
)

func getJSON(url string, dest interface{}, estimatedCost int, q *Quota) error {
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d", resp.StatusCode)
	}

	if q != nil {
		cost := estimatedCost
		if h := resp.Header.Get("X-Quota-Cost"); h != "" {
			if c, err := strconv.Atoi(h); err == nil {
				cost = c
			}
		}
		q.Used += cost
		if h := resp.Header.Get("X-Quota-Remaining"); h != "" {
			if r, err := strconv.Atoi(h); err == nil {
				q.Remaining = r
			}
		}
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

// ResolveChannel takes a channel URL and returns the channel info.
// Supports:
//
//	https://youtube.com/channel/UCxxx
//	https://youtube.com/@handle
//	https://youtube.com/c/customname
func ResolveChannel(apiKey, channelURL string, q *Quota) (*Channel, error) {
	u, err := url.Parse(channelURL)
	if err != nil {
		return nil, fmt.Errorf("invalid channel URL: %w", err)
	}

	// Normalize host — allow www.youtube.com, youtube.com, m.youtube.com
	host := strings.ToLower(u.Host)
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "m.")
	if host != "youtube.com" {
		return nil, fmt.Errorf("not a youtube.com URL: %s", channelURL)
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(segments) == 0 {
		return nil, fmt.Errorf("no path in URL: %s", channelURL)
	}

	switch segments[0] {
	case "channel":
		if len(segments) < 2 {
			return nil, fmt.Errorf("missing channel ID in URL: %s", channelURL)
		}
		channelID := segments[1]
		return fetchChannelByID(apiKey, channelID, q)

	case "@":
		// URL like youtube.com/@handle — segments[0] is "@" when the path is "/@"
		return fetchChannelByHandle(apiKey, segments[0], q)
	default:
		if strings.HasPrefix(segments[0], "@") {
			return fetchChannelByHandle(apiKey, segments[0], q)
		}
		// Try as a custom handle (for /c/ paths)
		return fetchChannelByHandle(apiKey, "@"+segments[0], q)
	}
}

func fetchChannelByID(apiKey, channelID string, q *Quota) (*Channel, error) {
	u := fmt.Sprintf("%s/channels?part=snippet&id=%s&key=%s", baseURL, url.QueryEscape(channelID), apiKey)
	var resp apiChannelResponse
	if err := getJSON(u, &resp, costChannels, q); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}
	return &Channel{ID: resp.Items[0].ID, Name: resp.Items[0].Snippet.Title}, nil
}

func fetchChannelByHandle(apiKey, handle string, q *Quota) (*Channel, error) {
	u := fmt.Sprintf("%s/channels?part=snippet&forHandle=%s&key=%s", baseURL, url.QueryEscape(handle), apiKey)
	var resp apiChannelResponse
	if err := getJSON(u, &resp, costChannels, q); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("channel not found for handle: %s", handle)
	}
	return &Channel{ID: resp.Items[0].ID, Name: resp.Items[0].Snippet.Title}, nil
}

// FetchVideos returns the most recent videos for a channel.
func FetchVideos(apiKey, channelID string, maxResults int, q *Quota) ([]Video, error) {
	u := fmt.Sprintf("%s/search?part=snippet&channelId=%s&order=date&maxResults=%d&key=%s",
		baseURL, url.QueryEscape(channelID), maxResults, apiKey)

	var resp apiSearchResponse
	if err := getJSON(u, &resp, costSearch, q); err != nil {
		return nil, err
	}

	videos := make([]Video, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item.ID.VideoID == "" {
			continue
		}
		videos = append(videos, Video{
			ID:    item.ID.VideoID,
			Title: item.Snippet.Title,
		})
	}
	return videos, nil
}

// FetchDescriptions takes video IDs and returns a map of ID -> full description.
func FetchDescriptions(apiKey string, videoIDs []string, q *Quota) (map[string]string, error) {
	if len(videoIDs) == 0 {
		return map[string]string{}, nil
	}

	// YouTube API accepts up to 50 IDs per request, batch if needed
	const batchSize = 50
	descriptions := make(map[string]string, len(videoIDs))

	for i := 0; i < len(videoIDs); i += batchSize {
		end := i + batchSize
		if end > len(videoIDs) {
			end = len(videoIDs)
		}
		batch := videoIDs[i:end]

		ids := make([]string, len(batch))
		for i, id := range batch {
			ids[i] = id
		}

		u := fmt.Sprintf("%s/videos?part=snippet&id=%s&key=%s",
			baseURL, strings.Join(batch, ","), apiKey)

		var resp apiVideosResponse
		if err := getJSON(u, &resp, costVideos, q); err != nil {
			return nil, fmt.Errorf("fetching descriptions: %w", err)
		}

		for _, item := range resp.Items {
			descriptions[item.ID] = item.Snippet.Description
		}
	}

	return descriptions, nil
}

// FetchChannel resolves a channel URL and returns both the channel and its videos with descriptions.
func FetchChannel(apiKey, channelURL string, maxResults int) (*Channel, []Video, *Quota, error) {
	q := &Quota{Remaining: -1}

	ch, err := ResolveChannel(apiKey, channelURL, q)
	if err != nil {
		return nil, nil, q, fmt.Errorf("resolving channel: %w", err)
	}

	videos, err := FetchVideos(apiKey, ch.ID, maxResults, q)
	if err != nil {
		return nil, nil, q, fmt.Errorf("fetching videos: %w", err)
	}

	videoIDs := make([]string, len(videos))
	for i, v := range videos {
		videoIDs[i] = v.ID
	}

	descs, err := FetchDescriptions(apiKey, videoIDs, q)
	if err != nil {
		return nil, nil, q, fmt.Errorf("fetching descriptions: %w", err)
	}

	for i, v := range videos {
		videos[i].Description = descs[v.ID]
	}

	return ch, videos, q, nil
}

// ParseChannelURL extracts the meaningful path segment(s) from a YouTube channel URL.
// Exported for testing.
func ParseChannelURL(rawURL string) (kind string, value string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(segments) == 0 {
		return "", "", fmt.Errorf("no path in URL")
	}
	if segments[0] == "channel" && len(segments) > 1 {
		return "channel", segments[1], nil
	}
	if strings.HasPrefix(segments[0], "@") {
		return "handle", segments[0], nil
	}
	return "handle", "@" + segments[0], nil
}

// VideoURL returns the YouTube watch URL for a video ID.
func VideoURL(videoID string) string {
	return "https://youtube.com/watch?v=" + videoID
}


