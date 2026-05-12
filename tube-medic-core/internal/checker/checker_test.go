package checker_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aquaflamingo/tubemedicmvp/internal/checker"
	"github.com/aquaflamingo/tubemedicmvp/internal/youtube"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name       string
		code       int
		err        error
		wantStatus string
		wantErr    string
	}{
		{name: "200 ok", code: 200, err: nil, wantStatus: "working", wantErr: ""},
		{name: "301 redirect", code: 301, err: nil, wantStatus: "working", wantErr: ""},
		{name: "403 forbidden", code: 403, err: nil, wantStatus: "warning", wantErr: "403"},
		{name: "404 not found", code: 404, err: nil, wantStatus: "broken", wantErr: "404"},
		{name: "410 gone", code: 410, err: nil, wantStatus: "broken", wantErr: "410"},
		{name: "500 server error", code: 500, err: nil, wantStatus: "broken", wantErr: "500"},
		{name: "dns failure", code: 0, err: errors.New("no such host"), wantStatus: "broken", wantErr: "DNS_FAIL"},
		{name: "timeout", code: 0, err: errors.New("timeout after 5s"), wantStatus: "broken", wantErr: "TIMEOUT"},
		{name: "connection refused", code: 0, err: errors.New("connection refused"), wantStatus: "broken", wantErr: "REFUSED"},
		{name: "tls error", code: 0, err: errors.New("tls handshake failed"), wantStatus: "broken", wantErr: "TLS_ERR"},
		{name: "generic network error", code: 0, err: errors.New("dial tcp: no route to host"), wantStatus: "broken", wantErr: "NET_ERR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, errStr := checker.Classify(tt.code, tt.err)
			if string(status) != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if errStr != tt.wantErr {
				t.Errorf("errStr = %q, want %q", errStr, tt.wantErr)
			}
		})
	}
}

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{url: "https://example.com", want: false},
		{url: "http://example.com", want: false},
		{url: "mailto:test@example.com", want: true},
		{url: "ftp://files.example.com", want: true},
		{url: "javascript:void(0)", want: true},
		{url: "", want: true},
		{url: "not-a-url", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := checker.ShouldSkip(tt.url)
			if got != tt.want {
				t.Errorf("ShouldSkip(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestCheckAll_integration(t *testing.T) {
	// Start a local server that returns specific status codes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/forbidden":
			w.WriteHeader(http.StatusForbidden)
		case "/redirect":
			http.Redirect(w, r, "/ok", http.StatusMovedPermanently)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts2.Close()

	links := []youtube.ScrapedLink{
		{URL: ts.URL + "/ok", VideoID: "v1", VideoTitle: "OK"},
		{URL: ts.URL + "/notfound", VideoID: "v2", VideoTitle: "Not Found"},
		{URL: ts.URL + "/redirect", VideoID: "v3", VideoTitle: "Redirect"},
		{URL: ts2.URL + "/err", VideoID: "v4", VideoTitle: "Server Error"},
		{URL: "mailto:test@example.com", VideoID: "v5", VideoTitle: "Skipped"},
		{URL: ts.URL + "/forbidden", VideoID: "v6", VideoTitle: "Forbidden"},
	}

	results := checker.CheckAll(links, 4)

	got := make(map[string]string)
	for _, r := range results {
		got[r.VideoID] = string(r.Status)
	}

	if got["v1"] != "working" {
		t.Errorf("v1 status = %q, want %q", got["v1"], "working")
	}
	if got["v2"] != "broken" {
		t.Errorf("v2 status = %q, want %q", got["v2"], "broken")
	}
	if got["v3"] != "working" {
		t.Errorf("v3 status = %q, want %q", got["v3"], "working")
	}
	if got["v4"] != "broken" {
		t.Errorf("v4 status = %q, want %q", got["v4"], "broken")
	}
	if got["v5"] != "skipped" {
		t.Errorf("v5 status = %q, want %q", got["v5"], "skipped")
	}
	if got["v6"] != "warning" {
		t.Errorf("v6 status = %q, want %q", got["v6"], "warning")
	}
}

func TestSummarize(t *testing.T) {
	results := []checker.CheckResult{
		{URL: "https://a.com", Status: checker.StatusWorking},
		{URL: "https://b.com", Status: checker.StatusBroken, Error: "404"},
		{URL: "https://c.com", Status: checker.StatusWarning, Error: "403"},
		{URL: "https://d.com", Status: checker.StatusWorking},
		{URL: "mailto:e@com", Status: checker.StatusSkipped},
	}

	s := checker.Summarize(results)
	if s.Total != 5 {
		t.Errorf("Total = %d, want 5", s.Total)
	}
	if s.Broken != 1 {
		t.Errorf("Broken = %d, want 1", s.Broken)
	}
	if s.Warnings != 1 {
		t.Errorf("Warnings = %d, want 1", s.Warnings)
	}
	if s.Live != 3 {
		t.Errorf("Live = %d, want 3", s.Live)
	}
	if len(s.BrokenLinks) != 1 {
		t.Errorf("len(BrokenLinks) = %d, want 1", len(s.BrokenLinks))
	}
	if len(s.WarnLinks) != 1 {
		t.Errorf("len(WarnLinks) = %d, want 1", len(s.WarnLinks))
	}
}
