package modules

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/aquaflamingo/ytleadbot/internal/config"
	"github.com/aquaflamingo/ytleadbot/internal/core"
	"github.com/aquaflamingo/ytleadbot/internal/utils"
)

var rosterPaths = []string{
	"/talent", "/roster", "/creators", "/our-creators",
	"/talent-roster", "/our-talent", "/artists", "/influencers",
}

var agencySearchTerms = []string{
	"creator talent agency",
	"YouTube MCN management",
	"influencer management company",
	"talent management agency",
	"digital talent agency",
	"creator management",
}

var agencyDescPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)managed\s+by\s+([A-Z][A-Za-z0-9.\s&]+)`),
	regexp.MustCompile(`(?i)represented\s+by\s+([A-Z][A-Za-z0-9.\s&]+)`),
	regexp.MustCompile(`(?i)booking\s+via\s+([A-Z][A-Za-z0-9.\s&]+)`),
	regexp.MustCompile(`(?i)(?:talent|agency)[:\s]+([A-Z][A-Za-z0-9.\s&]+)`),
	regexp.MustCompile(`(?i)part\s+of\s+the\s+([A-Z][A-Za-z0-9.\s&]+?)\s+(?:family|network|roster)`),
}

type agencyChannelLink struct {
	AgencyID  string
	ChannelID string
	Evidence  string
}

func RunAgencyScan(db *core.DB, cfg *config.Config) (int, int, error) {
	slog.Info("starting agency scan")

	var links []agencyChannelLink

	foundA, err := scanAgencyWebsites(db, cfg, &links)
	if err != nil {
		slog.Warn("agency website scan failed", "error", err)
	}
	slog.Info("strategy A complete", "channels_queued", foundA)

	foundB, err := searchAgencyChannels(db, cfg)
	if err != nil {
		slog.Warn("agency search failed", "error", err)
	}
	slog.Info("strategy B complete", "channels_queued", foundB)

	processed, failed, err := ProcessJobQueue(db, cfg, "enrich_channel")
	if err != nil {
		slog.Warn("enrich queue processing failed", "error", err)
	}
	slog.Info("enrichment complete", "processed", processed, "failed", failed)

	linked, err := linkAgencyChannels(db, links)
	if err != nil {
		slog.Warn("agency linking failed", "error", err)
	}
	slog.Info("agency links created", "links", linked)

	foundC, err := clusterByEmailDomain(db, cfg)
	if err != nil {
		slog.Warn("email domain clustering failed", "error", err)
	}
	slog.Info("strategy C complete", "agencies_found", foundC)

	foundD, err := scanDescriptionsForAgencies(db)
	if err != nil {
		slog.Warn("description scan failed", "error", err)
	}
	slog.Info("strategy D complete", "agencies_found", foundD)

	totalAgencies := foundC + foundD
	totalChannels := foundA + foundB

	runID := startRun(db, "agency_scan")
	if runID != 0 {
		finishRun(db, runID, totalChannels, processed, totalAgencies, 0, failed)
	}

	return totalAgencies, totalChannels, nil
}

func scanAgencyWebsites(db *core.DB, cfg *config.Config, links *[]agencyChannelLink) (int, error) {
	var total int
	for _, seed := range cfg.AgencySeeds {
		agencyID := extractDomain(seed.Website)
		if agencyID == "" {
			continue
		}

		a := &core.Agency{
			ID:              agencyID,
			Name:            seed.Name,
			Website:         seed.Website,
			DetectionMethod: "seed",
		}
		if err := db.UpsertAgency(a); err != nil {
			slog.Warn("failed to upsert agency", "agency", seed.Name, "error", err)
			continue
		}

		var found bool
		for _, path := range rosterPaths {
			rosterURL := strings.TrimRight(seed.Website, "/") + path
			ids, err := utils.ExtractYouTubeIDsFromPage(rosterURL)
			if err != nil {
				continue
			}
			if len(ids) == 0 {
				continue
			}

			slog.Info("found roster page", "agency", seed.Name, "url", rosterURL, "channels", len(ids))
			for _, ident := range ids {
				exists, err := db.ChannelExists(ident)
				if err != nil {
					continue
				}
				if !exists {
					job := &core.Job{
						Type:    "enrich_channel",
						Payload: ident,
					}
					if _, err := db.InsertJob(job); err != nil {
						slog.Warn("failed to queue enrich job", "id", ident, "error", err)
						continue
					}
					total++
				}
				*links = append(*links, agencyChannelLink{
					AgencyID:  agencyID,
					ChannelID: ident,
					Evidence:  "agency_roster",
				})
			}
			found = true
			break
		}
		if !found {
			slog.Debug("no roster page found", "agency", seed.Name, "website", seed.Website)
		}
	}
	return total, nil
}

func searchAgencyChannels(db *core.DB, cfg *config.Config) (int, error) {
	var total int
	for _, term := range agencySearchTerms {
		slog.Info("searching for agency channels", "term", term)
		ids, err := utils.SearchKeywords(term)
		if err != nil {
			slog.Warn("agency search failed", "term", term, "error", err)
			continue
		}

		seen := make(map[string]bool)
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			exists, err := db.ChannelExists(id)
			if err != nil || exists {
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
			total++
		}
	}
	return total, nil
}

func linkAgencyChannels(db *core.DB, links []agencyChannelLink) (int, error) {
	var count int
	for _, l := range links {
		exists, err := db.ChannelExists(l.ChannelID)
		if err != nil {
			continue
		}
		if !exists {
			continue
		}
		ac := &core.AgencyChannel{
			AgencyID:   l.AgencyID,
			ChannelID:  l.ChannelID,
			Confidence: 1.0,
			Evidence:   l.Evidence,
		}
		if err := db.UpsertAgencyChannel(ac); err != nil {
			slog.Warn("failed to link channel to agency", "agency", l.AgencyID, "channel", l.ChannelID, "error", err)
			continue
		}
		if err := db.SetChannelAgency(l.ChannelID, l.AgencyID); err != nil {
			slog.Warn("failed to set channel agency", "channel", l.ChannelID, "error", err)
		}
		count++
	}
	return count, nil
}

func clusterByEmailDomain(db *core.DB, cfg *config.Config) (int, error) {
	rows, err := db.Query(`SELECT id, email, name FROM channels
		WHERE email IS NOT NULL AND email != '' AND status = 'enriched'`)
	if err != nil {
		return 0, fmt.Errorf("query channels with email: %w", err)
	}
	defer rows.Close()

	type channelInfo struct {
		ID    string
		Name  string
		Email string
	}

	domainGroups := make(map[string][]channelInfo)
	for rows.Next() {
		var id, email, name string
		if err := rows.Scan(&id, &email, &name); err != nil {
			continue
		}
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			continue
		}
		domain := strings.ToLower(parts[1])
		domainGroups[domain] = append(domainGroups[domain], channelInfo{ID: id, Name: name, Email: email})
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	var agencies int
	for domain, channels := range domainGroups {
		if len(channels) < 3 {
			continue
		}

		agencyID := "email:" + domain
		agencyName := domain
		if len(channels) > 0 && channels[0].Name != "" {
			agencyName = channels[0].Name + " & co"
		}

		a := &core.Agency{
			ID:              agencyID,
			Name:            agencyName,
			Email:           "contact@" + domain,
			DetectionMethod: "shared_email",
		}
		if err := db.UpsertAgency(a); err != nil {
			slog.Warn("failed to upsert agency from email domain", "domain", domain, "error", err)
			continue
		}

		for _, ch := range channels {
			ac := &core.AgencyChannel{
				AgencyID:   agencyID,
				ChannelID:  ch.ID,
				Confidence: 0.7,
				Evidence:   "shared_email_domain",
			}
			if err := db.UpsertAgencyChannel(ac); err != nil {
				continue
			}
			db.SetChannelAgency(ch.ID, agencyID)
		}

		slog.Info("found agency from email domain", "domain", domain, "channels", len(channels))
		agencies++
	}
	return agencies, nil
}

func scanDescriptionsForAgencies(db *core.DB) (int, error) {
	rows, err := db.Query(`SELECT id, name, description FROM channels
		WHERE description IS NOT NULL AND description != '' AND status = 'enriched'`)
	if err != nil {
		return 0, fmt.Errorf("query channels with descriptions: %w", err)
	}
	defer rows.Close()

	type match struct {
		channelID   string
		agencyName  string
	}

	var allMatches []match
	seen := make(map[string]bool)

	for rows.Next() {
		var id, name, desc string
		if err := rows.Scan(&id, &name, &desc); err != nil {
			continue
		}

		for _, re := range agencyDescPatterns {
			matches := re.FindStringSubmatch(desc)
			if len(matches) < 2 {
				continue
			}
			agencyName := strings.TrimSpace(matches[1])
			agencyName = strings.TrimRight(agencyName, ".")
			if agencyName == "" || seen[id] {
				continue
			}
			seen[id] = true
			allMatches = append(allMatches, match{channelID: id, agencyName: agencyName})
			break
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	agencyGroups := make(map[string][]string)
	for _, m := range allMatches {
		slug := slugify(m.agencyName)
		agencyGroups[slug] = append(agencyGroups[slug], m.channelID)
	}

	var agencies int
	for slug, channels := range agencyGroups {
		name := slug
		for _, m := range allMatches {
			if slugify(m.agencyName) == slug {
				name = m.agencyName
				break
			}
		}

		agencyID := "desc:" + slug
		a := &core.Agency{
			ID:              agencyID,
			Name:            name,
			DetectionMethod: "description",
		}
		if err := db.UpsertAgency(a); err != nil {
			slog.Warn("failed to upsert agency from description", "name", name, "error", err)
			continue
		}

		for _, chID := range channels {
			ac := &core.AgencyChannel{
				AgencyID:   agencyID,
				ChannelID:  chID,
				Confidence: 0.6,
				Evidence:   "description_regex",
			}
			if err := db.UpsertAgencyChannel(ac); err != nil {
				continue
			}
			db.SetChannelAgency(chID, agencyID)
		}

		slog.Info("found agency from description", "name", name, "channels", len(channels))
		agencies++
	}
	return agencies, nil
}

func extractDomain(rawURL string) string {
	rawURL = strings.TrimPrefix(rawURL, "https://")
	rawURL = strings.TrimPrefix(rawURL, "http://")
	rawURL = strings.TrimPrefix(rawURL, "www.")
	parts := strings.Split(rawURL, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return rawURL
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "&", "and")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	return s
}
