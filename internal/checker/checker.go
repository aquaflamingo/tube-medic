package checker

import (
	"net/http"
	"sync"
	"time"

	"github.com/aquaflamingo/tubemedicmvp/internal/youtube"
)

var httpClient = &http.Client{
	Timeout:   5 * time.Second,
	Transport: &userAgentTransport{},
}

type userAgentTransport struct{}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	return http.DefaultTransport.RoundTrip(req)
}

// CheckAll checks all scraped links concurrently using a worker pool.
// concurrency sets the max number of simultaneous checks.
func CheckAll(links []youtube.ScrapedLink, concurrency int) []CheckResult {
	if concurrency < 1 {
		concurrency = 1
	}

	sem := make(chan struct{}, concurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup

	results := make([]CheckResult, 0, len(links))

	for _, link := range links {
		if ShouldSkip(link.URL) {
			results = append(results, CheckResult{
				URL:        link.URL,
				VideoID:    link.VideoID,
				VideoTitle: link.VideoTitle,
				Status:     StatusSkipped,
				Error:      "skipped (non-http scheme)",
			})
			continue
		}

		wg.Add(1)
		go func(l youtube.ScrapedLink) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r := checkOne(l)
			mu.Lock()
			results = append(results, r)
			mu.Unlock()
		}(link)
	}

	wg.Wait()
	close(sem)

	return results
}

func checkOne(link youtube.ScrapedLink) CheckResult {
	resp, err := httpClient.Get(link.URL)
	if err != nil {
		status, errStr := Classify(0, err)
		return CheckResult{
			URL:        link.URL,
			VideoID:    link.VideoID,
			VideoTitle: link.VideoTitle,
			Status:     status,
			Error:      errStr,
		}
	}
	defer resp.Body.Close()

	status, errStr := Classify(resp.StatusCode, nil)
	return CheckResult{
		URL:        link.URL,
		VideoID:    link.VideoID,
		VideoTitle: link.VideoTitle,
		Status:     status,
		StatusCode: resp.StatusCode,
		Error:      errStr,
	}
}

// Summarize groups results into a Summary struct for reporting.
func Summarize(results []CheckResult) Summary {
	s := Summary{}
	for _, r := range results {
		s.Total++
		switch r.Status {
		case StatusBroken:
			s.Broken++
			s.BrokenLinks = append(s.BrokenLinks, r)
		case StatusWarning:
			s.Warnings++
			s.WarnLinks = append(s.WarnLinks, r)
		default:
			s.Live++
		}
	}
	return s
}
