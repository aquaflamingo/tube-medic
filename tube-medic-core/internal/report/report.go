package report

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aqfl/tmcore/internal/checker"
	"github.com/aqfl/tmcore/internal/youtube"
)

const bar = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

var useColor = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"

func linkLabel(r checker.CheckResult) string {
	if r.Error != "" {
		return r.Error
	}
	return fmt.Sprintf("%d", r.StatusCode)
}

func printBrokenLinks(w io.Writer, header string, links []checker.CheckResult, useColor bool, colorCode string) {
	if len(links) == 0 {
		return
	}
	if useColor {
		fmt.Fprint(w, colorCode)
	}
	fmt.Fprintln(w, " "+header)
	fmt.Fprintln(w, " "+strings.Repeat("─", 44))
	if useColor {
		fmt.Fprint(w, colorReset)
	}
	for i, r := range links {
		fmt.Fprintf(w, "  %d. %-9s %s\n", i+1, linkLabel(r), r.URL)
		fmt.Fprintf(w, "     → %q\n", r.VideoTitle)
		fmt.Fprintf(w, "     %s\n", youtube.VideoURL(r.VideoID))
		fmt.Fprintln(w)
	}
}

// Print writes the full report to w.
func Print(w io.Writer, ch *youtube.Channel, videos []youtube.Video, summary checker.Summary, q youtube.Quota) {
	fmt.Fprintf(w, "Tube Medic Report ─ %q\n", ch.Name)
	fmt.Fprintln(w, bar)
	fmt.Fprintf(w, "  Videos scanned    %d\n", len(videos))
	fmt.Fprintf(w, "  Total links found %d\n", summary.Total)
	fmt.Fprintf(w, "  Working           %d\n", summary.Live)
	fmt.Fprintf(w, "  Broken            %d\n", summary.Broken)
	if summary.CriticalBroken > 0 {
		fmt.Fprintf(w, "  Revenue-critical  %d\n", summary.CriticalBroken)
	}
	if q.Used > 0 {
		remaining := "?"
		if q.Remaining >= 0 {
			remaining = fmt.Sprintf("%d", q.Remaining)
		}
		fmt.Fprintf(w, "  API quota used    %d / %s units\n", q.Used, remaining)
	}
	fmt.Fprintln(w)

	if summary.Broken == 0 && summary.Warnings == 0 {
		fmt.Fprintln(w, "No broken links found.")
		fmt.Fprintln(w)
	}

	if summary.Broken > 0 {
		printBrokenLinks(w, "Revenue-Critical Broken Links", summary.CriticalLinks, useColor, colorRed)
		printBrokenLinks(w, "Other Broken Links", summary.BrokenLinks, useColor, colorRed)
	}

	if summary.Warnings > 0 {
		fmt.Fprintf(w, "  %d warnings\n", summary.Warnings)
	}
}
