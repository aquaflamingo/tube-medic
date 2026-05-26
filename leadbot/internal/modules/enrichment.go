package modules

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/aqfl/tmleadbot/internal/config"
	"github.com/aqfl/tmleadbot/internal/core"
	"github.com/aqfl/tmleadbot/internal/utils"
)

func EnrichChannel(db *core.DB, cfg *config.Config, ident string) error {
	slog.Info("enriching channel", "ident", ident)

	info, err := utils.GetChannelInfo(ident)
	if err != nil {
		return fmt.Errorf("fetch channel info: %w", err)
	}

	if info.ChannelFollowerCount < cfg.MinSubs || info.ChannelFollowerCount > cfg.MaxSubs {
		slog.Info("channel outside subscriber range, skipping",
			"channel", info.Channel,
			"subs", info.ChannelFollowerCount,
			"range", fmt.Sprintf("%d-%d", cfg.MinSubs, cfg.MaxSubs))
		return nil
	}

	videos, err := utils.GetRecentVideos(info.ChannelID, 10)
	if err != nil {
		slog.Warn("failed to get recent videos", "channel", info.Channel, "error", err)
		videos = nil
	}

	links := utils.ExtractLinksFromDescription(info.Description)

	hasPricing := false
	if info.WebpageURL != "" {
		hasPricing = utils.HasPricingPage(info.WebpageURL)
	}

	score := utils.CalculateScore(info, videos, links, hasPricing)

	email := info.Email
	if email == "" {
		email = info.ChannelEmail
	}
	if email == "" && info.WebpageURL != "" {
		slog.Info("no email from yt-dlp, scraping website", "channel", info.Channel)
		emails := utils.ExtractFromWebsite(info.WebpageURL)
		if len(emails) > 0 {
			email = emails[0]
		}
	}
	if email == "" {
		for _, l := range links {
			if utils.IsLinktreeURL(l) {
				slog.Info("no email yet, scraping linktree", "url", l)
				emails := utils.ExtractFromLinktree(l)
				if len(emails) > 0 {
					email = emails[0]
					break
				}
			}
		}
	}

	signalsJSON := ""
	if len(score.Signals) > 0 {
		signalsStr := "["
		for i, s := range score.Signals {
			if i > 0 {
				signalsStr += ","
			}
			signalsStr += fmt.Sprintf("%q", s)
		}
		signalsStr += "]"
		signalsJSON = signalsStr
	}

	status := "enriched"
	if email == "" {
		status = "enriched_no_email"
	}

	ch := &core.Channel{
		ID:                  info.ChannelID,
		Handle:              info.UploaderID,
		Name:                info.Channel,
		SubscriberCount:     info.ChannelFollowerCount,
		ViewCount:           info.ViewCount,
		VideoCount:          info.PlaylistCount,
		UploadFrequency:     utils.CalculateUploadFrequency(videos),
		Description:         info.Description,
		Country:             info.ChannelCountry,
		Email:               email,
		Website:             info.WebpageURL,
		MonetizationScore:   score.Score,
		MonetizationTier:    score.Tier,
		MonetizationSignals: signalsJSON,
		Status:              status,
		DiscoveryKeyword:    "manual",
	}

	if score.Score < cfg.MinScore {
		ch.Status = "low_score"
		ch.MonetizationScore = score.Score
	}

	if err := db.UpsertChannel(ch); err != nil {
		return fmt.Errorf("upsert channel: %w", err)
	}
	slog.Info("channel saved to database", "name", info.Channel, "id", info.ChannelID)

	tierStr := score.Tier
	if tierStr == "" {
		tierStr = "standard"
	}

	slog.Info("channel enriched",
		"name", info.Channel,
		"subs", info.ChannelFollowerCount,
		"score", score.Score,
		"tier", tierStr,
		"email", email != "",
		"status", ch.Status)

	if err := db.ConsumeBudget("yt_dlp", 1, cfg.MaxEnrichments); err != nil {
		slog.Warn("failed to record yt-dlp budget", "error", err)
	}

	time.Sleep(time.Duration(cfg.RequestDelay) * time.Second)

	return nil
}
