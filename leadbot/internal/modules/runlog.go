package modules

import (
	"log/slog"
	"time"

	"github.com/aquaflamingo/ytleadbot/internal/core"
)

func startRun(db *core.DB, runType string) int64 {
	now := time.Now().UTC().Format(time.RFC3339)
	id, err := db.InsertRun(&core.Run{RunType: runType, StartedAt: now})
	if err != nil {
		slog.Warn("failed to start run log", "type", runType, "error", err)
		return 0
	}
	return id
}

func finishRun(db *core.DB, id int64, discovered, enriched, agencies, exported, failed int) {
	if id == 0 {
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	emails, err := db.CountByStatus("enriched")
	if err != nil {
		slog.Warn("failed to count enriched channels", "error", err)
	}
	if err := db.CompleteRun(id, &core.Run{
		CompletedAt:        now,
		ChannelsDiscovered: discovered,
		ChannelsEnriched:   enriched,
		EmailsFound:        emails,
		AgenciesFound:      agencies,
		LeadsExported:      exported,
		Errors:             failed,
	}); err != nil {
		slog.Warn("failed to complete run log", "id", id, "error", err)
	}
}
