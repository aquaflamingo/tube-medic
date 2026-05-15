package config

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	YTApiKey          string
	MaxEnrichments    int
	MinSubs           int64
	MaxSubs           int64
	MinScore          int
	DataDir           string
	LogDir            string
	ExportDir         string
	DBPath            string
	Keywords          []string
	NicheCategories   map[string]string
	AgencySeeds       []AgencySeed
	UserAgent         string
	RequestDelay      int
}

type AgencySeed struct {
	Name    string `json:"name"`
	Website string `json:"website"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		YTApiKey:          os.Getenv("YT_API_KEY"),
		MaxEnrichments:    getEnvInt("MAX_ENRICHMENTS", 800),
		MinSubs:           getEnvInt64("MIN_SUBS", 40000),
		MaxSubs:           getEnvInt64("MAX_SUBS", 500000),
		MinScore:          getEnvInt("MIN_SCORE", 3),
		DataDir:           getEnvStr("DATA_DIR", "data"),
		LogDir:            getEnvStr("LOG_DIR", "logs"),
		ExportDir:         getEnvStr("EXPORT_DIR", "exports"),
		DBPath:            getEnvStr("DB_PATH", "data/ytleadbot.db"),
		UserAgent:         getEnvStr("USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
		RequestDelay:      getEnvInt("REQUEST_DELAY", 5),
		NicheCategories:   defaultNicheCategories(),
	}

	keywordsPath := getEnvStr("KEYWORDS_PATH", "config/keywords_seed.json")
	keywords, categories, err := loadKeywordSeed(keywordsPath)
	if err != nil {
		cfg.Keywords = defaultKeywords()
	} else {
		cfg.Keywords = keywords
		if len(categories) > 0 {
			cfg.NicheCategories = categories
		}
	}

	agencies, err := loadAgencies(getEnvStr("AGENCIES_PATH", "config/agencies_seed.json"))
	if err != nil {
		cfg.AgencySeeds = defaultAgencies()
	} else {
		cfg.AgencySeeds = agencies
	}

	return cfg, nil
}

func getEnvStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvInt64(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

type KeywordSeedFile struct {
	Keywords        []string          `json:"keywords"`
	NicheCategories map[string]string `json:"niche_categories"`
}

func loadKeywordSeed(path string) ([]string, map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	// Try new format (object with keywords array + niche_categories)
	var seed KeywordSeedFile
	if err := json.Unmarshal(data, &seed); err == nil && len(seed.Keywords) > 0 {
		return seed.Keywords, seed.NicheCategories, nil
	}

	// Fall back to old format (flat string array)
	var keywords []string
	if err := json.Unmarshal(data, &keywords); err != nil {
		return nil, nil, err
	}
	return keywords, nil, nil
}

func loadAgencies(path string) ([]AgencySeed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var agencies []AgencySeed
	if err := json.Unmarshal(data, &agencies); err != nil {
		return nil, err
	}
	return agencies, nil
}

func defaultKeywords() []string {
	return []string{
		"fitness course",
		"fitness coaching",
		"personal finance how I make money",
		"personal finance my business",
		"tech sponsor",
		"tech collab",
		"business how I make money",
		"business coaching",
		"cooking course",
		"cooking my business",
		"travel sponsor",
		"travel collab",
		"gaming sponsor",
		"self improvement course",
		"self improvement coaching",
		"parenting how I make money",
		"beauty sponsor",
	}
}

func defaultNicheCategories() map[string]string {
	return map[string]string{
		"fitness":          "26",
		"personal_finance": "27",
		"tech":             "28",
		"business":         "27",
		"cooking":          "26",
		"travel":           "19",
		"gaming":           "20",
		"self_improvement": "27",
		"parenting":        "22",
		"beauty":           "22",
	}
}

func defaultAgencies() []AgencySeed {
	return []AgencySeed{
		{Name: "Whalar", Website: "https://whalar.com"},
		{Name: "Viral Nation", Website: "https://viralnation.com"},
		{Name: "Gleam Futures", Website: "https://gleamfutures.com"},
		{Name: "Studio71", Website: "https://studio71.com"},
		{Name: "Fullscreen", Website: "https://fullscreen.com"},
		{Name: "Night Media", Website: "https://nightmedia.co"},
		{Name: "Select Management", Website: "https://select.co"},
		{Name: "BBTV", Website: "https://bbtv.com"},
		{Name: "United Talent Agency", Website: "https://unitedtalent.com"},
		{Name: "Creative Artists Agency", Website: "https://caa.com"},
		{Name: "Endeavor", Website: "https://endeavor.com"},
		{Name: "Innovative Artists", Website: "https://innovativeartists.com"},
		{Name: "Paradigm Talent Agency", Website: "https://paradigmagency.com"},
		{Name: "Digital Brand Architects", Website: "https://digitalbrandarchitects.com"},
		{Name: "Socialyte", Website: "https://socialyte.co"},
		{Name: "TheSquad", Website: "https://thesquad.com"},
		{Name: "Loud Community", Website: "https://loudcommunity.com"},
		{Name: "Sway House", Website: "https://swayhouse.com"},
		{Name: "Clout Studios", Website: "https://cloutstudios.com"},
		{Name: "Valuetainment", Website: "https://valuetainment.com"},
	}
}
