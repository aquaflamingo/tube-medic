package core

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Channel struct {
	ID                  string  `json:"id"`
	Handle              string  `json:"handle"`
	Name                string  `json:"name"`
	SubscriberCount     int64   `json:"subscriber_count"`
	ViewCount           int64   `json:"view_count"`
	VideoCount          int64   `json:"video_count"`
	UploadFrequency     float64 `json:"upload_frequency"`
	Niche               string  `json:"niche"`
	Description         string  `json:"description"`
	Country             string  `json:"country"`
	Email               string  `json:"email"`
	Website             string  `json:"website"`
	SocialLinks         string  `json:"social_links"`
	MonetizationScore   int     `json:"monetization_score"`
	MonetizationTier    string  `json:"monetization_tier"`
	MonetizationSignals string  `json:"monetization_signals"`
	AgencyID            *string `json:"agency_id"`
	Status              string  `json:"status"`
	DiscoveryKeyword    string  `json:"discovery_keyword"`
	FirstSeenAt         string  `json:"first_seen_at"`
	EnrichedAt          *string `json:"enriched_at"`
	ExportedAt          *string `json:"exported_at"`
	RawYTMetadata       string  `json:"raw_yt_metadata"`
}

type Job struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Payload   string `json:"payload"`
	Status    string `json:"status"`
	Retries   int    `json:"retries"`
	Error     string `json:"error"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Agency struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Website         string `json:"website"`
	Email           string `json:"email"`
	RosterCount     int    `json:"roster_count"`
	DetectionMethod string `json:"detection_method"`
	CreatedAt       string `json:"created_at"`
}

type AgencyChannel struct {
	AgencyID   string  `json:"agency_id"`
	ChannelID  string  `json:"channel_id"`
	Confidence float64 `json:"confidence"`
	Evidence   string  `json:"evidence"`
}

type Run struct {
	ID                 int64  `json:"id"`
	RunType            string `json:"run_type"`
	StartedAt          string `json:"started_at"`
	CompletedAt        string `json:"completed_at"`
	ChannelsDiscovered int    `json:"channels_discovered"`
	ChannelsEnriched   int    `json:"channels_enriched"`
	EmailsFound        int    `json:"emails_found"`
	AgenciesFound      int    `json:"agencies_found"`
	LeadsExported      int    `json:"leads_exported"`
	Errors             int    `json:"errors"`
}

type DB struct {
	*sql.DB
}

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	d := &DB{db}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

func (db *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS channels (
		id                    TEXT PRIMARY KEY,
		handle                TEXT,
		name                  TEXT,
		subscriber_count      INTEGER,
		view_count            INTEGER,
		video_count           INTEGER,
		upload_frequency      REAL,
		niche                 TEXT,
		description           TEXT,
		country               TEXT,
		email                 TEXT,
		website               TEXT,
		social_links          TEXT,
		monetization_score    INTEGER DEFAULT 0,
		monetization_tier     TEXT,
		monetization_signals  TEXT,
		agency_id             TEXT REFERENCES agencies(id),
		status                TEXT DEFAULT 'new',
		discovery_keyword     TEXT,
		first_seen_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		enriched_at           TIMESTAMP,
		exported_at           TIMESTAMP,
		raw_yt_metadata       TEXT
	);

	CREATE TABLE IF NOT EXISTS agencies (
		id              TEXT PRIMARY KEY,
		name            TEXT,
		website         TEXT,
		email           TEXT,
		roster_count    INTEGER DEFAULT 0,
		detection_method TEXT,
		created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS agency_channels (
		agency_id   TEXT REFERENCES agencies(id),
		channel_id  TEXT REFERENCES channels(id),
		confidence  REAL DEFAULT 1.0,
		evidence    TEXT,
		PRIMARY KEY (agency_id, channel_id)
	);

	CREATE TABLE IF NOT EXISTS jobs (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		type        TEXT NOT NULL,
		payload     TEXT,
		status      TEXT DEFAULT 'pending',
		retries     INTEGER DEFAULT 0,
		error       TEXT,
		created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at  TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS runs (
		id                  INTEGER PRIMARY KEY AUTOINCREMENT,
		run_type            TEXT,
		started_at          TIMESTAMP,
		completed_at        TIMESTAMP,
		channels_discovered INTEGER DEFAULT 0,
		channels_enriched   INTEGER DEFAULT 0,
		emails_found        INTEGER DEFAULT 0,
		agencies_found      INTEGER DEFAULT 0,
		leads_exported      INTEGER DEFAULT 0,
		errors              INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS rate_limit_budget (
		id    TEXT NOT NULL,
		date  TEXT NOT NULL,
		used  INTEGER DEFAULT 0,
		max_units INTEGER DEFAULT 0,
		PRIMARY KEY (id, date)
	);

	CREATE VIEW IF NOT EXISTS v_export_ready AS
		SELECT * FROM channels
		WHERE status = 'enriched'
		  AND monetization_score >= 3
		  AND email IS NOT NULL AND email != ''
		ORDER BY monetization_score DESC;

	CREATE VIEW IF NOT EXISTS v_agency_rosters AS
		SELECT a.name AS agency_name, a.website, c.name AS creator_name,
			   c.subscriber_count, c.email, c.monetization_score
		FROM agencies a
		JOIN agency_channels ac ON ac.agency_id = a.id
		JOIN channels c ON c.id = ac.channel_id
		ORDER BY a.name, c.subscriber_count DESC;
	`
	_, err := db.Exec(schema)
	return err
}

func (db *DB) UpsertChannel(ch *Channel) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`
		INSERT INTO channels (id, handle, name, subscriber_count, view_count, video_count,
			upload_frequency, niche, description, country, email, website, social_links,
			monetization_score, monetization_tier, monetization_signals, status,
			discovery_keyword, enriched_at, raw_yt_metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			handle=excluded.handle, name=excluded.name,
			subscriber_count=excluded.subscriber_count, view_count=excluded.view_count,
			video_count=excluded.video_count, upload_frequency=excluded.upload_frequency,
			niche=excluded.niche, description=excluded.description,
			country=excluded.country,
			email=CASE WHEN excluded.email != '' THEN excluded.email ELSE channels.email END,
			website=CASE WHEN excluded.website != '' THEN excluded.website ELSE channels.website END,
			social_links=excluded.social_links,
			monetization_score=excluded.monetization_score,
			monetization_tier=excluded.monetization_tier,
			monetization_signals=excluded.monetization_signals,
			status=CASE WHEN channels.status = 'new' THEN excluded.status ELSE channels.status END,
			enriched_at=excluded.enriched_at,
			raw_yt_metadata=excluded.raw_yt_metadata
	`, ch.ID, ch.Handle, ch.Name, ch.SubscriberCount, ch.ViewCount, ch.VideoCount,
		ch.UploadFrequency, ch.Niche, ch.Description, ch.Country, ch.Email, ch.Website,
		ch.SocialLinks, ch.MonetizationScore, ch.MonetizationTier, ch.MonetizationSignals,
		ch.Status, ch.DiscoveryKeyword, now, ch.RawYTMetadata)
	return err
}

func (db *DB) GetChannel(id string) (*Channel, error) {
	row := db.QueryRow(`SELECT id, handle, name, subscriber_count, view_count, video_count,
		upload_frequency, niche, description, country, email, website, social_links,
		monetization_score, monetization_tier, monetization_signals, agency_id, status,
		discovery_keyword, first_seen_at, enriched_at, exported_at, raw_yt_metadata
		FROM channels WHERE id = ?`, id)
	ch := &Channel{}
	var handle, name, niche, desc, country, email, website, socialLinks sql.NullString
	var monTier, status, discKeyword, firstSeen, rawYT sql.NullString
	var enrichedAt, exportedAt, agencyID sql.NullString
	var monScore int
	err := row.Scan(&ch.ID, &handle, &name, &ch.SubscriberCount, &ch.ViewCount, &ch.VideoCount,
		&ch.UploadFrequency, &niche, &desc, &country, &email, &website, &socialLinks,
		&monScore, &monTier, &ch.MonetizationSignals, &agencyID, &status,
		&discKeyword, &firstSeen, &enrichedAt, &exportedAt, &rawYT)
	if err != nil {
		return nil, err
	}
	ch.Handle = handle.String
	ch.Name = name.String
	ch.Niche = niche.String
	ch.Description = desc.String
	ch.Country = country.String
	ch.Email = email.String
	ch.Website = website.String
	ch.SocialLinks = socialLinks.String
	ch.MonetizationScore = monScore
	ch.MonetizationTier = monTier.String
	if agencyID.Valid {
		ch.AgencyID = &agencyID.String
	}
	ch.Status = status.String
	ch.DiscoveryKeyword = discKeyword.String
	ch.FirstSeenAt = firstSeen.String
	if enrichedAt.Valid {
		ch.EnrichedAt = &enrichedAt.String
	}
	if exportedAt.Valid {
		ch.ExportedAt = &exportedAt.String
	}
	ch.RawYTMetadata = rawYT.String
	return ch, nil
}

func (db *DB) GetExportReadyChannels() ([]*Channel, error) {
	rows, err := db.Query(`SELECT id, handle, name, subscriber_count, view_count, video_count,
		upload_frequency, niche, description, country, email, website, social_links,
		monetization_score, monetization_tier, monetization_signals, status,
		discovery_keyword, first_seen_at, enriched_at, raw_yt_metadata
		FROM channels WHERE status = 'enriched' AND monetization_score >= 3 AND email IS NOT NULL AND email != ''
		ORDER BY monetization_score DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		ch := &Channel{}
		var handle, name, niche, desc, country, email, website, socialLinks sql.NullString
		var monTier, status, discKeyword, firstSeen, rawYT sql.NullString
		var enrichedAt sql.NullString
		var monScore int
		err := rows.Scan(&ch.ID, &handle, &name, &ch.SubscriberCount, &ch.ViewCount, &ch.VideoCount,
			&ch.UploadFrequency, &niche, &desc, &country, &email, &website, &socialLinks,
			&monScore, &monTier, &ch.MonetizationSignals, &status,
			&discKeyword, &firstSeen, &enrichedAt, &rawYT)
		if err != nil {
			return nil, err
		}
		ch.Handle = handle.String
		ch.Name = name.String
		ch.Niche = niche.String
		ch.Description = desc.String
		ch.Country = country.String
		ch.Email = email.String
		ch.Website = website.String
		ch.SocialLinks = socialLinks.String
		ch.MonetizationScore = monScore
		ch.MonetizationTier = monTier.String
		ch.Status = status.String
		ch.DiscoveryKeyword = discKeyword.String
		ch.FirstSeenAt = firstSeen.String
		if enrichedAt.Valid {
			ch.EnrichedAt = &enrichedAt.String
		}
		ch.RawYTMetadata = rawYT.String
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func (db *DB) MarkChannelsExported(ids []string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		_, err := tx.Exec(`UPDATE channels SET status = 'exported', exported_at = ? WHERE id = ?`, now, id)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) InsertJob(j *Job) (int64, error) {
	res, err := db.Exec(`INSERT INTO jobs (type, payload, status) VALUES (?, ?, 'pending')`,
		j.Type, j.Payload)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) DequeueJob(jobType string) (*Job, error) {
	var j Job
	var errVal sql.NullString
	err := db.QueryRow(`UPDATE jobs SET status = 'running', updated_at = CURRENT_TIMESTAMP
		WHERE id = (SELECT id FROM jobs WHERE type = ? AND status = 'pending' ORDER BY created_at LIMIT 1)
		RETURNING id, type, payload, status, retries, error, created_at, updated_at`,
		jobType).Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Retries, &errVal, &j.CreatedAt, &j.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	j.Error = errVal.String
	return &j, nil
}

func (db *DB) MarkJobDone(id int64) error {
	_, err := db.Exec(`UPDATE jobs SET status = 'done', updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

func (db *DB) MarkJobFailed(id int64, errMsg string) error {
	_, err := db.Exec(`UPDATE jobs SET status = 'failed', error = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, errMsg, id)
	return err
}

func (db *DB) ResetStaleJobs() error {
	_, err := db.Exec(`UPDATE jobs SET status = 'pending', updated_at = CURRENT_TIMESTAMP WHERE status = 'running'`)
	return err
}

func (db *DB) InsertRun(r *Run) (int64, error) {
	res, err := db.Exec(`INSERT INTO runs (run_type, started_at) VALUES (?, ?)`, r.RunType, r.StartedAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) CompleteRun(id int64, r *Run) error {
	_, err := db.Exec(`UPDATE runs SET completed_at = ?,
		channels_discovered = ?, channels_enriched = ?, emails_found = ?,
		agencies_found = ?, leads_exported = ?, errors = ?
		WHERE id = ?`, r.CompletedAt, r.ChannelsDiscovered, r.ChannelsEnriched,
		r.EmailsFound, r.AgenciesFound, r.LeadsExported, r.Errors, id)
	return err
}

func (db *DB) UpsertAgency(a *Agency) error {
	_, err := db.Exec(`INSERT INTO agencies (id, name, website, email, detection_method)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, website=excluded.website,
			email=CASE WHEN excluded.email != '' THEN excluded.email ELSE agencies.email END,
			detection_method=excluded.detection_method`,
		a.ID, a.Name, a.Website, a.Email, a.DetectionMethod)
	return err
}

func (db *DB) UpsertAgencyChannel(ac *AgencyChannel) error {
	_, err := db.Exec(`INSERT INTO agency_channels (agency_id, channel_id, confidence, evidence)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(agency_id, channel_id) DO UPDATE SET
			confidence=excluded.confidence, evidence=excluded.evidence`,
		ac.AgencyID, ac.ChannelID, ac.Confidence, ac.Evidence)
	return err
}

func (db *DB) SetChannelAgency(channelID, agencyID string) error {
	_, err := db.Exec(`UPDATE channels SET agency_id = ? WHERE id = ?`, agencyID, channelID)
	return err
}

func (db *DB) ChannelExists(id string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM channels WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *DB) GetStaleChannelIDs(maxAgeDays int) ([]string, error) {
	rows, err := db.Query(`SELECT id FROM channels
		WHERE enriched_at IS NOT NULL
		AND julianday('now') - julianday(enriched_at) > ?
		ORDER BY enriched_at`, maxAgeDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (db *DB) RequeueJob(id int64, retries int) error {
	_, err := db.Exec(`UPDATE jobs SET status = 'pending', retries = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, retries, id)
	return err
}

func (db *DB) ConsumeBudget(id string, amount, maxUnits int) error {
	today := time.Now().UTC().Format("2006-01-02")
	_, err := db.Exec(`INSERT INTO rate_limit_budget (id, date, used, max_units) VALUES (?, ?, ?, ?)
		ON CONFLICT(id, date) DO UPDATE SET used = used + ?`,
		id, today, amount, maxUnits, amount)
	return err
}

func (db *DB) BudgetRemaining(id string) (int, error) {
	today := time.Now().UTC().Format("2006-01-02")
	var used, maxUnits int
	err := db.QueryRow(`SELECT used, max_units FROM rate_limit_budget WHERE id = ? AND date = ?`, id, today).Scan(&used, &maxUnits)
	if err == sql.ErrNoRows {
		return -1, nil
	}
	if err != nil {
		return 0, err
	}
	rem := maxUnits - used
	if rem < 0 {
		return 0, nil
	}
	return rem, nil
}

func (db *DB) CountByStatus(status string) (int, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM channels WHERE status = ?`, status).Scan(&n)
	return n, err
}
