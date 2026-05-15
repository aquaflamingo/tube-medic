package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aquaflamingo/ytleadbot/internal/config"
	"github.com/aquaflamingo/ytleadbot/internal/core"
	"github.com/aquaflamingo/ytleadbot/internal/modules"
	"github.com/aquaflamingo/ytleadbot/internal/utils"
)

func main() {
	job := flag.String("job", "", "Job type: enrich, export, discover, agency_scan, reenrich_stale")
	channel := flag.String("channel", "", "Channel ID or @handle (optional for enrich in batch mode)")
	logLevel := flag.String("log", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	var level slog.Level
	switch strings.ToLower(*logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		slog.Error("failed to create data dir", "error", err)
		os.Exit(1)
	}

	db, err := core.NewDB(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.ResetStaleJobs()

	var ytClient *utils.YoutubeClient
	if cfg.YTApiKey != "" {
		client, err := utils.NewYoutubeClient(cfg.YTApiKey)
		if err != nil {
			slog.Warn("failed to create youtube api client", "error", err)
		} else {
			ytClient = client
			slog.Info("youtube api client initialized", "daily_quota", 10000)
		}
	} else {
		slog.Info("no YT_API_KEY set, category search and related channels disabled")
	}

	switch *job {
	case "enrich":
		if *channel != "" {
			start := time.Now()
			slog.Info("starting enrich job", "channel", *channel)
			if err := modules.EnrichChannel(db, cfg, *channel); err != nil {
				slog.Error("enrich failed", "channel", *channel, "error", err)
				os.Exit(1)
			}
			slog.Info("enrich complete", "channel", *channel, "duration", time.Since(start).Round(time.Second))
		} else {
			start := time.Now()
			slog.Info("starting batch enrich")
			runStart := time.Now().UTC().Format(time.RFC3339)
			runID, _ := db.InsertRun(&core.Run{RunType: "enrich", StartedAt: runStart})
			processed, failed, err := modules.ProcessJobQueue(db, cfg, "enrich_channel")
			if err != nil {
				slog.Error("batch enrich failed", "error", err)
				os.Exit(1)
			}
			if runID != 0 {
				now := time.Now().UTC().Format(time.RFC3339)
				emails, _ := db.CountByStatus("enriched")
				db.CompleteRun(runID, &core.Run{CompletedAt: now, ChannelsEnriched: processed, EmailsFound: emails, Errors: failed})
			}
			slog.Info("batch enrich complete", "processed", processed, "failed", failed, "duration", time.Since(start).Round(time.Second))
		}

	case "export":
		start := time.Now()
		slog.Info("starting export job")
		runStart := time.Now().UTC().Format(time.RFC3339)
		runID, _ := db.InsertRun(&core.Run{RunType: "export", StartedAt: runStart})
		count, err := modules.ExportLeads(db, cfg.ExportDir)
		if err != nil {
			slog.Error("export failed", "error", err)
			os.Exit(1)
		}
		if runID != 0 {
			now := time.Now().UTC().Format(time.RFC3339)
			emails, _ := db.CountByStatus("enriched")
			db.CompleteRun(runID, &core.Run{CompletedAt: now, LeadsExported: count, EmailsFound: emails})
		}
		slog.Info("export complete", "leads", count, "duration", time.Since(start).Round(time.Second))

	case "discover":
		start := time.Now()
		slog.Info("starting discover pipeline")
		discovered, enriched, failed, err := modules.RunDiscoverPipeline(db, cfg, ytClient)
		if err != nil {
			slog.Error("discover pipeline failed", "error", err)
			os.Exit(1)
		}
		slog.Info("discover pipeline complete",
			"discovered", discovered,
			"enriched", enriched,
			"failed", failed,
			"duration", time.Since(start).Round(time.Second))

	case "agency_scan":
		start := time.Now()
		slog.Info("starting agency scan pipeline")
		agencies, channels, err := modules.RunAgencyScan(db, cfg)
		if err != nil {
			slog.Error("agency scan failed", "error", err)
			os.Exit(1)
		}
		slog.Info("agency scan complete",
			"agencies", agencies,
			"channels", channels,
			"duration", time.Since(start).Round(time.Second))

	case "reenrich_stale":
		start := time.Now()
		slog.Info("starting reenrich stale job")
		queued, err := modules.RunReenrichStale(db, 30)
		if err != nil {
			slog.Error("reenrich stale failed", "error", err)
			os.Exit(1)
		}
		slog.Info("reenrich stale complete", "queued", queued, "duration", time.Since(start).Round(time.Second))

		if queued > 0 {
			processed, failed, err := modules.ProcessJobQueue(db, cfg, "enrich_channel")
			if err != nil {
				slog.Error("reenrich enrichment failed", "error", err)
				os.Exit(1)
			}
			slog.Info("reenrich enrichment complete", "processed", processed, "failed", failed)
		}

	default:
		fmt.Print(`YTLeadBot - YouTube Creator Lead Finder

Usage:
  ytleadbot --job discover                        Search keywords, queue & enrich channels
  ytleadbot --job enrich [--channel @handle]      Process enrichment queue or single channel
  ytleadbot --job export                          Export enriched leads to JSON
  ytleadbot --job agency_scan                     Scan agency rosters + detect agencies
  ytleadbot --job reenrich_stale                  Re-enrich channels last enriched >30 days ago

Flags:
  --job      Job type: enrich, export, discover, agency_scan, reenrich_stale
  --channel  Channel ID (UCxxxx) or @handle (optional for enrich in batch mode)
  --log      Log level: debug, info, warn, error (default: info)
`)
	}
}
