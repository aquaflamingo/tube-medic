package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aqfl/tubemedic-web/internal/handlers"
)

func main() {
	// Load environment variables from .env if present
	loadEnv()

	// Verify YouTube API key is configured
	if os.Getenv("YT_API_KEY") == "" {
		log.Fatal("FATAL: YT_API_KEY environment variable is not set. Please add it to your .env file.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Static files (if needed later)
	// mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", handlers.HandleIndex)
	mux.HandleFunc("/scan", handlers.HandleScan)

	fmt.Printf("Tube Medic Web starting on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

// loadEnv looks for a .env file in standard locations and loads its variables.
func loadEnv() {
	paths := []string{
		".env",
		"../.env",
		"../tube-medic-core/.env",
	}

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
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
		log.Printf("Loaded environment variables from %s", path)
		return
	}
	log.Println("Note: No .env file loaded from default paths; relying on system environment.")
}
