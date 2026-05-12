package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	APIKey     string
	ChannelURL string
	MaxVideos  int
	OutputFile string
}

func Load() (*Config, error) {
	loadEnvFile()

	apiKey := flag.String("api-key", "", "YouTube Data API key (or set YT_API_KEY in .env)")
	channel := flag.String("channel", "", "Full YouTube channel URL (e.g. https://youtube.com/@mkbhd)")
	maxVideos := flag.Int("max-videos", 50, "Number of recent videos to scan")
	output := flag.String("output", "", "Save report to file")
	flag.Parse()

	if *apiKey == "" {
		*apiKey = os.Getenv("YT_API_KEY")
	}

	if *apiKey == "" {
		return nil, fmt.Errorf("API key required: set YT_API_KEY in .env or pass --api-key")
	}

	if *channel == "" {
		return nil, fmt.Errorf("--channel is required")
	}

	return &Config{
		APIKey:     *apiKey,
		ChannelURL: *channel,
		MaxVideos:  *maxVideos,
		OutputFile: *output,
	}, nil
}

func loadEnvFile() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
