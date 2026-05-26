package core

import (
	"testing"
)

func TestNewDB_InMemory(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM channels").Scan(&count)
	if err != nil {
		t.Fatalf("channels table not created: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM jobs").Scan(&count)
	if err != nil {
		t.Fatalf("jobs table not created: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM agencies").Scan(&count)
	if err != nil {
		t.Fatalf("agencies table not created: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM runs").Scan(&count)
	if err != nil {
		t.Fatalf("runs table not created: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM rate_limit_budget").Scan(&count)
	if err != nil {
		t.Fatalf("rate_limit_budget table not created: %v", err)
	}
}

func TestUpsertAndGetChannel(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	ch := &Channel{
		ID:                "UCtest123",
		Handle:            "@testcreator",
		Name:              "Test Creator",
		SubscriberCount:   100000,
		ViewCount:         5000000,
		VideoCount:        200,
		UploadFrequency:   4.5,
		Niche:             "tech",
		Description:       "A test channel",
		Country:           "US",
		Email:             "test@example.com",
		Website:           "https://example.com",
		MonetizationScore: 7,
		MonetizationTier:  "high",
		Status:            "enriched",
		DiscoveryKeyword:  "tech sponsor",
	}

	if err := db.UpsertChannel(ch); err != nil {
		t.Fatalf("UpsertChannel: %v", err)
	}

	got, err := db.GetChannel("UCtest123")
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}

	if got.ID != ch.ID {
		t.Errorf("ID = %q, want %q", got.ID, ch.ID)
	}
	if got.Name != ch.Name {
		t.Errorf("Name = %q, want %q", got.Name, ch.Name)
	}
	if got.SubscriberCount != ch.SubscriberCount {
		t.Errorf("SubscriberCount = %d, want %d", got.SubscriberCount, ch.SubscriberCount)
	}
	if got.MonetizationScore != ch.MonetizationScore {
		t.Errorf("MonetizationScore = %d, want %d", got.MonetizationScore, ch.MonetizationScore)
	}
	if got.Email != ch.Email {
		t.Errorf("Email = %q, want %q", got.Email, ch.Email)
	}
	if got.Status != ch.Status {
		t.Errorf("Status = %q, want %q", got.Status, ch.Status)
	}
}

func TestChannelExists(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	exists, err := db.ChannelExists("nonexistent")
	if err != nil {
		t.Fatalf("ChannelExists: %v", err)
	}
	if exists {
		t.Error("expected false for nonexistent channel")
	}

	db.UpsertChannel(&Channel{ID: "UCexists", Status: "new"})

	exists, err = db.ChannelExists("UCexists")
	if err != nil {
		t.Fatalf("ChannelExists: %v", err)
	}
	if !exists {
		t.Error("expected true for existing channel")
	}
}

func TestUpsertChannel_UpdatesExisting(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	ch := &Channel{ID: "UCtest", Name: "Old Name", Email: "old@example.com", SubscriberCount: 50000, Status: "enriched"}
	db.UpsertChannel(ch)

	ch2 := &Channel{ID: "UCtest", Name: "New Name", Email: "new@example.com", SubscriberCount: 60000, Status: "enriched"}
	db.UpsertChannel(ch2)

	got, _ := db.GetChannel("UCtest")
	if got.Name != "New Name" {
		t.Errorf("Name = %q, want %q", got.Name, "New Name")
	}
	if got.SubscriberCount != 60000 {
		t.Errorf("SubscriberCount = %d, want %d", got.SubscriberCount, 60000)
	}
}

func TestGetStaleChannelIDs(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	ids, err := db.GetStaleChannelIDs(30)
	if err != nil {
		t.Fatalf("GetStaleChannelIDs: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected 0 stale channels, got %d", len(ids))
	}
}

func TestInsertAndDequeueJob(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	id, err := db.InsertJob(&Job{Type: "enrich_channel", Payload: "UCtest"})
	if err != nil {
		t.Fatalf("InsertJob: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero job ID")
	}

	job, err := db.DequeueJob("enrich_channel")
	if err != nil {
		t.Fatalf("DequeueJob: %v", err)
	}
	if job == nil {
		t.Fatal("expected job, got nil")
	}
	if job.Type != "enrich_channel" {
		t.Errorf("job.Type = %q, want %q", job.Type, "enrich_channel")
	}
	if job.Payload != "UCtest" {
		t.Errorf("job.Payload = %q, want %q", job.Payload, "UCtest")
	}
	if job.Status != "running" {
		t.Errorf("job.Status = %q, want %q", job.Status, "running")
	}
}

func TestDequeueJob_Empty(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	job, err := db.DequeueJob("enrich_channel")
	if err != nil {
		t.Fatalf("DequeueJob: %v", err)
	}
	if job != nil {
		t.Fatal("expected nil for empty queue")
	}
}

func TestMarkJobDone(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.InsertJob(&Job{Type: "enrich_channel", Payload: "UCtest"})
	job, _ := db.DequeueJob("enrich_channel")

	if err := db.MarkJobDone(job.ID); err != nil {
		t.Fatalf("MarkJobDone: %v", err)
	}

	// should get nil for next dequeue
	next, _ := db.DequeueJob("enrich_channel")
	if next != nil {
		t.Error("expected nil, job was not marked done")
	}
}

func TestMarkJobFailed(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.InsertJob(&Job{Type: "test", Payload: "x"})
	job, _ := db.DequeueJob("test")

	if err := db.MarkJobFailed(job.ID, "something went wrong"); err != nil {
		t.Fatalf("MarkJobFailed: %v", err)
	}

	next, _ := db.DequeueJob("test")
	if next != nil {
		t.Error("expected nil for failed job")
	}
}

func TestRequeueJob(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.InsertJob(&Job{Type: "test", Payload: "x"})
	job, _ := db.DequeueJob("test")

	if err := db.RequeueJob(job.ID, 1); err != nil {
		t.Fatalf("RequeueJob: %v", err)
	}

	requeued, _ := db.DequeueJob("test")
	if requeued == nil {
		t.Fatal("expected requeued job")
	}
	if requeued.Retries != 1 {
		t.Errorf("Retries = %d, want 1", requeued.Retries)
	}
}

func TestResetStaleJobs(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.InsertJob(&Job{Type: "test", Payload: "x"})
	db.DequeueJob("test") // marks it running
	db.ResetStaleJobs()

	job, _ := db.DequeueJob("test")
	if job == nil {
		t.Error("expected stale job to be reset to pending")
	}
}

func TestInsertAndCompleteRun(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	id, err := db.InsertRun(&Run{RunType: "discover", StartedAt: "2025-01-01T00:00:00Z"})
	if err != nil {
		t.Fatalf("InsertRun: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero run ID")
	}

	err = db.CompleteRun(id, &Run{
		CompletedAt:        "2025-01-01T01:00:00Z",
		ChannelsDiscovered: 10,
		ChannelsEnriched:   5,
		EmailsFound:        3,
		Errors:             1,
	})
	if err != nil {
		t.Fatalf("CompleteRun: %v", err)
	}
}

func TestUpsertAgency(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	a := &Agency{ID: "testagency", Name: "Test Agency", Website: "https://testagency.com", DetectionMethod: "seed"}
	if err := db.UpsertAgency(a); err != nil {
		t.Fatalf("UpsertAgency: %v", err)
	}
}

func TestUpsertAgencyChannel(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertAgency(&Agency{ID: "agency1", Name: "Agency 1"})
	db.UpsertChannel(&Channel{ID: "UCch1", Status: "enriched"})

	ac := &AgencyChannel{AgencyID: "agency1", ChannelID: "UCch1", Confidence: 0.8, Evidence: "test"}
	if err := db.UpsertAgencyChannel(ac); err != nil {
		t.Fatalf("UpsertAgencyChannel: %v", err)
	}
}

func TestSetChannelAgency(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertAgency(&Agency{ID: "agency1", Name: "Agency 1"})
	db.UpsertChannel(&Channel{ID: "UCch1", Status: "enriched"})

	if err := db.SetChannelAgency("UCch1", "agency1"); err != nil {
		t.Fatalf("SetChannelAgency: %v", err)
	}

	ch, _ := db.GetChannel("UCch1")
	if ch.AgencyID == nil || *ch.AgencyID != "agency1" {
		t.Errorf("AgencyID = %v, want %q", ch.AgencyID, "agency1")
	}
}

func TestConsumeBudget(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	if err := db.ConsumeBudget("test_budget", 100, 10000); err != nil {
		t.Fatalf("ConsumeBudget: %v", err)
	}

	rem, err := db.BudgetRemaining("test_budget")
	if err != nil {
		t.Fatalf("BudgetRemaining: %v", err)
	}
	if rem < 0 {
		t.Fatal("expected non-negative remaining budget")
	}
}

func TestBudgetRemaining_NoRow(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	rem, err := db.BudgetRemaining("untracked")
	if err != nil {
		t.Fatalf("BudgetRemaining: %v", err)
	}
	if rem != -1 {
		t.Errorf("expected -1 for untracked budget, got %d", rem)
	}
}

func TestCountByStatus(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertChannel(&Channel{ID: "UC1", Status: "enriched"})
	db.UpsertChannel(&Channel{ID: "UC2", Status: "enriched"})
	db.UpsertChannel(&Channel{ID: "UC3", Status: "new"})

	n, err := db.CountByStatus("enriched")
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 enriched, got %d", n)
	}
}

func TestGetExportReadyChannels(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertChannel(&Channel{ID: "UC1", Status: "enriched", MonetizationScore: 5, Email: "a@b.com"})
	db.UpsertChannel(&Channel{ID: "UC2", Status: "enriched", MonetizationScore: 2, Email: "a@b.com"})
	db.UpsertChannel(&Channel{ID: "UC3", Status: "enriched", MonetizationScore: 5, Email: ""})
	db.UpsertChannel(&Channel{ID: "UC4", Status: "new", MonetizationScore: 5, Email: "a@b.com"})

	channels, err := db.GetExportReadyChannels()
	if err != nil {
		t.Fatalf("GetExportReadyChannels: %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("expected 1 export-ready channel, got %d", len(channels))
	}
	if len(channels) > 0 && channels[0].ID != "UC1" {
		t.Errorf("expected UC1, got %s", channels[0].ID)
	}
}

func TestMarkChannelsExported(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertChannel(&Channel{ID: "UC1", Status: "enriched"})
	db.UpsertChannel(&Channel{ID: "UC2", Status: "enriched"})

	if err := db.MarkChannelsExported([]string{"UC1", "UC2"}); err != nil {
		t.Fatalf("MarkChannelsExported: %v", err)
	}

	n, _ := db.CountByStatus("exported")
	if n != 2 {
		t.Errorf("expected 2 exported, got %d", n)
	}
}

func TestVExportReadyView(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()

	db.UpsertChannel(&Channel{ID: "UC1", Status: "enriched", MonetizationScore: 7, Email: "c@d.com"})

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM v_export_ready").Scan(&count)
	if err != nil {
		t.Fatalf("v_export_ready error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 in v_export_ready, got %d", count)
	}
}
