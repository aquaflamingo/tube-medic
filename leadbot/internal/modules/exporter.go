package modules

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aquaflamingo/ytleadbot/internal/core"
)

type ExportLead struct {
	ChannelID           string   `json:"channel_id"`
	Name                string   `json:"name"`
	Handle              string   `json:"handle"`
	Email               string   `json:"email"`
	SubscriberCount     int64    `json:"subscriber_count"`
	Niche               string   `json:"niche"`
	MonetizationScore   int      `json:"monetization_score"`
	MonetizationSignals []string `json:"monetization_signals"`
	Website             string   `json:"website"`
	Agency              *string  `json:"agency"`
	AgencyID            *string  `json:"agency_id"`
	Country             string   `json:"country"`
	ChannelURL          string   `json:"channel_url"`
	Status              string   `json:"status"`
}

func ExportLeads(db *core.DB, exportDir string) (int, error) {
	channels, err := db.GetExportReadyChannels()
	if err != nil {
		return 0, fmt.Errorf("get export-ready channels: %w", err)
	}

	if len(channels) == 0 {
		slog.Info("no leads to export")
		return 0, nil
	}

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return 0, fmt.Errorf("create export dir: %w", err)
	}

	var leads []ExportLead
	var ids []string

	for _, ch := range channels {
		var signals []string
		if ch.MonetizationSignals != "" {
			json.Unmarshal([]byte(ch.MonetizationSignals), &signals)
		}

		handle := strings.TrimPrefix(ch.Handle, "@")
		channelURL := fmt.Sprintf("https://youtube.com/@%s", handle)

		lead := ExportLead{
			ChannelID:           ch.ID,
			Name:                ch.Name,
			Handle:              ch.Handle,
			Email:               ch.Email,
			SubscriberCount:     ch.SubscriberCount,
			Niche:               ch.Niche,
			MonetizationScore:   ch.MonetizationScore,
			MonetizationSignals: signals,
			Website:             ch.Website,
			Country:             ch.Country,
			ChannelURL:          channelURL,
			Status:              ch.Status,
		}
		leads = append(leads, lead)
		ids = append(ids, ch.ID)
	}

	timestamp := time.Now().UTC().Format("20060102_150405")
	filename := filepath.Join(exportDir, fmt.Sprintf("leads_%s.json", timestamp))

	data, err := json.MarshalIndent(leads, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("marshal leads: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return 0, fmt.Errorf("write export file: %w", err)
	}

	if err := db.MarkChannelsExported(ids); err != nil {
		return 0, fmt.Errorf("mark channels exported: %w", err)
	}

	slog.Info("leads exported", "count", len(leads), "file", filename)
	return len(leads), nil
}
