package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aquaflamingo/tmcore"
	"github.com/aquaflamingo/tmcore/internal/config"
	"github.com/aquaflamingo/tmcore/internal/report"
)

func main() {
	cfg, outputFile, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	r, err := tmcore.RunScan(*cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	out := io.Writer(os.Stdout)
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = io.MultiWriter(os.Stdout, f)
	}
	report.Print(out, r.Channel, r.Videos, r.Summary, r.Quota)

	if r.Summary.Broken > 0 {
		os.Exit(1)
	}
}
