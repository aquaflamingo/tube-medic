package modules

import (
	"log/slog"

	"github.com/aqfl/tmleadbot/internal/config"
	"github.com/aqfl/tmleadbot/internal/core"
	"github.com/aqfl/tmleadbot/internal/utils"
)

func RunDiscoverPipeline(db *core.DB, cfg *config.Config, ytClient *utils.YoutubeClient) (disc int, enr int, fail int, rerr error) {
	runID := startRun(db, "discover")
	defer func() {
		finishRun(db, runID, disc, enr, 0, 0, fail)
	}()

	// 1. Keyword search via yt-dlp
	fromKeywords, err := DiscoverChannels(db, cfg)
	if err != nil {
		rerr = err
		return
	}
	disc += fromKeywords
	slog.Info("keyword discovery complete", "queued", fromKeywords)

	// 2. YouTube API category search
	fromCategories, err := DiscoverByCategory(db, cfg, ytClient)
	if err != nil {
		slog.Warn("category discovery failed", "error", err)
	}
	disc += fromCategories
	slog.Info("category discovery complete", "queued", fromCategories)

	// 3. Enrichment pass for keyword + category discoveries
	processed, failed, err := ProcessJobQueue(db, cfg, "enrich_channel")
	if err != nil {
		rerr = err
		enr = processed
		fail = failed
		return
	}
	enr = processed
	fail = failed
	slog.Info("first enrichment pass complete", "processed", processed, "failed", failed)

	// 4. Related channels from enriched channels
	fromRelated, err := DiscoverRelatedFromEnriched(db, cfg)
	if err != nil {
		slog.Warn("related channel discovery failed", "error", err)
	}
	disc += fromRelated
	slog.Info("related channel discovery complete", "queued", fromRelated)

	// 5. Second enrichment pass for related channels
	if fromRelated > 0 {
		p2, f2, err2 := ProcessJobQueue(db, cfg, "enrich_channel")
		if err2 != nil {
			rerr = err2
			enr += p2
			fail += f2
			return
		}
		enr += p2
		fail += f2
		slog.Info("second enrichment pass complete", "processed", p2, "failed", f2)
	}

	return
}

func RunReenrichStale(db *core.DB, maxAgeDays int) (int, error) {
	ids, err := db.GetStaleChannelIDs(maxAgeDays)
	if err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		slog.Info("no stale channels to re-enrich")
		return 0, nil
	}

	slog.Info("queuing stale channels for re-enrichment", "count", len(ids))
	for _, id := range ids {
		job := &core.Job{
			Type:    "enrich_channel",
			Payload: id,
		}
		if _, err := db.InsertJob(job); err != nil {
			slog.Warn("failed to queue re-enrich job", "id", id, "error", err)
		}
	}

	return len(ids), nil
}
