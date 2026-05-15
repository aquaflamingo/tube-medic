package modules

import (
	"fmt"
	"log/slog"

	"github.com/aquaflamingo/ytleadbot/internal/config"
	"github.com/aquaflamingo/ytleadbot/internal/core"
	"github.com/aquaflamingo/ytleadbot/internal/utils"
)

func DiscoverChannels(db *core.DB, cfg *config.Config) (int, error) {
	var total int
	for _, kw := range cfg.Keywords {
		slog.Info("searching keyword", "keyword", kw)
		ids, err := utils.SearchKeywords(kw)
		if err != nil {
			slog.Warn("keyword search failed", "keyword", kw, "error", err)
			continue
		}

		slog.Info("keyword search returned channels", "keyword", kw, "count", len(ids))
		for _, id := range ids {
			exists, err := db.ChannelExists(id)
			if err != nil {
				slog.Warn("error checking channel", "id", id, "error", err)
				continue
			}
			if exists {
				slog.Debug("channel already in db, skipping", "id", id)
				continue
			}

			job := &core.Job{
				Type:    "enrich_channel",
				Payload: id,
			}
			if _, err := db.InsertJob(job); err != nil {
				slog.Warn("failed to queue enrich job", "id", id, "error", err)
				continue
			}
			slog.Info("queued channel for enrichment", "keyword", kw, "channel_id", id)
			total++
		}
	}
	return total, nil
}

func DiscoverByCategory(db *core.DB, cfg *config.Config, ytClient *utils.YoutubeClient) (int, error) {
	if ytClient == nil {
		slog.Info("no youtube api client, skipping category discovery")
		return 0, nil
	}

	rem, err := db.BudgetRemaining("youtube_api")
	if err != nil {
		slog.Warn("failed to check youtube api budget", "error", err)
	} else if rem < 100 {
		slog.Warn("youtube api budget nearly exhausted, skipping category discovery", "remaining", rem)
		return 0, nil
	}

	var total int
	for niche, catID := range cfg.NicheCategories {
		nicheQueued := 0
		slog.Info("searching category", "niche", niche, "category_id", catID)
		ids, err := ytClient.SearchByCategory(catID, 50)
		if err != nil {
			slog.Warn("category search failed", "niche", niche, "category_id", catID, "error", err)
			continue
		}
		if err := db.ConsumeBudget("youtube_api", 100, 10000); err != nil {
			slog.Warn("failed to record youtube api budget", "error", err)
		}

		slog.Info("category search returned channels", "niche", niche, "count", len(ids))
		for _, id := range ids {
			exists, err := db.ChannelExists(id)
			if err != nil {
				slog.Warn("error checking channel", "id", id, "error", err)
				continue
			}
			if exists {
				slog.Debug("channel already in db, skipping", "id", id)
				continue
			}

			job := &core.Job{
				Type:    "enrich_channel",
				Payload: id,
			}
			if _, err := db.InsertJob(job); err != nil {
				slog.Warn("failed to queue enrich job", "id", id, "error", err)
				continue
			}
			slog.Info("queued channel for enrichment", "niche", niche, "channel_id", id)
			nicheQueued++
			total++
		}
		slog.Info("category search complete", "niche", niche, "new_queued", nicheQueued)
	}
	return total, nil
}

func DiscoverRelatedFromEnriched(db *core.DB, cfg *config.Config) (int, error) {
	rows, err := db.Query(`SELECT id, handle FROM channels
		WHERE status IN ('enriched', 'enriched_no_email')
		AND enriched_at >= datetime('now', '-1 day')
		ORDER BY enriched_at DESC LIMIT 10`)
	if err != nil {
		return 0, fmt.Errorf("query enriched channels for related: %w", err)
	}
	defer rows.Close()

	type chanInfo struct {
		ID     string
		Handle string
	}
	var candidates []chanInfo
	for rows.Next() {
		var ci chanInfo
		if err := rows.Scan(&ci.ID, &ci.Handle); err != nil {
			continue
		}
		candidates = append(candidates, ci)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(candidates) == 0 {
		slog.Info("no recently enriched channels for related discovery")
		return 0, nil
	}

	var total int
	for _, ch := range candidates {
		slog.Info("finding related channels", "channel", ch.ID, "handle", ch.Handle)

		var relatedIDs []string

		channelURL := fmt.Sprintf("https://www.youtube.com/@%s/channels", ch.Handle)
		ids, err := utils.ExtractYouTubeIDsFromPage(channelURL)
		if err == nil && len(ids) > 0 {
			relatedIDs = ids
		}

		if len(relatedIDs) == 0 {
			slog.Debug("no related channels found", "handle", ch.Handle)
			continue
		}

		slog.Info("found related channels", "handle", ch.Handle, "count", len(relatedIDs))
		for _, id := range relatedIDs {
			if id == ch.ID {
				continue
			}
			exists, err := db.ChannelExists(id)
			if err != nil {
				slog.Warn("error checking channel", "id", id, "error", err)
				continue
			}
			if exists {
				slog.Debug("related channel already in db, skipping", "id", id)
				continue
			}
			job := &core.Job{
				Type:    "enrich_channel",
				Payload: id,
			}
			if _, err := db.InsertJob(job); err != nil {
				slog.Warn("failed to queue related channel", "id", id, "error", err)
				continue
			}
			slog.Info("queued related channel for enrichment",
				"from_handle", ch.Handle, "channel_id", id)
			total++
		}
	}

	slog.Info("related channel discovery complete", "channels_processed", len(candidates), "new_queued", total)
	return total, nil
}
