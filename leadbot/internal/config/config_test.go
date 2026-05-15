package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Unset any env vars that might interfere
	for _, k := range []string{"YT_API_KEY", "MAX_ENRICHMENTS", "MIN_SUBS", "MAX_SUBS", "MIN_SCORE", "DATA_DIR", "DB_PATH", "KEYWORDS_PATH", "AGENCIES_PATH"} {
		os.Unsetenv(k)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.MaxEnrichments != 800 {
		t.Errorf("MaxEnrichments = %d, want 800", cfg.MaxEnrichments)
	}
	if cfg.MinSubs != 40000 {
		t.Errorf("MinSubs = %d, want 40000", cfg.MinSubs)
	}
	if cfg.MaxSubs != 500000 {
		t.Errorf("MaxSubs = %d, want 500000", cfg.MaxSubs)
	}
	if cfg.MinScore != 3 {
		t.Errorf("MinScore = %d, want 3", cfg.MinScore)
	}
	if cfg.DataDir != "data" {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, "data")
	}
	if cfg.DBPath != "data/ytleadbot.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "data/ytleadbot.db")
	}
	if cfg.RequestDelay != 5 {
		t.Errorf("RequestDelay = %d, want 5", cfg.RequestDelay)
	}
	if cfg.YTApiKey != "" {
		t.Errorf("YTApiKey = %q, want empty", cfg.YTApiKey)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("YT_API_KEY", "test-key-123")
	os.Setenv("MAX_ENRICHMENTS", "500")
	os.Setenv("MIN_SUBS", "10000")
	os.Setenv("MAX_SUBS", "999999")
	os.Setenv("MIN_SCORE", "5")
	os.Setenv("REQUEST_DELAY", "10")
	defer func() {
		os.Unsetenv("YT_API_KEY")
		os.Unsetenv("MAX_ENRICHMENTS")
		os.Unsetenv("MIN_SUBS")
		os.Unsetenv("MAX_SUBS")
		os.Unsetenv("MIN_SCORE")
		os.Unsetenv("REQUEST_DELAY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.YTApiKey != "test-key-123" {
		t.Errorf("YTApiKey = %q", cfg.YTApiKey)
	}
	if cfg.MaxEnrichments != 500 {
		t.Errorf("MaxEnrichments = %d", cfg.MaxEnrichments)
	}
	if cfg.MinSubs != 10000 {
		t.Errorf("MinSubs = %d", cfg.MinSubs)
	}
	if cfg.MaxSubs != 999999 {
		t.Errorf("MaxSubs = %d", cfg.MaxSubs)
	}
	if cfg.MinScore != 5 {
		t.Errorf("MinScore = %d", cfg.MinScore)
	}
	if cfg.RequestDelay != 10 {
		t.Errorf("RequestDelay = %d", cfg.RequestDelay)
	}
}

func TestLoad_KeywordsAndAgencies(t *testing.T) {
	os.Unsetenv("KEYWORDS_PATH")
	os.Unsetenv("AGENCIES_PATH")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(cfg.Keywords) == 0 {
		t.Error("expected non-empty keywords")
	}
	if len(cfg.NicheCategories) == 0 {
		t.Error("expected non-empty niche categories")
	}
	if len(cfg.AgencySeeds) == 0 {
		t.Error("expected non-empty agency seeds")
	}
}

func TestDefaultNicheCategories(t *testing.T) {
	cats := defaultNicheCategories()
	expected := []string{"fitness", "personal_finance", "tech", "business", "cooking", "travel", "gaming", "self_improvement", "parenting", "beauty"}
	for _, n := range expected {
		if _, ok := cats[n]; !ok {
			t.Errorf("missing niche category: %s", n)
		}
	}
}

func TestDefaultAgencies(t *testing.T) {
	agencies := defaultAgencies()
	if len(agencies) < 10 {
		t.Errorf("expected >= 10 default agencies, got %d", len(agencies))
	}
	found := false
	for _, a := range agencies {
		if a.Name == "Whalar" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Whalar in default agencies")
	}
}
