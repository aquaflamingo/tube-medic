package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) > 5 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

func FetchDocument(rawURL string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fetch %s: status %d", rawURL, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}
	return doc, nil
}

func FetchURL(rawURL string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("fetch %s: status %d", rawURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func HasPricingPage(baseURL string) bool {
	paths := []string{"/pricing", "/shop", "/buy", "/store", "/products", "/merch", "/checkout"}
	for _, p := range paths {
		u, err := url.JoinPath(baseURL, p)
		if err != nil {
			continue
		}
		req, err := http.NewRequest("HEAD", u, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode < 400 {
			return true
		}
	}
	return false
}

func HasPricingPageGet(baseURL string) bool {
	paths := []string{"/pricing", "/shop", "/store", "/products", "/merch"}
	for _, p := range paths {
		u, err := url.JoinPath(baseURL, p)
		if err != nil {
			continue
		}
		body, err := FetchURL(u)
		if err != nil {
			continue
		}
		if len(body) > 1000 {
			return true
		}
	}
	return false
}

func ExtractLinksFromDocument(doc *goquery.Document) []string {
	var links []string
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && href != "" {
			links = append(links, href)
		}
	})
	return links
}

func IsLinktreeURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	return strings.Contains(lower, "linktr.ee") ||
		strings.Contains(lower, "beacons.ai") ||
		strings.Contains(lower, "koji.to") ||
		strings.Contains(lower, "bio.link") ||
		strings.Contains(lower, "bento.me")
}

func IsPatreonURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	return strings.Contains(lower, "patreon.com") ||
		strings.Contains(lower, "ko-fi.com") ||
		strings.Contains(lower, "buymeacoffee.com")
}

func IsCourseURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	return strings.Contains(lower, "teachable.com") ||
		strings.Contains(lower, "kajabi.com") ||
		strings.Contains(lower, "gumroad.com") ||
		strings.Contains(lower, "podia.com") ||
		strings.Contains(lower, "skool.com") ||
		strings.Contains(lower, "thinkific.com") ||
		strings.Contains(lower, "udemy.com")
}

func IsMerchURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	return strings.Contains(lower, "shopify.com") ||
		strings.Contains(lower, "teespring.com") ||
		strings.Contains(lower, "spring.com") ||
		strings.Contains(lower, "merch")
}

func IsAmazonURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	return strings.Contains(lower, "amazon.com") ||
		strings.Contains(lower, "amazon.co.uk") ||
		strings.Contains(lower, "amzn.to")
}

func ExtractYouTubeChannelIdent(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	if !strings.Contains(u.Host, "youtube.com") && !strings.Contains(u.Host, "youtu.be") {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, part := range parts {
		if part == "channel" && i+1 < len(parts) {
			return parts[i+1]
		}
		if part == "c" && i+1 < len(parts) {
			return "@" + parts[i+1]
		}
		if part == "user" && i+1 < len(parts) {
			return "@" + parts[i+1]
		}
		if strings.HasPrefix(part, "@") {
			return part
		}
	}
	return ""
}

func ExtractYouTubeIDsFromPage(pageURL string) ([]string, error) {
	doc, err := FetchDocument(pageURL)
	if err != nil {
		return nil, err
	}
	var ids []string
	seen := make(map[string]bool)
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}
		ident := ExtractYouTubeChannelIdent(href)
		if ident != "" && !seen[ident] {
			seen[ident] = true
			ids = append(ids, ident)
		}
	})
	return ids, nil
}
