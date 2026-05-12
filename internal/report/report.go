package report

import (
	"fmt"
	"strings"

	"github.com/aquaflamingo/tubemedicmvp/internal/checker"
	"github.com/aquaflamingo/tubemedicmvp/internal/youtube"
)

const bar = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

// Print writes the full report to stdout.
func Print(ch *youtube.Channel, videos []youtube.Video, summary checker.Summary) {
	fmt.Printf("Tube Medic Report ─ %q\n", ch.Name)
	fmt.Println(bar)
	fmt.Printf("  Videos scanned    %d\n", len(videos))
	fmt.Printf("  Total links found %d\n", summary.Total)
	fmt.Printf("  Working           %d\n", summary.Live)
	fmt.Printf("  Broken            %d\n", summary.Broken)
	fmt.Println()

	if summary.Broken == 0 {
		fmt.Println("No broken links found.")
		return
	}

	fmt.Println(" Broken Links")
	fmt.Println(" " + strings.Repeat("─", 44))

	for i, r := range summary.BrokenLinks {
		label := r.Error
		if label == "" {
			label = fmt.Sprintf("%d", r.StatusCode)
		}
		fmt.Printf("  %d. %-9s %s\n", i+1, label, r.URL)
		fmt.Printf("     → %q\n", r.VideoTitle)
		fmt.Printf("     %s\n", youtube.VideoURL(r.VideoID))
		fmt.Println()
	}
}
