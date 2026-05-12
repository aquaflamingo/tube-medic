package checker

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aquaflamingo/tubemedicmvp/internal/classifier"
)

// Status represents the health of a checked URL.
type Status string

const (
	StatusWorking Status = "working"
	StatusWarning Status = "warning"
	StatusBroken  Status = "broken"
	StatusSkipped Status = "skipped"
)

// CheckResult holds the outcome of checking a single URL.
type CheckResult struct {
	URL        string
	VideoID    string
	VideoTitle string
	Status     Status
	StatusCode int
	Error      string
	Priority   classifier.Priority
}

// Summary groups results by status for reporting.
type Summary struct {
	Total     int
	Broken    int
	Warnings  int
	Live      int
	BrokenLinks    []CheckResult
	WarnLinks      []CheckResult
	CriticalBroken int
	CriticalLinks  []CheckResult
}

// Classify determines the status based on status code and error.
func Classify(code int, err error) (Status, string) {
	if err != nil {
		errStr := errorKind(err)
		return StatusBroken, errStr
	}
	if code >= 200 && code < 400 {
		return StatusWorking, ""
	}
	if code == 403 {
		return StatusWarning, fmt.Sprintf("%d", code)
	}
	return StatusBroken, fmt.Sprintf("%d", code)
}

func errorKind(err error) string {
	s := err.Error()
	if strings.Contains(s, "no such host") || strings.Contains(s, "DNS") {
		return "DNS_FAIL"
	}
	if strings.Contains(s, "timeout") || strings.Contains(s, "Timeout") {
		return "TIMEOUT"
	}
	if strings.Contains(s, "connection refused") {
		return "REFUSED"
	}
	if strings.Contains(s, "tls") || strings.Contains(s, "TLS") || strings.Contains(s, "certificate") {
		return "TLS_ERR"
	}
	return "NET_ERR"
}

// ShouldSkip returns true if the URL should not be checked.
func ShouldSkip(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return true
	}
	switch u.Scheme {
	case "http", "https":
		return false
	default:
		return true
	}
}

// IsDNSError checks if an error is specifically a DNS resolution failure.
func IsDNSError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "no such host")
}
