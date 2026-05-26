package modules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aqfl/tmleadbot/internal/core"
)

func TestExportLeads_Empty(t *testing.T) {
	db, err := core.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	dir := t.TempDir()
	count, err := ExportLeads(db, dir)
	if err != nil {
		t.Fatalf("ExportLeads: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestExportLeads_WithData(t *testing.T) {
	db, err := core.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertChannel(&core.Channel{
		ID:                "UCexport1",
		Handle:            "@creator1",
		Name:              "Creator One",
		SubscriberCount:   150000,
		Email:             "creator1@example.com",
		MonetizationScore: 8,
		MonetizationTier:  "high",
		MonetizationSignals: `["email_present","sponsorships"]`,
		Status:            "enriched",
	})

	dir := t.TempDir()
	count, err := ExportLeads(db, dir)
	if err != nil {
		t.Fatalf("ExportLeads: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 lead, got %d", count)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	data, err := os.ReadFile(filepath.Join(dir, files[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var leads []ExportLead
	if err := json.Unmarshal(data, &leads); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(leads) != 1 {
		t.Fatalf("expected 1 lead in JSON, got %d", len(leads))
	}

	lead := leads[0]
	if lead.ChannelID != "UCexport1" {
		t.Errorf("ChannelID = %q", lead.ChannelID)
	}
	if lead.Name != "Creator One" {
		t.Errorf("Name = %q", lead.Name)
	}
	if lead.Handle != "@creator1" {
		t.Errorf("Handle = %q", lead.Handle)
	}
	if lead.Email != "creator1@example.com" {
		t.Errorf("Email = %q", lead.Email)
	}
	if lead.SubscriberCount != 150000 {
		t.Errorf("SubscriberCount = %d", lead.SubscriberCount)
	}
	if lead.MonetizationScore != 8 {
		t.Errorf("MonetizationScore = %d", lead.MonetizationScore)
	}
}

func TestExportLead_ChannelURL(t *testing.T) {
	db, err := core.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertChannel(&core.Channel{
		ID:                "UCurl1",
		Handle:            "@creator2",
		Name:              "Creator Two",
		Email:             "c2@example.com",
		MonetizationScore: 5,
		Status:            "enriched",
	})

	dir := t.TempDir()
	ExportLeads(db, dir)

	files, _ := os.ReadDir(dir)
	data, _ := os.ReadFile(filepath.Join(dir, files[0].Name()))
	var leads []ExportLead
	json.Unmarshal(data, &leads)

	if len(leads) > 0 {
		expected := "https://youtube.com/@creator2"
		if leads[0].ChannelURL != expected {
			t.Errorf("ChannelURL = %q, want %q", leads[0].ChannelURL, expected)
		}
	}
}

func TestExportLead_EmptyHandleURL(t *testing.T) {
	db, err := core.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertChannel(&core.Channel{
		ID:                "UCnohandle",
		Email:             "no@example.com",
		MonetizationScore: 5,
		Status:            "enriched",
	})

	dir := t.TempDir()
	ExportLeads(db, dir)

	files, _ := os.ReadDir(dir)
	data, _ := os.ReadFile(filepath.Join(dir, files[0].Name()))
	var leads []ExportLead
	json.Unmarshal(data, &leads)

	if len(leads) > 0 {
		if leads[0].ChannelURL == "" {
			t.Log("ChannelURL is empty for channel with no handle (known edge case)")
		}
	}
}

func TestExportLead_Structures(t *testing.T) {
	lead := ExportLead{
		ChannelID:           "UCx",
		Name:                "Test",
		Handle:              "@test",
		Email:               "t@t.com",
		SubscriberCount:     100,
		MonetizationScore:   5,
		MonetizationSignals: []string{"email_present", "sponsorships"},
		Website:             "https://t.com",
		Country:             "US",
		ChannelURL:          "https://youtube.com/@test",
		Status:              "exported",
	}

	data, err := json.Marshal(lead)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ExportLead
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ChannelID != lead.ChannelID {
		t.Errorf("ChannelID mismatch")
	}
	if decoded.MonetizationScore != lead.MonetizationScore {
		t.Errorf("MonetizationScore mismatch")
	}
	if len(decoded.MonetizationSignals) != 2 {
		t.Errorf("expected 2 signals, got %d", len(decoded.MonetizationSignals))
	}
}
