package utils

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YoutubeClient struct {
	service       *youtube.Service
	apiKey        string
	quotaUsed     int
	quotaReset    time.Time
	maxDailyQuota int
}

func NewYoutubeClient(apiKey string) (*YoutubeClient, error) {
	ctx := context.Background()
	svc, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}
	now := time.Now()
	reset := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	if reset.Before(now) {
		reset = reset.Add(24 * time.Hour)
	}
	return &YoutubeClient{
		service:       svc,
		apiKey:        apiKey,
		maxDailyQuota: 10000,
		quotaReset:    reset,
	}, nil
}

func (c *YoutubeClient) quotaOK(units int) bool {
	now := time.Now()
	if now.After(c.quotaReset) {
		c.quotaUsed = 0
		c.quotaReset = now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	}
	return c.quotaUsed+units <= c.maxDailyQuota
}

func (c *YoutubeClient) RemainingQuota() int {
	return c.maxDailyQuota - c.quotaUsed
}

func (c *YoutubeClient) SearchByCategory(categoryID string, maxResults int64) ([]string, error) {
	if !c.quotaOK(100) {
		return nil, fmt.Errorf("daily quota exhausted (%d/%d)", c.quotaUsed, c.maxDailyQuota)
	}

	call := c.service.Search.List([]string{"snippet"}).
		Type("channel").
		VideoCategoryId(categoryID).
		MaxResults(maxResults)

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("search by category %s: %w", categoryID, err)
	}

	c.quotaUsed += 100
	slog.Debug("youtube api search by category", "category", categoryID, "results", len(resp.Items), "quota_used", c.quotaUsed)

	var ids []string
	for _, item := range resp.Items {
		if item.Id != nil && item.Id.ChannelId != "" {
			ids = append(ids, item.Id.ChannelId)
		}
	}
	return ids, nil
}

func (c *YoutubeClient) GetChannelDetails(channelIDs []string) ([]*youtube.Channel, error) {
	if len(channelIDs) == 0 {
		return nil, nil
	}
	cost := len(channelIDs)
	if !c.quotaOK(cost) {
		return nil, fmt.Errorf("daily quota exhausted (%d/%d)", c.quotaUsed, c.maxDailyQuota)
	}

	call := c.service.Channels.List([]string{"snippet", "statistics"}).
		Id(channelIDs...)

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("get channel details: %w", err)
	}

	c.quotaUsed += cost
	slog.Debug("youtube api channel details", "channels", len(channelIDs), "quota_used", c.quotaUsed)

	return resp.Items, nil
}
