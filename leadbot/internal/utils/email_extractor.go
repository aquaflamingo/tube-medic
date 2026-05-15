package utils

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

var filterDomains = map[string]bool{
	"noreply": true,
	"no-reply": true,
	"donotreply": true,
	"info": true,
	"hello": true,
	"hi": true,
	"support": true,
	"contact": true,
	"admin": true,
	"mail": true,
	"team": true,
}

var filterTLDs = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true,
	".css": true, ".js": true, ".json": true, ".xml": true,
}

func ExtractEmails(text string) []string {
	if text == "" {
		return nil
	}
	matches := emailRegex.FindAllString(text, -1)
	var emails []string
	seen := make(map[string]bool)
	for _, e := range matches {
		e = strings.TrimSpace(e)
		e = strings.Trim(e, ".")
		e = strings.ToLower(e)
		if e == "" || seen[e] {
			continue
		}
		seen[e] = true

		local := strings.Split(e, "@")[0]
		if filterDomains[local] {
			continue
		}

		for tld := range filterTLDs {
			if strings.HasSuffix(e, tld) {
				goto skip
			}
		}
		emails = append(emails, e)
	skip:
	}
	return emails
}

func ExtractFromWebsite(url string) []string {
	paths := []string{"/contact", "/about", "/work-with-me", "/business", "/contact-us", "/hire-me", "/collab"}
	var allEmails []string
	seen := make(map[string]bool)

	for _, p := range paths {
		fullURL := url + p
		body, err := FetchURL(fullURL)
		if err != nil {
			continue
		}
		emails := ExtractEmails(body)
		for _, e := range emails {
			if !seen[e] {
				seen[e] = true
				allEmails = append(allEmails, e)
			}
		}
		if len(allEmails) > 0 {
			break
		}
	}

	return allEmails
}

func ExtractFromLinktree(url string) []string {
	doc, err := FetchDocument(url)
	if err != nil {
		return nil
	}

	var allEmails []string
	doc.Find("a[href^='mailto:']").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			email := strings.TrimPrefix(href, "mailto:")
			email = strings.TrimSpace(email)
			if email != "" {
				allEmails = append(allEmails, email)
			}
		}
	})

	body := doc.Text()
	emails := ExtractEmails(body)
	seen := make(map[string]bool)
	for _, e := range emails {
		if !seen[e] {
			seen[e] = true
			allEmails = append(allEmails, e)
		}
	}

	return allEmails
}
